"""
Tests for the config module functionality.
"""

import pytest
import tempfile
import yaml
from pathlib import Path
from pydantic import ValidationError
import typer

from hip3_pusher.config import (
    DexConfig,
    ConfigSection,
    MarketConfig,
    Hip3Config,
    load_and_validate_config,
    validate_config_early
)


class TestDexConfig:
    """Test DexConfig model."""
    
    def test_valid_dex_config(self):
        """Test creating a valid DexConfig."""
        config = DexConfig(name="hyperliquid")
        assert config.name == "hyperliquid"
        assert config.testnet is False  # default value
        
    def test_dex_config_with_testnet(self):
        """Test DexConfig with testnet flag."""
        config = DexConfig(name="hyperliquid", testnet=True)
        assert config.name == "hyperliquid"
        assert config.testnet is True
        
    def test_dex_config_missing_name(self):
        """Test DexConfig validation fails without name."""
        with pytest.raises(ValidationError):
            DexConfig()


class TestMarketConfig:
    """Test MarketConfig model."""
    
    def test_valid_market_config(self):
        """Test creating a valid MarketConfig."""
        config = MarketConfig(
            hip3_name="BTCUSD",
            stork_spot_asset="BTCUSD",
            stork_mark_asset="BTCUSD"
        )
        assert config.hip3_name == "BTCUSD"
        assert config.stork_spot_asset == "BTCUSD"
        assert config.stork_mark_asset == "BTCUSD"
        assert config.autocalculate_ext is False  # default value
        
    def test_market_config_with_autocalculate(self):
        """Test MarketConfig with autocalculate_ext enabled."""
        config = MarketConfig(
            hip3_name="ETHUSD",
            stork_spot_asset="ETHUSD",
            stork_mark_asset="ETHUSD",
            autocalculate_ext=True
        )
        assert config.autocalculate_ext is True
        
    def test_market_config_empty_strings(self):
        """Test MarketConfig validation fails with empty strings."""
        with pytest.raises(ValidationError) as exc_info:
            MarketConfig(
                hip3_name="",
                stork_spot_asset="BTCUSD",
                stork_mark_asset="BTCUSD"
            )
        assert "Field cannot be empty" in str(exc_info.value)
        
    def test_market_config_whitespace_strings(self):
        """Test MarketConfig validation fails with whitespace-only strings."""
        with pytest.raises(ValidationError) as exc_info:
            MarketConfig(
                hip3_name="   ",
                stork_spot_asset="BTCUSD",
                stork_mark_asset="BTCUSD"
            )
        assert "Field cannot be empty" in str(exc_info.value)
        
    def test_market_config_strips_whitespace(self):
        """Test MarketConfig strips whitespace from string fields."""
        config = MarketConfig(
            hip3_name="  BTCUSD  ",
            stork_spot_asset="  BTCUSD  ",
            stork_mark_asset="  BTCUSD  "
        )
        assert config.hip3_name == "BTCUSD"
        assert config.stork_spot_asset == "BTCUSD"
        assert config.stork_mark_asset == "BTCUSD"
        
    def test_market_config_missing_required_fields(self):
        """Test MarketConfig validation fails with missing required fields."""
        with pytest.raises(ValidationError):
            MarketConfig(hip3_name="BTCUSD")


class TestHip3Config:
    """Test Hip3Config model."""
    
    def test_valid_hip3_config(self):
        """Test creating a valid Hip3Config."""
        dex_config = DexConfig(name="hyperliquid", testnet=True)
        config_section = ConfigSection(dex=dex_config)
        market_config = MarketConfig(
            hip3_name="BTCUSD",
            stork_spot_asset="BTCUSD",
            stork_mark_asset="BTCUSD"
        )
        
        hip3_config = Hip3Config(
            config=config_section,
            markets=[market_config]
        )
        
        assert hip3_config.config.dex.name == "hyperliquid"
        assert hip3_config.config.dex.testnet is True
        assert len(hip3_config.markets) == 1
        assert hip3_config.markets[0].hip3_name == "BTCUSD"
        
    def test_hip3_config_multiple_markets(self):
        """Test Hip3Config with multiple markets."""
        dex_config = DexConfig(name="hyperliquid")
        config_section = ConfigSection(dex=dex_config)
        markets = [
            MarketConfig(
                hip3_name="BTCUSD",
                stork_spot_asset="BTCUSD",
                stork_mark_asset="BTCUSD"
            ),
            MarketConfig(
                hip3_name="ETHUSD",
                stork_spot_asset="ETHUSD",
                stork_mark_asset="ETHUSD"
            )
        ]
        
        hip3_config = Hip3Config(
            config=config_section,
            markets=markets
        )
        
        assert len(hip3_config.markets) == 2
        assert hip3_config.markets[0].hip3_name == "BTCUSD"
        assert hip3_config.markets[1].hip3_name == "ETHUSD"
        
    def test_hip3_config_empty_markets(self):
        """Test Hip3Config validation fails with empty markets list."""
        dex_config = DexConfig(name="hyperliquid")
        config_section = ConfigSection(dex=dex_config)
        
        with pytest.raises(ValidationError) as exc_info:
            Hip3Config(config=config_section, markets=[])
        assert "at least 1" in str(exc_info.value)
        
    def test_hip3_config_duplicate_market_names(self):
        """Test Hip3Config validation fails with duplicate market names."""
        dex_config = DexConfig(name="hyperliquid")
        config_section = ConfigSection(dex=dex_config)
        markets = [
            MarketConfig(
                hip3_name="BTCUSD",
                stork_spot_asset="BTCUSD",
                stork_mark_asset="BTCUSD"
            ),
            MarketConfig(
                hip3_name="BTCUSD",  # Duplicate name
                stork_spot_asset="BTCUSD2",
                stork_mark_asset="BTCUSD2"
            )
        ]
        
        with pytest.raises(ValidationError) as exc_info:
            Hip3Config(config=config_section, markets=markets)
        assert "must be unique" in str(exc_info.value)


