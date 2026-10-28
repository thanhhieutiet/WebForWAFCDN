#!/usr/bin/env bash
# Deploy Promtail Agent on EC2 to ship Nginx access logs to remote Loki
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${ROOT_DIR}"

LOKI_URL="${1:-${LOKI_URL:-http://localhost:3100}}"
# Normalize if user typed http:/ instead of http://
if [[ "${LOKI_URL}" == http:/* ]] && [[ "${LOKI_URL}" != http://* ]]; then
  LOKI_URL="http://${LOKI_URL#http:/}"
fi
if [[ "${LOKI_URL}" == https:/* ]] && [[ "${LOKI_URL}" != https://* ]]; then
  LOKI_URL="https://${LOKI_URL#https:/}"
fi
LOKI_PUSH_URL="${LOKI_URL%/}/loki/api/v1/push"

echo "=========================================="
echo "  Deploying Promtail Agent (EC2 Origin)"
echo "  Target Loki: ${LOKI_PUSH_URL}"
echo "=========================================="

mkdir -p logs/nginx

# Generate agent configuration
cat <<EOF > scripts/promtail-agent-config.yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0
  log_level: info

positions:
  filename: /tmp/positions.yaml

clients:
  - url: ${LOKI_PUSH_URL}
    batchwait: 1s
    batchsize: 1048576

scrape_configs:
  - job_name: nginx-access
    static_configs:
      - targets:
          - localhost
        labels:
          job: nginx
          service: labshop-origin
          __path__: /var/log/nginx/access.log

    pipeline_stages:
      - json:
          expressions:
            time_local: time_local
            status: status
            method: method
            host: host
            uri: uri
            remote_addr: remote_addr
            realip_remote_addr: realip_remote_addr
            xff: xff
            cf_ip: cf_ip
            true_client_ip: true_client_ip
            request_time: request_time
            upstream_response_time: upstream_response_time
            upstream_status: upstream_status
            body_bytes_sent: body_bytes_sent
            ua: ua
            referer: referer
            cdn_cache: cdn_cache
            ssl_protocol: ssl_protocol
      - labels:
          status:
          method:
          host:
          cdn_cache:
      - timestamp:
          source: time_local
          format: RFC3339

  - job_name: nginx-error
    static_configs:
      - targets:
          - localhost
        labels:
          job: nginx-error
          service: labshop-origin
          __path__: /var/log/nginx/error.log
EOF

echo "[INFO] Starting Promtail agent container..."
docker compose -f scripts/promtail-agent.yml up -d

echo "[SUCCESS] Promtail agent is running and shipping logs to ${LOKI_PUSH_URL}"
docker ps --filter "name=labshop-promtail-agent"
