"""
Structured logging configuration for hip3_pusher.

This module provides a centralized logging configuration using structlog
for consistent, structured logging throughout the application.
"""

import sys
import logging
from typing import Any, Dict, Optional
import structlog
from structlog.types import FilteringBoundLogger


def configure_logging(
    level: str = "INFO",
    json_logs: bool = False,
    show_locals: bool = False,
    service_name: str = "hip3_pusher"
) -> None:
    """
    Configure structured logging for the application.
    
    Args:
        level: Log level (DEBUG, INFO, WARNING, ERROR, CRITICAL)
        json_logs: If True, output logs in JSON format. If False, use human-readable format.
        show_locals: If True, include local variables in exception logs (dev only)
        service_name: Name of the service for log context
    """
    # Configure standard library logging
    logging.basicConfig(
        format="%(message)s",
        stream=sys.stdout,
        level=getattr(logging, level.upper()),
    )
    
    # Common processors for all configurations
    shared_processors = [
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.StackInfoRenderer(),
    ]
    
    if show_locals:
        shared_processors.append(structlog.dev.set_exc_info)
    else:
        shared_processors.append(structlog.processors.format_exc_info)
    
    # Configure processors based on output format
    if json_logs:
        # JSON output for production/structured logging
        processors = shared_processors + [
            structlog.processors.TimeStamper(fmt="iso"),
            structlog.processors.JSONRenderer()
        ]
    else:
        # Human-readable output for development
        processors = shared_processors + [
            structlog.processors.TimeStamper(fmt="%Y-%m-%d %H:%M:%S"),
            structlog.dev.ConsoleRenderer(colors=True)
        ]
    
    # Configure structlog
    structlog.configure(
        processors=processors,
        wrapper_class=structlog.make_filtering_bound_logger(
            getattr(logging, level.upper())
        ),
        logger_factory=structlog.PrintLoggerFactory(),
        cache_logger_on_first_use=True,
    )
    
    # Add service context to all logs
    structlog.contextvars.clear_contextvars()
    structlog.contextvars.bind_contextvars(service=service_name)


def get_logger(name: Optional[str] = None, **context: Any) -> FilteringBoundLogger:
    """
    Get a logger instance with optional context.
    
    Args:
        name: Logger name (typically __name__)
        **context: Additional context to bind to this logger
        
    Returns:
        Configured structlog logger
    """
    logger = structlog.get_logger(name)
    if context:
        logger = logger.bind(**context)
    return logger


def bind_context(**context: Any) -> None:
    """
    Bind context variables that will be included in all subsequent log messages
    within the current context (thread/async task).
    
    Args:
        **context: Key-value pairs to add to log context
    """
    structlog.contextvars.bind_contextvars(**context)


def clear_context() -> None:
    """Clear all context variables."""
    structlog.contextvars.clear_contextvars()


def with_context(**context: Any):
    """
    Decorator to add context to all log messages within a function.
    
    Args:
        **context: Key-value pairs to add to log context
        
    Example:
        @with_context(operation="config_validation")
        def validate_config(config_path):
            logger = get_logger(__name__)
            logger.info("Starting validation", config_path=str(config_path))
            # ... validation logic ...
            logger.info("Validation completed")
    """
    def decorator(func):
        def wrapper(*args, **kwargs):
            # Store current context
            current_context = structlog.contextvars.get_contextvars()
            try:
                # Add new context
                bind_context(**context)
                return func(*args, **kwargs)
            finally:
                # Restore previous context
                structlog.contextvars.clear_contextvars()
                if current_context:
                    structlog.contextvars.bind_contextvars(**current_context)
        return wrapper
    return decorator


# Convenience functions for common logging patterns
def log_function_entry(logger: FilteringBoundLogger, func_name: str, **kwargs: Any) -> None:
    """Log function entry with parameters."""
    logger.debug("Function entry", function=func_name, **kwargs)


def log_function_exit(logger: FilteringBoundLogger, func_name: str, **kwargs: Any) -> None:
    """Log function exit with results."""
    logger.debug("Function exit", function=func_name, **kwargs)


def log_performance(logger: FilteringBoundLogger, operation: str, duration_ms: float, **kwargs: Any) -> None:
    """Log performance metrics."""
    logger.info("Performance metric", 
                operation=operation, 
                duration_ms=round(duration_ms, 2), 
                **kwargs)


def log_error_with_context(logger: FilteringBoundLogger, error: Exception, operation: str, **context: Any) -> None:
    """Log an error with full context."""
    logger.error("Operation failed",
                 operation=operation,
                 error_type=type(error).__name__,
                 error_message=str(error),
                 **context,
                 exc_info=True)
