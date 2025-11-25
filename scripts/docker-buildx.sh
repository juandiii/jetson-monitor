#!/bin/sh
set -euo pipefail

# This script expects the following environment variables to be set by the caller:
# DOCKER_PUSH, PLATFORMS, LOCAL_PLATFORM, PROJECT, VERSION

if [ -z "${PLATFORMS:-}" ]; then
  echo "PLATFORMS is required" >&2
  exit 2
fi
if [ -z "${LOCAL_PLATFORM:-}" ]; then
  echo "LOCAL_PLATFORM is required" >&2
  exit 2
fi
if [ -z "${PROJECT:-}" ]; then
  echo "PROJECT is required" >&2
  exit 2
fi
if [ -z "${VERSION:-}" ]; then
  echo "VERSION is required" >&2
  exit 2
fi

DOCKER_PUSH=${DOCKER_PUSH:-false}

# Detect if buildx supports the necessary build command; if not, fall back to docker build
if ! docker buildx version >/dev/null 2>&1; then
  echo "docker buildx not available; falling back to 'docker build' for local platform"
  docker build -t "$PROJECT:$VERSION" .
  exit 0
fi

if [ "$DOCKER_PUSH" = "true" ]; then
    echo "Running multi-arch build and pushing to registry for platforms: $PLATFORMS"
    docker buildx build --platform "$PLATFORMS" -t "$PROJECT:$VERSION" --pull --progress=plain --push .
else
    if [ "$PLATFORMS" = "$LOCAL_PLATFORM" ]; then
        echo "Building for local platform $LOCAL_PLATFORM with buildx (loading into Docker)..."
        docker buildx build --platform "$LOCAL_PLATFORM" -t "$PROJECT:$VERSION" --pull --progress=plain --load .
    else
        echo "PLATFORMS ($PLATFORMS) contains multiple platforms. To produce a multi-arch image you must push to a registry."
        echo "Falling back to building and loading local platform $LOCAL_PLATFORM..."
        docker buildx build --platform "$LOCAL_PLATFORM" -t "$PROJECT:$VERSION" --pull --progress=plain --load .
    fi
fi
