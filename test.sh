#!/bin/bash
# Run tests inside Docker with cached modules
# Usage: ./test.sh

set -e

IMAGE="golang:1.22-alpine"
CONTAINER_NAME="agentdock-test"

echo "Running AgentDock tests..."

# Remove old test container if exists
docker rm -f "$CONTAINER_NAME" 2>/dev/null || true

# Use a persistent volume for Go module cache
docker volume create agentdock-gomod 2>/dev/null || true

# Run tests with module cache
docker run --rm \
  --name "$CONTAINER_NAME" \
  -v "$(pwd):/app" \
  -v agentdock-gomod:/go/pkg/mod \
  -w /app \
  -e CGO_ENABLED=0 \
  -e PORT=8080 \
  -e NTFY_ENABLED=0 \
  "$IMAGE" \
  sh -c "go test -v -count=1 ./..."

echo ""
echo "✅ All tests passed!"
