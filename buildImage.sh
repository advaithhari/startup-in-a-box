#!/bin/bash

cleanup() {
  echo "stopping program"

   sudo systemctl stop docker

   sudo systemctl stop docker.socket
}

sudo systemctl start docker

trap cleanup INT

# Path where your app lives (fixed spot)
APP_DIR="C:/Users/advai/Documents/projects/stonks/ports-backend"

# Name of the image you want to build
IMAGE_NAME="myapp:latest"

# Builder (you can swap this for another like paketobuildpacks/builder:base)
BUILDER="paketobuildpacks/builder:base"

echo "Building app from $APP_DIR with Buildpacks..."

# Move into app directory
cd "$APP_DIR" || {
  echo "App directory not found"
  exit 1
}

# Build the container image with pack
pack-cli build "$IMAGE_NAME" --path . --builder "$BUILDER"

# Check if build succeeded
if [ $? -eq 0 ]; then
  echo "✅ Build successful! Image created: $IMAGE_NAME"
  docker run -it --rm $IMAGE_NAME

else
  echo "❌ Build failed."
  systemctl stop docker
  systemctl stop docker.socket
  exit 1
fi
