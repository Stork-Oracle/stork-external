"""
Configuration models and validation for HIP3 pusher.
"""

from __future__ import annotations
from pathlib import Path
from typing import Annotated, List, Literal, Union
import yaml
from pydantic import BaseModel, Field, ValidationError, field_validator, Tag
import typer


class StorkAsset(BaseModel):
    """Configuration for a Stork asset with a fixed identifier."""
    type: Literal["stork"] = Field(default="stork", description="Type discriminator for StorkAsset")
    identifier: str = Field(..., description="Stork asset identifier")

    @field_validator('identifier')
    @classmethod
    def validate_non_empty_identifier(cls, v: str) -> str:
        """Ensure identifier is not empty."""
        if not v or not v.strip():
            raise ValueError("Identifier cannot be empty or whitespace-only")
        return v.strip()


class Random(BaseModel):
    """Configuration for a random value generator that oscillates between min and max."""
    type: Literal["random"] = Field(default="random", description="Type discriminator for Random")
    min_value: float = Field(..., description="Minimum value for random oscillation")
    max_value: float = Field(..., description="Maximum value for random oscillation")

    @field_validator('max_value')
    @classmethod
    def validate_max_greater_than_or_equal_min(cls, v: float, info) -> float:
        """Ensure max_value is greater than or equal to min_value."""
        if 'min_value' in info.data and v < info.data['min_value']:
            raise ValueError("max_value must be greater than or equal to min_value")
        return v


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
    spot_asset: Annotated[Union[Annotated[StorkAsset, Tag("stork")], Annotated[Random, Tag("random")]], Field(discriminator="type", description="Stork spot asset configuration")]
    mark_asset: Annotated[Union[Annotated[StorkAsset, Tag("stork")], Annotated[Random, Tag("random")]], Field(discriminator="type", description="Stork mark asset configuration")]
    external_asset: Annotated[Union[Annotated[StorkAsset, Tag("stork")], Annotated[Random, Tag("random")]], Field(discriminator="type", description="Stork external asset configuration")]

    @field_validator('hip3_name')
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
