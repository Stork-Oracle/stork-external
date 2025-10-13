#!/bin/bash

set -e

SERVICE=$1
TYPE=$2

if [ -z "$SERVICE" ]; then
  echo "Please provide the service name as an argument"
  exit 1
fi

if [ -z "$TYPE" ]; then
  echo "Please provide the type (dev|release) as an argument"
  exit 1
fi

if [ "$TYPE" == "release" ]; then
  TAG=$(cat version.txt)
  TYPE_TAG="latest"
elif [ "$TYPE" == "dev" ]; then
  TAG=$(git rev-parse --short=7 HEAD)
  TYPE_TAG="dev"
else
  echo "Invalid type"
  exit 1
fi

DOCKERHUB_USERNAME="storknetwork"

# Convert underscores to dashes in image name
IMAGE_NAME=${SERVICE//_/-}

docker buildx use stork-cli-builder
docker buildx build --platform linux/amd64,linux/arm64 -f ./docker/images/Dockerfile -t "$DOCKERHUB_USERNAME"/"$IMAGE_NAME":"$TAG" -t "$DOCKERHUB_USERNAME"/"$IMAGE_NAME":"$TYPE_TAG" --push --progress=plain . --build-arg SERVICE=$SERVICE
echo "Pushed image successfully"
