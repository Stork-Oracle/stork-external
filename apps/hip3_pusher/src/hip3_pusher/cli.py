from __future__ import annotations
import sys
from pathlib import Path
import typer
from typing import Optional
import logging
import json
import datetime
from .runner import run

from .config import validate_config_early

class JSONFormatter(logging.Formatter):
    """Simple JSON formatter for log records."""
    
    def format(self, record):
        log_entry = {
            "timestamp": datetime.datetime.fromtimestamp(record.created).isoformat(),
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
        }
        
        if record.exc_info:
            log_entry["exception"] = self.formatException(record.exc_info)
            
        return json.dumps(log_entry)


app = typer.Typer(help="HIP3 pusher CLI: push HIP-3 configs/events somewhere.", add_completion=False)

@app.command()
def push(
    config: Path = typer.Argument(..., help="Main config file", metavar="CONFIG"),
    private_key_file: str = typer.Option(
        ..., "--private-key-file", "-k", help="Private key file"
    ),
    stork_ws_endpoint: str = typer.Option(
        "wss://api.jp.stork-oracle.network", "--stork-ws-url", help="Stork WebSocket URL"
    ),
    stork_ws_auth: str = typer.Option(
        ..., "--stork-ws-auth", "-a", help="Stork WebSocket authentication"
    ),
    verbose: int = typer.Option(
        0, "--verbose", "-v", count=True, help="Increase verbosity (-v, -vv)"
    ),
    json_logs: bool = typer.Option(
        False, "--json-logs", help="Output logs in JSON format"
    ),
):
    """
    Push using HIP-3 configs.

    Example:
      hip3-pusher push ./examples/sample_config.yaml -c ./examples/creds.toml -vv
    """
    # Configure logging based on verbosity and format preference
    log_level = logging.ERROR
    if verbose >= 1:
        log_level = logging.INFO
    if verbose >= 2:
        log_level = logging.DEBUG

    # Configure basic logging
    if json_logs:
        # Use JSON formatter for structured logging
        handler = logging.StreamHandler()
        handler.setFormatter(JSONFormatter())
        logging.basicConfig(
            level=log_level,
            handlers=[handler]
        )
    else:
        # Use standard text format
        logging.basicConfig(
            level=log_level,
            format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
            datefmt="%Y-%m-%d %H:%M:%S"
        )
    
    logger = logging.getLogger("hip3_pusher")
    logger.info("Starting HIP3 push operation")
    
    # Validate config structure early - fail fast if invalid
    try:
        validated_config = validate_config_early(config)
        logger.info(f"Configuration validated successfully: {config}")
    except Exception as e:
        logger.error(f"Configuration validation failed: {config} - {e}")
        raise typer.Exit(1)

    private_key = open(private_key_file, "r").read().strip()

    # Log configuration details
    network_type = "testnet" if validated_config.config.dex.testnet else "mainnet"
    logger.info(f"DEX: {validated_config.config.dex.name} ({network_type})")
    logger.info(f"Markets: {len(validated_config.markets)} configured")
    stork_ws_assets = [market.stork_spot_asset for market in validated_config.markets]
    
    # Log detailed market information in debug mode
    for i, market in enumerate(validated_config.markets, 1):
        logger.debug(f"Market {i}: {market.hip3_name} -> {market.stork_spot_asset} (autocalc: {market.autocalculate_ext})")

    run(stork_ws_endpoint, stork_ws_auth, stork_ws_assets, private_key, validated_config.config.dex)

@app.callback()
def main(
    version: Optional[bool] = typer.Option(
        None, "--version", callback=lambda v: (print("hip3_pusher 0.1.0") or sys.exit(0)) if v else None,
        is_eager=True, help="Show version and exit"
    )
):
    # Runs before subcommands; useful for global flags in bigger CLIs.
    pass
