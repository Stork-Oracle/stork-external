#!/bin/bash
set -e

# Execute the service binary with all passed arguments
exec "/app/${SERVICE}" "$@"
