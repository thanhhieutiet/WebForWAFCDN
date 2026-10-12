#!/usr/bin/env bash
# Deploy script for EC2
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${ROOT_DIR}"

echo "=== Pulling latest changes ==="
git pull --ff-only

echo "=== Building and deploying Docker containers ==="
docker compose up -d --build

echo "=== Waiting for API service to become healthy ==="
MAX_RETRIES=30
RETRY_COUNT=0
HEALTHY=false

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  STATUS=$(docker inspect --format='{{json .State.Health.Status}}' "$(docker compose ps -q api)" 2>/dev/null || echo '"unknown"')
  if [ "$STATUS" = '"healthy"' ]; then
    echo "API is healthy!"
    HEALTHY=true
    break
  fi
  echo "Waiting for health check... ($((RETRY_COUNT + 1))/$MAX_RETRIES) - current status: $STATUS"
  sleep 2
  RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ "$HEALTHY" = false ]; then
  echo "Error: API container failed to become healthy."
  docker compose logs api
  exit 1
fi

echo "=== Running smoke test if present ==="
if [ -f "./bin/smoke" ]; then
  ./bin/smoke -base-url="http://localhost"
elif [ -d "./cmd/smoke" ]; then
  go run ./cmd/smoke -base-url="http://localhost"
else
  echo "Smoke test binary not found, verifying via curl..."
  curl -fsS http://localhost/api/health
  echo ""
fi

echo "=== Deployment completed successfully! ==="
