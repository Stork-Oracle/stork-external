import queue
import threading
import time
import json
import asyncio
import random
from hip3_pusher.config import DexConfig
from websockets import connect
import logging

from hyperliquid.info import Info
from hyperliquid.exchange import Exchange
from hyperliquid.utils import constants
import eth_account
from eth_account.signers.local import LocalAccount

logger = logging.getLogger("hip3_pusher")

MSG_QUEUE_MAX = 10_000
FLUSH_INTERVAL_SEC = 3.0

q = queue.Queue(maxsize=MSG_QUEUE_MAX)
stop_event = threading.Event()

state = {}
state_lock = threading.Lock()

def send_to_endpoint(snapshot: dict, exchange: Exchange, dex: str):
    """Send snapshot data to the configured endpoint."""
    logger.info(f"Sending data to endpoint: {len(snapshot)} items - {snapshot}")
    oracle_pxs = {}
    market_pxs = []
    external_pxs = {}
    for asset in snapshot:
        oracle_pxs[f"{dex}:{asset}"] = snapshot[asset]
        market_pxs.append({ f"{dex}:{asset}": snapshot[asset] })
        external_pxs[f"{dex}:{asset}"] = snapshot[asset]

    set_oracle_result = exchange.perp_deploy_set_oracle(
        dex,
        oracle_pxs,
        market_pxs,
        external_pxs,
    )
    logger.info(f"Set oracle result: {set_oracle_result}")


async def connect_with_basic_auth(endpoint: str, auth: str, assets: list[str]):
    """Connect to WebSocket with basic authentication and subscribe to assets."""

    logger.info(f"Establishing WebSocket connection to {endpoint} for {len(assets)} assets")
    
    async with connect(endpoint + "/evm/subscribe", additional_headers=[("Authorization", f"Basic {auth}")]) as ws:
        # Send JSON as string
        subscription_msg = {"type": "subscribe", "data": assets}
        await ws.send(json.dumps(subscription_msg))
        logger.info(f"Subscribed to assets: {assets}")
        
        message_count = 0
        async for msg in ws:
            if stop_event.is_set():
                logger.info("Stop event received, closing WebSocket connection")
                break
            
            message_count += 1
            q.put(msg)
            

def run_websocket(endpoint: str, auth: str, assets: list[str]):
    """Run WebSocket connection with retry logic."""
    max_retries = 100
    base_delay = 1.0  # Start with 1 second
    max_delay = 5.0  # Cap at 5 seconds
    retry_count = 0
    
    logger.info(f"Starting WebSocket connection manager for {endpoint} (max retries: {max_retries})")
    
    while not stop_event.is_set() and retry_count < max_retries:
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        
        try:
            logger.info(f"Attempting WebSocket connection (attempt {retry_count + 1}/{max_retries})")
            
            start_time = time.monotonic()
            loop.run_until_complete(connect_with_basic_auth(endpoint, auth, assets))
            duration = (time.monotonic() - start_time) * 1000
            
            # If we get here, connection was successful and closed normally
            logger.info(f"WebSocket connection closed normally (duration: {duration:.2f}ms)")
            break
            
        except Exception as e:
            retry_count += 1
            logger.warning(f"WebSocket connection failed (attempt {retry_count}/{max_retries}): {e}")
            
            if retry_count >= max_retries:
                logger.error("Max retries reached, WebSocket thread exiting")
                break
                
            # Calculate delay with exponential backoff and jitter
            delay = min(base_delay * (2 ** (retry_count - 1)), max_delay)
            jitter = random.uniform(0.1, 0.3) * delay  # Add 10-30% jitter
            total_delay = delay + jitter
            
            logger.info(f"Retrying in {total_delay:.1f} seconds (attempt {retry_count + 1})")
            
            # Wait for retry delay or stop event
            if stop_event.wait(timeout=total_delay):
                logger.info("Stop event received during retry delay")
                break
                
        finally:
            loop.close()
    
    if retry_count >= max_retries:
        error_msg = f"WebSocket connection failed permanently after {retry_count} retries"
        logger.error(error_msg)
        raise Exception(error_msg)
    elif stop_event.is_set():
        logger.info("WebSocket thread stopped due to stop event")

def coordinator(private_key: str, dex_config: DexConfig):
    """Coordinate message processing and data flushing."""

    account: LocalAccount = eth_account.Account.from_key(private_key)
    exchange = Exchange(
        account, 
        constants.TESTNET_API_URL if dex_config.testnet else constants.MAINNET_API_URL, 
        account_address=account.address, perp_dexs=None
    )

    next_flush = time.monotonic() + FLUSH_INTERVAL_SEC
    logger.info(f"Coordinator started (flush interval: {FLUSH_INTERVAL_SEC}s)")
    
    while not stop_event.is_set():
        remaining = max(0.0, next_flush - time.monotonic())

        if remaining <= 0:
            # Time to flush
            with state_lock:
                snapshot = dict(state)  # shallow copy
                send_to_endpoint(snapshot, exchange, dex_config.name)
                next_flush += FLUSH_INTERVAL_SEC
            continue

        try:
            msg = q.get(timeout=remaining)  # blocks until msg or timeout
        except queue.Empty:
            continue

        try:
            msg_data = json.loads(msg)
        except json.JSONDecodeError as e:
            logger.warning(f"Failed to parse message JSON: {e}")
            continue

        if msg_data["type"] != "oracle_prices":
            logger.debug(f"Ignoring non-oracle message: {msg_data.get('type')}")
            continue

        with state_lock:
            if "data" in msg_data:
                for asset in msg_data["data"]:
                    # price is scaled be 1e18. Add a period to the left of the least significant digit by 18 places using string manipulation
                    price = msg_data["data"][asset]["stork_signed_price"]["price"]
                    price = price.zfill(18)
                    price = price[:-18] + "." + price[-18:]
                    state[asset] = f"{float(price)}"

        q.task_done()

    # Optional final flush on shutdown
    logger.info(f"Coordinator shutting down")
    
    with state_lock:
        snapshot = dict(state)
    send_to_endpoint(snapshot, exchange, dex_config.name)
    
    logger.info("Coordinator shutdown complete")

def run(stork_ws_endpoint: str, stork_ws_auth: str, stork_ws_assets: list[str], private_key: str, dex_config: DexConfig):
    """Run the hip3_pusher service."""
    # Configure basic logging for service mode

    ws_t = threading.Thread(target=run_websocket, args=(stork_ws_endpoint, stork_ws_auth, stork_ws_assets), name="ws", daemon=True)
    coord_t = threading.Thread(target=coordinator, name="coord", args=(private_key, dex_config), daemon=True)

    ws_t.start()
    coord_t.start()

    try:
        # Keep main thread alive but responsive to signals
        while not stop_event.is_set():
            stop_event.wait(timeout=1.0)
    except KeyboardInterrupt:
        logger.info("KeyboardInterrupt received, shutting down...")
        stop_event.set()
    
    # Give threads a moment to clean up
    logger.info("Waiting for threads to finish...")
    ws_t.join(timeout=5.0)
    coord_t.join(timeout=5.0)
    
    logger.info("Shutdown complete")
