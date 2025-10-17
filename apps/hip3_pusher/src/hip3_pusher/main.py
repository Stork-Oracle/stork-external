import os
import queue
import threading
import time
import json
import asyncio
import random
import yaml
from yaml.loader import SafeLoader
from websockets import connect

from .logging import setup_logging, get_logger

MSG_QUEUE_MAX = 10_000
FLUSH_INTERVAL_SEC = 3.0

q = queue.Queue(maxsize=MSG_QUEUE_MAX)
stop_event = threading.Event()

state = {}
state_lock = threading.Lock()

def send_to_endpoint(snapshot: dict):
    """Send snapshot data to the configured endpoint."""
    logger = get_logger(__name__)
    
    # Replace with real HTTP call
    # requests.post("https://example.com/ingest", json=snapshot, timeout=5)
    logger.info(f"Sending data to endpoint: {len(snapshot)} items - {list(snapshot.keys())}")

async def connect_with_basic_auth(url: str, auth: str, assets: list[str]):
    """Connect to WebSocket with basic authentication and subscribe to assets."""
    logger = get_logger(__name__)
    
    logger.info(f"Establishing WebSocket connection to {url} for {len(assets)} assets")
    
    async with connect(url, additional_headers=[("Authorization", f"Basic {auth}")]) as ws:
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
            
            if message_count % 100 == 0:  # Log every 100 messages
                logger.debug(f"WebSocket messages received: {message_count}")
            

def run_websocket(url: str, auth: str, assets: list[str]):
    """Run WebSocket connection with retry logic."""
    logger = get_logger(__name__)
    
    max_retries = 100
    base_delay = 1.0  # Start with 1 second
    max_delay = 5.0  # Cap at 5 seconds
    retry_count = 0
    
    logger.info(f"Starting WebSocket connection manager for {url} (max retries: {max_retries})")
    
    while not stop_event.is_set() and retry_count < max_retries:
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        
        try:
            logger.info(f"Attempting WebSocket connection (attempt {retry_count + 1}/{max_retries})")
            
            start_time = time.monotonic()
            loop.run_until_complete(connect_with_basic_auth(url, auth, assets))
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

def coordinator():
    """Coordinate message processing and data flushing."""
    logger = get_logger(__name__)
    
    next_flush = time.monotonic() + FLUSH_INTERVAL_SEC
    logger.info(f"Coordinator started (flush interval: {FLUSH_INTERVAL_SEC}s)")
    
    messages_processed = 0
    flushes_completed = 0

    while not stop_event.is_set():
        remaining = max(0.0, next_flush - time.monotonic())

        if remaining <= 0:
            # Time to flush
            with state_lock:
                snapshot = dict(state)  # shallow copy
                send_to_endpoint(snapshot)
                next_flush += FLUSH_INTERVAL_SEC
                flushes_completed += 1
                logger.debug(f"Data flush completed ({flushes_completed} total, {len(snapshot)} items)")
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
            if "data" in msg_data and "BTCUSD" in msg_data["data"]:
                state["BTCUSD"] = msg_data["data"]["BTCUSD"]
                state["count"] = state.get("count", 0) + 1
                messages_processed += 1
                
                if messages_processed % 50 == 0:  # Log every 50 processed messages
                    logger.debug(f"Messages processed: {messages_processed}")

        q.task_done()

    # Optional final flush on shutdown
    logger.info(f"Coordinator shutting down (processed {messages_processed} messages, {flushes_completed} flushes)")
    
    with state_lock:
        snapshot = dict(state)
    send_to_endpoint(snapshot)
    
    logger.info("Coordinator shutdown complete")

def main():
    """Main entry point for the hip3_pusher service."""
    # Configure logging for service mode
    setup_logging(level="INFO", json_format=True)
    logger = get_logger(__name__)
    
    logger.info("HIP3 pusher service starting")
    
    config_path = os.getenv("CONFIG_PATH")
    if not config_path:
        logger.error("CONFIG_PATH environment variable not set")
        return
    
    try:
        with open(config_path, "r") as f:
            config = yaml.load(f, Loader=SafeLoader)
        logger.info(f"Configuration loaded successfully: {config_path}")
    except Exception as e:
        logger.error(f"Failed to load configuration {config_path}: {e}")
        return
    # url = os.getenv("STORK_WS_URL")
    # auth = os.getenv("STORK_WS_AUTH")
    # assets = os.getenv("STORK_WS_ASSETS").split(",")
    # ws_t = threading.Thread(target=run_websocket, args=(url, auth, assets), name="ws", daemon=True)
    # coord_t = threading.Thread(target=coordinator, name="coord", daemon=True)

    # ws_t.start()
    # coord_t.start()

    # try:
    #     # Keep main thread alive but responsive to signals
    #     while not stop_event.is_set():
    #         stop_event.wait(timeout=1.0)
    # except KeyboardInterrupt:
    #     print("KeyboardInterrupt received, shutting down...")
    #     stop_event.set()
    
    # # Give threads a moment to clean up
    # print("Waiting for threads to finish...")
    # ws_t.join(timeout=5.0)
    # coord_t.join(timeout=5.0)
    
    # print("Shutdown complete")
