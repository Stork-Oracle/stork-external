#!/bin/bash

set -e

TAG="v1.0.0"
DOCKERHUB_USERNAME="storknetwork"
IMAGE_NAME="stork-cli"
docker build -f "$IMAGE_NAME".Dockerfile -t "$IMAGE_NAME":"$TAG" .
docker tag "$IMAGE_NAME":"$TAG" "$DOCKERHUB_USERNAME"/"$IMAGE_NAME":"$TAG"
docker push "$DOCKERHUB_USERNAME"/"$IMAGE_NAME":"$TAG"
echo "Pushed image successfully"