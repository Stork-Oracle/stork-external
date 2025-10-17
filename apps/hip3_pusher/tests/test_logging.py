"""
Tests for the logging module.
"""

import logging
import io
import sys
from hip3_pusher.logging import setup_logging, get_logger


def test_setup_logging_readable_format():
    """Test logging setup with readable format."""
    # Capture log output
    log_capture = io.StringIO()
    
    # Setup logging with readable format
    setup_logging(level="INFO", json_format=False)
    
    # Get logger and test it
    logger = get_logger("test_logger")
    
    # Add our capture handler
    handler = logging.StreamHandler(log_capture)
    handler.setLevel(logging.INFO)
    logger.addHandler(handler)
    
    # Test logging
    logger.info("Test message")
    
    # Check output
    output = log_capture.getvalue()
    assert "Test message" in output
    assert "test_logger" in output
    assert "INFO" in output


def test_setup_logging_json_format():
    """Test logging setup with JSON format."""
    # Capture log output
    log_capture = io.StringIO()
    
    # Setup logging with JSON format
    setup_logging(level="DEBUG", json_format=True)
    
    # Get logger and test it
    logger = get_logger("json_test_logger")
    
    # Add our capture handler
    handler = logging.StreamHandler(log_capture)
    handler.setLevel(logging.DEBUG)
    logger.addHandler(handler)
    
    # Test logging
    logger.debug("Debug message")
    
    # Check output
    output = log_capture.getvalue()
    assert "Debug message" in output
    assert "json_test_logger" in output
    assert "DEBUG" in output


def test_get_logger():
    """Test getting logger instances."""
    logger1 = get_logger("test1")
    logger2 = get_logger("test2")
    logger3 = get_logger()  # Should use default name
    
    assert logger1.name == "test1"
    assert logger2.name == "test2"
    assert logger3.name == "hip3_pusher.logging"  # Default from module
    
    # Same name should return same logger
    logger1_again = get_logger("test1")
    assert logger1 is logger1_again


def test_log_levels():
    """Test different log levels work correctly."""
    setup_logging(level="WARNING", json_format=False)
    
    log_capture = io.StringIO()
    logger = get_logger("level_test")
    
    handler = logging.StreamHandler(log_capture)
    handler.setLevel(logging.DEBUG)  # Capture everything
    logger.addHandler(handler)
    
    # These should be logged (WARNING level and above)
    logger.warning("Warning message")
    logger.error("Error message")
    logger.critical("Critical message")
    
    # This should NOT be logged (below WARNING level)
    logger.info("Info message")
    logger.debug("Debug message")
    
    output = log_capture.getvalue()
    assert "Warning message" in output
    assert "Error message" in output
    assert "Critical message" in output
    assert "Info message" not in output
    assert "Debug message" not in output
