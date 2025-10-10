"""HIP3 Pusher: a sample multi-file Python project."""

from .main import main

__all__ = ["main"]
__version__ = "0.1.0"

import logging
logging.getLogger(__name__).addHandler(logging.NullHandler())
