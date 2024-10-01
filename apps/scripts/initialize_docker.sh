#!/bin/bash

set -e

docker login
docker buildx create --use --name stork-cli-builder --driver docker-container
