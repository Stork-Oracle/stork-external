#!/bin/bash

set -e

IMAGE_NAME=$1

if [ -z "$IMAGE_NAME" ]; then
  echo "Please provide the image name as an argument"
  exit 1
fi

TAG="v1.0.0"
DOCKERHUB_USERNAME="storknetwork"

docker buildx use stork-cli-builder
docker buildx build --platform linux/amd64,linux/arm64 -f "$IMAGE_NAME".Dockerfile -t "$DOCKERHUB_USERNAME"/"$IMAGE_NAME":"$TAG" --push --progress=plain .
echo "Pushed image successfully"
