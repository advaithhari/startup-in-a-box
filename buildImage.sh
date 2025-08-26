#!/bin/bash

cleanup() {
  echo "stopping program"

    systemctl stop docker

    systemctl stop docker.socket
}
apt update
apt install -y docker.io
(curl -sSL "https://github.com/buildpacks/pack/releases/download/v0.38.2/pack-v0.38.2-linux.tgz" | sudo tar -C /usr/local/bin/ --no-same-owner -xzv pack)
systemctl start docker

trap cleanup INT

# Path where your app lives (fixed spot)
APP_DIR="/mnt/c/Users/advai/Documents/projects/stonks/ports-backend/ports"

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
pack build "$IMAGE_NAME" --path . --builder "$BUILDER"

# Check if build succeeded
if [ $? -eq 0 ]; then
  echo "Build successful! Image created: $IMAGE_NAME"
  docker run --rm $IMAGE_NAME

else
  echo "❌ Build failed."
  systemctl stop docker
  systemctl stop docker.socket
  exit 1
fi
