#!/bin/bash

set -e

TAG="v1.0.0"
DOCKERHUB_USERNAME="storknetwork"
IMAGE_NAME="stork-cli"
docker buildx use stork-cli-builder
docker buildx build --platform linux/amd64,linux/arm64 -f "$IMAGE_NAME".Dockerfile -t "$DOCKERHUB_USERNAME"/"$IMAGE_NAME":"$TAG" --push .
echo "Pushed image successfully"