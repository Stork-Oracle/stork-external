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
    StorkAsset,
    Random,
    load_and_validate_config,
    validate_config_early
)


class TestStorkAsset:
    """Test StorkAsset model."""

    def test_valid_stork_asset(self):
        """Test creating a valid StorkAsset."""
        asset = StorkAsset(identifier="BTCUSD")
        assert asset.identifier == "BTCUSD"

    def test_stork_asset_strips_whitespace(self):
        """Test StorkAsset strips whitespace from identifier."""
        asset = StorkAsset(identifier="  BTCUSD  ")
        assert asset.identifier == "BTCUSD"

    def test_stork_asset_empty_identifier(self):
        """Test StorkAsset validation fails with empty identifier."""
        with pytest.raises(ValidationError) as exc_info:
            StorkAsset(identifier="")
        assert "Identifier cannot be empty" in str(exc_info.value)

    def test_stork_asset_whitespace_only_identifier(self):
        """Test StorkAsset validation fails with whitespace-only identifier."""
        with pytest.raises(ValidationError) as exc_info:
            StorkAsset(identifier="   ")
        assert "Identifier cannot be empty" in str(exc_info.value)

    def test_stork_asset_missing_identifier(self):
        """Test StorkAsset validation fails without identifier."""
        with pytest.raises(ValidationError):
            StorkAsset()