class TestConfigLoading:
    """Test config file loading and validation functions."""
    
    def test_load_valid_config_file(self):
        """Test loading a valid config file."""
        config_data = {
            'config': {
                'dex': {
                    'name': 'hyperliquid',
                    'testnet': True
                }
            },
            'markets': [
                {
                    'hip3_name': 'BTCUSD',
                    'stork_spot_asset': 'BTCUSD',
                    'stork_mark_asset': 'BTCUSD',
                    'autocalculate_ext': True
                }
            ]
        }
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            yaml.dump(config_data, f)
            config_path = Path(f.name)
            
        try:
            config = load_and_validate_config(config_path)
            assert config.config.dex.name == 'hyperliquid'
            assert config.config.dex.testnet is True
            assert len(config.markets) == 1
            assert config.markets[0].hip3_name == 'BTCUSD'
            assert config.markets[0].autocalculate_ext is True
        finally:
            config_path.unlink()
            
    def test_load_config_file_not_found(self):
        """Test loading a non-existent config file."""
        non_existent_path = Path("/non/existent/config.yaml")
        
        with pytest.raises(typer.BadParameter) as exc_info:
            load_and_validate_config(non_existent_path)
        assert "Config file not found" in str(exc_info.value)
        
    def test_load_invalid_yaml(self):
        """Test loading a file with invalid YAML."""
        invalid_yaml = "config:\n  dex:\n    name: hyperliquid\n  invalid: [\n"
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            f.write(invalid_yaml)
            config_path = Path(f.name)
            
        try:
            with pytest.raises(typer.BadParameter) as exc_info:
                load_and_validate_config(config_path)
            assert "Invalid YAML" in str(exc_info.value)
        finally:
            config_path.unlink()
            
    def test_load_empty_config_file(self):
        """Test loading an empty config file."""
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            f.write("")
            config_path = Path(f.name)
            
        try:
            with pytest.raises(typer.BadParameter) as exc_info:
                load_and_validate_config(config_path)
            assert "Config file is empty" in str(exc_info.value)
        finally:
            config_path.unlink()
            
    def test_load_config_validation_error(self):
        """Test loading a config file with validation errors."""
        invalid_config_data = {
            'config': {
                'dex': {
                    'name': 'hyperliquid'
                }
            },
            'markets': []  # Empty markets list should fail validation
        }
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            yaml.dump(invalid_config_data, f)
            config_path = Path(f.name)
            
        try:
            with pytest.raises(typer.BadParameter) as exc_info:
                load_and_validate_config(config_path)
            assert "Config validation failed" in str(exc_info.value)
            assert "markets" in str(exc_info.value)
        finally:
            config_path.unlink()
            
    def test_load_config_missing_required_fields(self):
        """Test loading a config file with missing required fields."""
        incomplete_config_data = {
            'config': {
                'dex': {
                    'name': 'hyperliquid'
                }
            },
            'markets': [
                {
                    'hip3_name': 'BTCUSD',
                    'stork_spot_asset': 'BTCUSD'
                    # Missing stork_mark_asset
                }
            ]
        }
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            yaml.dump(incomplete_config_data, f)
            config_path = Path(f.name)
            
        try:
            with pytest.raises(typer.BadParameter) as exc_info:
                load_and_validate_config(config_path)
            assert "Config validation failed" in str(exc_info.value)
            assert "stork_mark_asset" in str(exc_info.value)
        finally:
            config_path.unlink()
            
    def test_validate_config_early_function(self):
        """Test the validate_config_early convenience function."""
        config_data = {
            'config': {
                'dex': {
                    'name': 'hyperliquid',
                    'testnet': False
                }
            },
            'markets': [
                {
                    'hip3_name': 'ETHUSD',
                    'stork_spot_asset': 'ETHUSD',
                    'stork_mark_asset': 'ETHUSD'
                }
            ]
        }
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            yaml.dump(config_data, f)
            config_path = Path(f.name)
            
        try:
            config = validate_config_early(config_path)
            assert config.config.dex.name == 'hyperliquid'
            assert config.config.dex.testnet is False
            assert len(config.markets) == 1
            assert config.markets[0].hip3_name == 'ETHUSD'
        finally:
            config_path.unlink()


class TestConfigIntegration:
    """Integration tests using the example config file."""
    
    def test_load_example_config(self):
        """Test loading the example config file from the project."""
        example_config_path = Path(__file__).parent.parent / "examples" / "test_config.yaml"
        
        if example_config_path.exists():
            config = load_and_validate_config(example_config_path)
            assert config.config.dex.name == "tsts"
            assert config.config.dex.testnet is True
            assert len(config.markets) == 1
            assert config.markets[0].hip3_name == "BTCUSD"
            assert config.markets[0].stork_spot_asset == "BTCUSD"
            assert config.markets[0].stork_mark_asset == "BTCUSD"
            assert config.markets[0].autocalculate_ext is True
        else:
            pytest.skip("Example config file not found")
