#!/bin/bash

set -e

SERVICE=$1

if [ -z "$SERVICE" ]; then
  echo "Please provide the service name as an argument"
  exit 1
fi

TAG=$(cat version.txt)
DOCKERHUB_USERNAME="storknetwork"

# Convert underscores to dashes in image name
IMAGE_NAME=${SERVICE//_/-}

docker buildx use stork-cli-builder
docker buildx build --platform linux/amd64,linux/arm64 -f Dockerfile -t "$DOCKERHUB_USERNAME"/"$IMAGE_NAME":"$TAG" --push --progress=plain . --build-arg SERVICE=$SERVICE
echo "Pushed image successfully"