class TestRandom:
    """Test Random model."""

    def test_valid_random(self):
        """Test creating a valid Random configuration."""
        random = Random(min_value=100.0, max_value=200.0)
        assert random.min_value == 100.0
        assert random.max_value == 200.0

    def test_random_negative_values(self):
        """Test Random with negative values."""
        random = Random(min_value=-50.0, max_value=50.0)
        assert random.min_value == -50.0
        assert random.max_value == 50.0

    def test_random_max_equals_min(self):
        """Test Random allows max to equal min (constant value)."""
        random = Random(min_value=100.0, max_value=100.0)
        assert random.min_value == 100.0
        assert random.max_value == 100.0

    def test_random_max_less_than_min(self):
        """Test Random validation fails when max is less than min."""
        with pytest.raises(ValidationError) as exc_info:
            Random(min_value=200.0, max_value=100.0)
        assert "max_value must be greater than or equal to min_value" in str(exc_info.value)

    def test_random_missing_required_fields(self):
        """Test Random validation fails with missing required fields."""
        with pytest.raises(ValidationError):
            Random(min_value=100.0)
        with pytest.raises(ValidationError):
            Random(max_value=200.0)


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
    """Test MarketConfig model with Union types."""

    def test_market_config_all_stork_assets(self):
        """Test MarketConfig with all fields as StorkAsset."""
        config = MarketConfig(
            hip3_name="BTCUSD",
            spot_asset=StorkAsset(identifier="BTCUSD"),
            mark_asset=StorkAsset(identifier="BTCUSD"),
            external_asset=StorkAsset(identifier="BTCUSD")
        )
        assert config.hip3_name == "BTCUSD"
        assert isinstance(config.spot_asset, StorkAsset)
        assert config.spot_asset.identifier == "BTCUSD"
        assert isinstance(config.mark_asset, StorkAsset)
        assert config.mark_asset.identifier == "BTCUSD"
        assert isinstance(config.external_asset, StorkAsset)
        assert config.external_asset.identifier == "BTCUSD"

    def test_market_config_all_random(self):
        """Test MarketConfig with all fields as Random."""
        config = MarketConfig(
            hip3_name="ETHUSD",
            spot_asset=Random(min_value=1000.0, max_value=2000.0),
            mark_asset=Random(min_value=1000.0, max_value=2000.0),
            external_asset=Random(min_value=1000.0, max_value=2000.0)
        )
        assert config.hip3_name == "ETHUSD"
        assert isinstance(config.spot_asset, Random)
        assert config.spot_asset.min_value == 1000.0
        assert config.spot_asset.max_value == 2000.0
        assert isinstance(config.mark_asset, Random)
        assert isinstance(config.external_asset, Random)

    def test_market_config_mixed_spot_random_others_asset(self):
        """Test MarketConfig with spot as Random, others as StorkAsset."""
        config = MarketConfig(
            hip3_name="SOLUSD",
            spot_asset=Random(min_value=50.0, max_value=150.0),
            mark_asset=StorkAsset(identifier="SOLUSD"),
            external_asset=StorkAsset(identifier="SOLUSD")
        )
        assert config.hip3_name == "SOLUSD"
        assert isinstance(config.spot_asset, Random)
        assert config.spot_asset.min_value == 50.0
        assert isinstance(config.mark_asset, StorkAsset)
        assert isinstance(config.external_asset, StorkAsset)

    def test_market_config_mixed_mark_random_others_asset(self):
        """Test MarketConfig with mark as Random, others as StorkAsset."""
        config = MarketConfig(
            hip3_name="AVAXUSD",
            spot_asset=StorkAsset(identifier="AVAXUSD"),
            mark_asset=Random(min_value=20.0, max_value=40.0),
            external_asset=StorkAsset(identifier="AVAXUSD")
        )
        assert isinstance(config.spot_asset, StorkAsset)
        assert isinstance(config.mark_asset, Random)
        assert isinstance(config.external_asset, StorkAsset)

    def test_market_config_mixed_external_random_others_asset(self):
        """Test MarketConfig with external as Random, others as StorkAsset."""
        config = MarketConfig(
            hip3_name="MATICUSD",
            spot_asset=StorkAsset(identifier="MATICUSD"),
            mark_asset=StorkAsset(identifier="MATICUSD"),
            external_asset=Random(min_value=0.5, max_value=1.5)
        )
        assert isinstance(config.spot_asset, StorkAsset)
        assert isinstance(config.mark_asset, StorkAsset)
        assert isinstance(config.external_asset, Random)

    def test_market_config_mixed_all_different_types(self):
        """Test MarketConfig with each field using different Random ranges."""
        config = MarketConfig(
            hip3_name="LINKUSD",
            spot_asset=Random(min_value=5.0, max_value=10.0),
            mark_asset=Random(min_value=10.0, max_value=20.0),
            external_asset=Random(min_value=15.0, max_value=25.0)
        )
        assert isinstance(config.spot_asset, Random)
        assert config.spot_asset.min_value == 5.0
        assert config.spot_asset.max_value == 10.0
        assert isinstance(config.mark_asset, Random)
        assert config.mark_asset.min_value == 10.0
        assert isinstance(config.external_asset, Random)
        assert config.external_asset.max_value == 25.0

    def test_market_config_empty_hip3_name(self):
        """Test MarketConfig validation fails with empty hip3_name."""
        with pytest.raises(ValidationError) as exc_info:
            MarketConfig(
                hip3_name="",
                spot_asset=StorkAsset(identifier="BTCUSD"),
                mark_asset=StorkAsset(identifier="BTCUSD"),
                external_asset=StorkAsset(identifier="BTCUSD")
            )
        assert "Field cannot be empty" in str(exc_info.value)

    def test_market_config_whitespace_hip3_name(self):
        """Test MarketConfig validation fails with whitespace-only hip3_name."""
        with pytest.raises(ValidationError) as exc_info:
            MarketConfig(
                hip3_name="   ",
                spot_asset=StorkAsset(identifier="BTCUSD"),
                mark_asset=StorkAsset(identifier="BTCUSD"),
                external_asset=StorkAsset(identifier="BTCUSD")
            )
        assert "Field cannot be empty" in str(exc_info.value)

    def test_market_config_strips_whitespace_from_hip3_name(self):
        """Test MarketConfig strips whitespace from hip3_name."""
        config = MarketConfig(
            hip3_name="  BTCUSD  ",
            spot_asset=StorkAsset(identifier="BTCUSD"),
            mark_asset=StorkAsset(identifier="BTCUSD"),
            external_asset=StorkAsset(identifier="BTCUSD")
        )
        assert config.hip3_name == "BTCUSD"

    def test_market_config_missing_required_fields(self):
        """Test MarketConfig validation fails with missing required fields."""
        with pytest.raises(ValidationError):
            MarketConfig(hip3_name="BTCUSD")
        with pytest.raises(ValidationError):
            MarketConfig(
                hip3_name="BTCUSD",
                spot_asset=StorkAsset(identifier="BTCUSD")
            )
        with pytest.raises(ValidationError):
            MarketConfig(
                hip3_name="BTCUSD",
                spot_asset=StorkAsset(identifier="BTCUSD"),
                mark_asset=StorkAsset(identifier="BTCUSD")
            )


