from __future__ import annotations
import sys
from pathlib import Path
import typer
from typing import Optional
import logging
import json
import datetime

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
    creds: Path = typer.Option(
        None,
        "--creds", "-c",
        help="Credentials file (TOML/YAML/JSON)",
    ),
    dry_run: bool = typer.Option(
        False, "--dry-run", help="Validate and print actions without executing"
    ),
    endpoint: Optional[str] = typer.Option(
        None, "--endpoint", envvar="HIP3_ENDPOINT", help="Target endpoint (or HIP3_ENDPOINT)"
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

    # Log configuration details
    network_type = "testnet" if validated_config.config.dex.testnet else "mainnet"
    logger.info(f"DEX: {validated_config.config.dex.name} ({network_type})")
    logger.info(f"Markets: {len(validated_config.markets)} configured")
    
    # Log detailed market information in debug mode
    for i, market in enumerate(validated_config.markets, 1):
        logger.debug(f"Market {i}: {market.hip3_name} -> {market.stork_spot_asset} (autocalc: {market.autocalculate_ext})")
    
    if creds:
        logger.info(f"Credentials: {creds}")
    if endpoint:
        logger.info(f"Endpoint: {endpoint}")

    if dry_run:
        logger.info("Dry run completed successfully - no actions executed")
        typer.echo("✅ Dry run: config validation passed, no actions executed.")
        raise typer.Exit(0)

    # TODO: Real pushing logic here - use validated_config
    logger.info("Push operation completed successfully")
    typer.echo("Pushed successfully.")

@app.callback()
def main(
    version: Optional[bool] = typer.Option(
        None, "--version", callback=lambda v: (print("hip3_pusher 0.1.0") or sys.exit(0)) if v else None,
        is_eager=True, help="Show version and exit"
    )
):
    # Runs before subcommands; useful for global flags in bigger CLIs.
    pass
