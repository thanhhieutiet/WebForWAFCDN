#!/usr/bin/env bash
# Deploy script for EC2
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${ROOT_DIR}"

echo "=========================================="
echo "  Deploying LabShop on EC2 / Server"
echo "=========================================="

# Check .env file
if [ ! -f ".env" ]; then
  echo "[INFO] .env not found. Initializing from .env.example..."
  cp .env.example .env
  echo "[WARNING] Please review .env configuration (e.g. JWT_SECRET, SSL settings) if needed."
fi

# Check if native host Nginx is running
if systemctl is-active --quiet nginx 2>/dev/null; then
  echo "[WARNING] Host Nginx service is currently ACTIVE and binding port 80/443!"
  echo "[INFO] Stopping and disabling host Nginx to prevent port conflicts..."
  sudo systemctl stop nginx
  sudo systemctl disable nginx || true
  echo "[OK] Host Nginx stopped."
fi

echo "=== Pulling latest changes ==="
git pull --ff-only || {
  echo "[WARNING] git pull --ff-only failed. Proceeding with local repository state..."
}

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

echo "=== Running smoke test ==="
if [ -f "./bin/smoke" ]; then
  ./bin/smoke -base-url="http://localhost"
elif command -v go >/dev/null 2>&1; then
  go run ./cmd/smoke -base-url="http://localhost"
else
  echo "Testing via curl..."
  curl -fsS http://localhost/api/health
  echo ""
fi

echo "=========================================="
echo "  Deployment completed successfully!"
echo "=========================================="