class TestHip3Config:
    """Test Hip3Config model."""

    def test_valid_hip3_config(self):
        """Test creating a valid Hip3Config."""
        dex_config = DexConfig(name="hyperliquid", testnet=True)
        config_section = ConfigSection(dex=dex_config)
        market_config = MarketConfig(
            hip3_name="BTCUSD",
            spot_asset=StorkAsset(identifier="BTCUSD"),
            mark_asset=StorkAsset(identifier="BTCUSD"),
            external_asset=StorkAsset(identifier="BTCUSD")
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
                spot_asset=StorkAsset(identifier="BTCUSD"),
                mark_asset=StorkAsset(identifier="BTCUSD"),
                external_asset=StorkAsset(identifier="BTCUSD")
            ),
            MarketConfig(
                hip3_name="ETHUSD",
                spot_asset=StorkAsset(identifier="ETHUSD"),
                mark_asset=StorkAsset(identifier="ETHUSD"),
                external_asset=StorkAsset(identifier="ETHUSD")
            )
        ]

        hip3_config = Hip3Config(
            config=config_section,
            markets=markets
        )

        assert len(hip3_config.markets) == 2
        assert hip3_config.markets[0].hip3_name == "BTCUSD"
        assert hip3_config.markets[1].hip3_name == "ETHUSD"

    def test_hip3_config_multiple_markets_mixed_types(self):
        """Test Hip3Config with multiple markets using mixed union types."""
        dex_config = DexConfig(name="hyperliquid")
        config_section = ConfigSection(dex=dex_config)
        markets = [
            MarketConfig(
                hip3_name="BTCUSD",
                spot_asset=StorkAsset(identifier="BTCUSD"),
                mark_asset=StorkAsset(identifier="BTCUSD"),
                external_asset=StorkAsset(identifier="BTCUSD")
            ),
            MarketConfig(
                hip3_name="ETHUSD",
                spot_asset=Random(min_value=1000.0, max_value=2000.0),
                mark_asset=Random(min_value=1000.0, max_value=2000.0),
                external_asset=Random(min_value=1000.0, max_value=2000.0)
            ),
            MarketConfig(
                hip3_name="SOLUSD",
                spot_asset=Random(min_value=50.0, max_value=150.0),
                mark_asset=StorkAsset(identifier="SOLUSD"),
                external_asset=StorkAsset(identifier="SOLUSD")
            )
        ]

        hip3_config = Hip3Config(
            config=config_section,
            markets=markets
        )

        assert len(hip3_config.markets) == 3
        assert isinstance(hip3_config.markets[0].spot_asset, StorkAsset)
        assert isinstance(hip3_config.markets[1].spot_asset, Random)
        assert isinstance(hip3_config.markets[2].spot_asset, Random)
        assert isinstance(hip3_config.markets[2].mark_asset, StorkAsset)

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
                spot_asset=StorkAsset(identifier="BTCUSD"),
                mark_asset=StorkAsset(identifier="BTCUSD"),
                external_asset=StorkAsset(identifier="BTCUSD")
            ),
            MarketConfig(
                hip3_name="BTCUSD",  # Duplicate name
                spot_asset=StorkAsset(identifier="BTCUSD2"),
                mark_asset=StorkAsset(identifier="BTCUSD2"),
                external_asset=StorkAsset(identifier="BTCUSD2")
            )
        ]

        with pytest.raises(ValidationError) as exc_info:
            Hip3Config(config=config_section, markets=markets)
        assert "must be unique" in str(exc_info.value)


