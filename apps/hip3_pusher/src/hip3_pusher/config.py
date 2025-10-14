"""
Configuration models and validation for HIP3 pusher.
"""

from __future__ import annotations
from pathlib import Path
from typing import List
import yaml
from pydantic import BaseModel, Field, ValidationError, field_validator
import typer


class DexConfig(BaseModel):
    """Configuration for the DEX."""
    name: str = Field(..., description="Name of the DEX")
    testnet: bool = Field(default=False, description="Whether this is a testnet configuration")


class ConfigSection(BaseModel):
    """Top-level config section."""
    dex: DexConfig = Field(..., description="DEX configuration")


class MarketConfig(BaseModel):
    """Configuration for a single market."""
    hip3_name: str = Field(..., description="HIP3 market name")
    stork_spot_asset: str = Field(..., description="Stork spot asset identifier")
    stork_mark_asset: str = Field(..., description="Stork mark asset identifier")
    autocalculate_ext: bool = Field(default=False, description="Whether to auto-calculate external data")

    @field_validator('hip3_name', 'stork_spot_asset', 'stork_mark_asset')
    @classmethod
    def validate_non_empty_string(cls, v: str) -> str:
        """Ensure string fields are not empty."""
        if not v or not v.strip():
            raise ValueError("Field cannot be empty or whitespace-only")
        return v.strip()


class Hip3Config(BaseModel):
    """Complete HIP3 pusher configuration."""
    config: ConfigSection = Field(..., description="Configuration section")
    markets: List[MarketConfig] = Field(..., min_length=1, description="List of markets to process")

    @field_validator('markets')
    @classmethod
    def validate_unique_market_names(cls, markets: List[MarketConfig]) -> List[MarketConfig]:
        """Ensure all market names are unique."""
        hip3_names = [market.hip3_name for market in markets]
        if len(hip3_names) != len(set(hip3_names)):
            raise ValueError("Market hip3_name values must be unique")
        return markets


def load_and_validate_config(config_path: Path) -> Hip3Config:
    """
    Load and validate config file early in program lifecycle.
    
    Args:
        config_path: Path to the YAML config file
        
    Returns:
        Validated Hip3Config object
        
    Raises:
        typer.BadParameter: If the config is invalid or cannot be loaded
    """
    try:
        with open(config_path, 'r', encoding='utf-8') as f:
            raw_config = yaml.safe_load(f)
    except yaml.YAMLError as e:
        raise typer.BadParameter(f"Invalid YAML in config file: {e}")
    except FileNotFoundError:
        raise typer.BadParameter(f"Config file not found: {config_path}")
    except Exception as e:
        raise typer.BadParameter(f"Error reading config file: {e}")

    if raw_config is None:
        raise typer.BadParameter("Config file is empty")

    try:
        # This will raise ValidationError if invalid
        return Hip3Config(**raw_config)
    except ValidationError as e:
        # Format validation errors nicely
        error_messages = []
        for error in e.errors():
            field_path = " -> ".join(str(loc) for loc in error['loc'])
            error_messages.append(f"  {field_path}: {error['msg']}")
        
        formatted_errors = "\n".join(error_messages)
        raise typer.BadParameter(f"Config validation failed:\n{formatted_errors}")


def validate_config_early(config_path: Path) -> Hip3Config:
    """
    Convenience function for early config validation in CLI.
    
    This function is designed to be called early in the program lifecycle
    to fail fast if the configuration is invalid.
    """
    return load_and_validate_config(config_path)
