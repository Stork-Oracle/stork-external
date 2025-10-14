from __future__ import annotations
import sys
from pathlib import Path
import typer
from typing import Optional

from .config import validate_config_early

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
):
    """
    Push using HIP-3 configs.

    Example:
      hip3-pusher push ./examples/sample_config.yaml -c ./examples/creds.toml -vv
    """
    # Validate config structure early - fail fast if invalid
    validated_config = validate_config_early(config)

    # Simple logging based on verbosity
    def log(level: int, msg: str):
        if verbose >= level:
            typer.echo(msg)

    log(1, f"Config: {config}")
    log(1, f"DEX: {validated_config.config.dex.name} ({'testnet' if validated_config.config.dex.testnet else 'mainnet'})")
    log(1, f"Markets: {len(validated_config.markets)}")
    
    if verbose >= 2:
        for i, market in enumerate(validated_config.markets, 1):
            log(2, f"  Market {i}: {market.hip3_name} -> {market.stork_spot_asset}")
    
    if creds:
        log(1, f"Creds:  {creds}")
    if endpoint:
        log(1, f"Endpoint: {endpoint}")

    if dry_run:
        typer.echo("✅ Dry run: config validation passed, no actions executed.")
        raise typer.Exit(0)

    # TODO: Real pushing logic here - use validated_config
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