class TestConfigLoading:
    """Test config file loading and validation functions."""

    def test_load_valid_config_file_with_stork_assets(self):
        """Test loading a valid config file with StorkAsset types."""
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
                    'spot_asset': {
                        'identifier': 'BTCUSD'
                    },
                    'mark_asset': {
                        'identifier': 'BTCUSD'
                    },
                    'external_asset': {
                        'identifier': 'BTCUSD'
                    }
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
            assert isinstance(config.markets[0].spot_asset, StorkAsset)
            assert config.markets[0].spot_asset.identifier == 'BTCUSD'
        finally:
            config_path.unlink()

    def test_load_valid_config_file_with_random(self):
        """Test loading a valid config file with Random types."""
        config_data = {
            'config': {
                'dex': {
                    'name': 'hyperliquid',
                    'testnet': True
                }
            },
            'markets': [
                {
                    'hip3_name': 'ETHUSD',
                    'spot_asset': {
                        'min_value': 1000.0,
                        'max_value': 2000.0
                    },
                    'mark_asset': {
                        'min_value': 1000.0,
                        'max_value': 2000.0
                    },
                    'external_asset': {
                        'min_value': 1000.0,
                        'max_value': 2000.0
                    }
                }
            ]
        }

        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            yaml.dump(config_data, f)
            config_path = Path(f.name)

        try:
            config = load_and_validate_config(config_path)
            assert config.config.dex.name == 'hyperliquid'
            assert len(config.markets) == 1
            assert config.markets[0].hip3_name == 'ETHUSD'
            assert isinstance(config.markets[0].spot_asset, Random)
            assert config.markets[0].spot_asset.min_value == 1000.0
            assert config.markets[0].spot_asset.max_value == 2000.0
        finally:
            config_path.unlink()

    def test_load_valid_config_file_with_mixed_types(self):
        """Test loading a valid config file with mixed StorkAsset and Random types."""
        config_data = {
            'config': {
                'dex': {
                    'name': 'hyperliquid',
                    'testnet': False
                }
            },
            'markets': [
                {
                    'hip3_name': 'SOLUSD',
                    'spot_asset': {
                        'min_value': 50.0,
                        'max_value': 150.0
                    },
                    'mark_asset': {
                        'identifier': 'SOLUSD'
                    },
                    'external_asset': {
                        'identifier': 'SOLUSD'
                    }
                }
            ]
        }

        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            yaml.dump(config_data, f)
            config_path = Path(f.name)

        try:
            config = load_and_validate_config(config_path)
            assert len(config.markets) == 1
            assert isinstance(config.markets[0].spot_asset, Random)
            assert isinstance(config.markets[0].mark_asset, StorkAsset)
            assert isinstance(config.markets[0].external_asset, StorkAsset)
            assert config.markets[0].mark_asset.identifier == 'SOLUSD'
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
                    'spot_asset': {
                        'identifier': 'BTCUSD'
                    }
                    # Missing mark_asset and external_asset
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
                    'spot_asset': {
                        'identifier': 'ETHUSD'
                    },
                    'mark_asset': {
                        'identifier': 'ETHUSD'
                    },
                    'external_asset': {
                        'identifier': 'ETHUSD'
                    }
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
            assert len(config.markets) == 2

            # First market uses Random types
            assert config.markets[0].hip3_name == "TEST2"
            assert isinstance(config.markets[0].spot_asset, Random)
            assert isinstance(config.markets[0].mark_asset, Random)
            assert isinstance(config.markets[0].external_asset, Random)

            # Second market uses StorkAsset types
            assert config.markets[1].hip3_name == "BTCUSD"
            assert isinstance(config.markets[1].spot_asset, StorkAsset)
            assert config.markets[1].spot_asset.identifier == "BTCUSD"
            assert isinstance(config.markets[1].mark_asset, StorkAsset)
            assert config.markets[1].mark_asset.identifier == "BTCUSD"
            assert isinstance(config.markets[1].external_asset, StorkAsset)
            assert config.markets[1].external_asset.identifier == "BTCUSD"
        else:
            pytest.skip("Example config file not found")
