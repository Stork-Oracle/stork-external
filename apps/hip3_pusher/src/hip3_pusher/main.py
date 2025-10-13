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

MSG_QUEUE_MAX = 10_000
FLUSH_INTERVAL_SEC = 3.0

q = queue.Queue(maxsize=MSG_QUEUE_MAX)
stop_event = threading.Event()

state = {}
state_lock = threading.Lock()

def send_to_endpoint(snapshot: dict):
    # Replace with real HTTP call
    # requests.post("https://example.com/ingest", json=snapshot, timeout=5)
    print("POST ->", json.dumps(snapshot))

async def connect_with_basic_auth(url: str, auth: str, assets: list[str]):
    async with connect(url, additional_headers=[("Authorization", f"Basic {auth}")]) as ws:
        # Send JSON as string
        await ws.send(json.dumps({"type": "subscribe", "data": assets}))
        async for msg in ws:
            if stop_event.is_set():
                break
            q.put(msg)
            

def run_websocket(url: str, auth: str, assets: list[str]):
    """Wrapper to run async websocket in a thread with retry logic"""
    max_retries = 100
    base_delay = 1.0  # Start with 1 second
    max_delay = 5.0  # Cap at 5 seconds
    retry_count = 0
    
    while not stop_event.is_set() and retry_count < max_retries:
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        
        try:
            print(f"Attempting WebSocket connection (attempt {retry_count + 1}/{max_retries})")
            loop.run_until_complete(connect_with_basic_auth(url, auth, assets))
            # If we get here, connection was successful and closed normally
            print("WebSocket connection closed normally")
            break
            
        except Exception as e:
            retry_count += 1
            print(f"WebSocket error (attempt {retry_count}/{max_retries}): {e}")
            
            if retry_count >= max_retries:
                print("Max retries reached. WebSocket thread exiting.")
                break
                
            # Calculate delay with exponential backoff and jitter
            delay = min(base_delay * (2 ** (retry_count - 1)), max_delay)
            jitter = random.uniform(0.1, 0.3) * delay  # Add 10-30% jitter
            total_delay = delay + jitter
            
            print(f"Retrying in {total_delay:.1f} seconds...")
            
            # Wait for retry delay or stop event
            if stop_event.wait(timeout=total_delay):
                print("Stop event received during retry delay")
                break
                
        finally:
            loop.close()
    
    if retry_count >= max_retries:
        raise Exception("WebSocket connection failed permanently after max retries")
    elif stop_event.is_set():
        print("WebSocket thread stopped due to stop event")

def coordinator():
    """
    Emulates Go’s: select { case <-msgCh; case <-ticker.C }
    by using Queue.get(timeout=remaining_time).
    """
    next_flush = time.monotonic() + FLUSH_INTERVAL_SEC
    print("Coordinator started")

    while not stop_event.is_set():
        remaining = max(0.0, next_flush - time.monotonic())

        if remaining <= 0:
            with state_lock:
                snapshot = dict(state)  # shallow copy
                send_to_endpoint(snapshot)
                next_flush += FLUSH_INTERVAL_SEC
                continue

        try:
            msg = q.get(timeout=remaining)  # blocks until msg or timeout
        except queue.Empty:
            continue

        msg = json.loads(msg)

        if msg["type"] != "oracle_prices":
            continue

        with state_lock:
            state["BTCUSD"] = msg["data"]["BTCUSD"]
            state["count"] = state.get("count", 0) + 1

        q.task_done()

    # Optional final flush on shutdown
    with state_lock:
        snapshot = dict(state)
    send_to_endpoint(snapshot)

def main():
    config_path = os.getenv("CONFIG_PATH")
    with open(config_path, "r") as f:
        config = yaml.load(f, Loader=SafeLoader)
    print(config)
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
