#!/bin/bash
# OrionOS Install Script
# Usage: curl -fsSL https://raw.githubusercontent.com/marvel-254/orionos/main/install.sh | sh

set -e

REPO="marvel-254/orionos"
IMAGE="ghcr.io/${REPO}:latest"
CONTAINER_NAME="orionos"
PORT="${ORIONOS_PORT:-8080}"

echo ""
echo "╔═══════════════════════════════════════════╗"
echo "║  OrionOS — Lightweight AI Agent Monitor   ║"
echo "╚═══════════════════════════════════════════╝"
echo ""

if ! command -v docker &>/dev/null; then
  echo "❌ Docker is required. Install it first:"
  echo "   https://docs.docker.com/engine/install/"
  exit 1
fi

USE_COMPOSE=false
if command -v docker compose &>/dev/null || docker compose version &>/dev/null 2>&1; then
  USE_COMPOSE=true
fi

echo "📦 Pulling OrionOS image..."
docker pull "${IMAGE}" 2>&1 || {
  echo "⚠️  Could not pull from GHCR. Building locally..."
  
  if command -v git &>/dev/null; then
    TMPDIR=$(mktemp -d)
    git clone --depth 1 "https://github.com/${REPO}.git" "${TMPDIR}/orionos"
    cd "${TMPDIR}/orionos"
    docker build -t "${IMAGE}" .
    cd -
    rm -rf "${TMPDIR}"
  else
    echo "❌ Cannot build locally. Install git first."
    exit 1
  fi
}

if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
  echo "🔴 Stopping existing container..."
  docker stop "${CONTAINER_NAME}" 2>/dev/null || true
  docker rm "${CONTAINER_NAME}" 2>/dev/null || true
fi

echo "🚀 Starting OrionOS..."

if [ "$USE_COMPOSE" = true ] && [ -f "docker-compose.yml" ]; then
  docker compose up -d
else
  docker run -d \
    --name "${CONTAINER_NAME}" \
    --restart unless-stopped \
    -p "${PORT}:8080" \
    -v orionos-data:/data \
    -v /proc:/host/proc:ro \
    -e HOST_PROC=/host/proc \
    -e TZ="$(cat /etc/timezone 2>/dev/null || echo 'UTC')" \
    -e NTFY_ENABLED=0 \
    "${IMAGE}"
fi

echo ""
echo "✅ OrionOS is running!"
echo ""
echo "   Dashboard:  http://localhost:${PORT}"
echo "   Health:     http://localhost:${PORT}/api/health"
echo "   Agents API: http://localhost:${PORT}/api/agents"
echo ""
echo "📋 Commands:"
echo "   Stop:    docker stop ${CONTAINER_NAME}"
echo "   Logs:    docker logs -f ${CONTAINER_NAME}"
echo "   Remove:  docker rm -f ${CONTAINER_NAME}"
echo ""
echo "🔔 To enable ntfy notifications:"
echo "   Set NTFY_URL, NTFY_TOPIC env vars"
echo "   Then: docker restart ${CONTAINER_NAME}"
echo ""
echo "🐑 Happy herding!"
