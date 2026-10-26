# ==============================================================================
# CMC Telecom LabShop - Load Testing Makefile
# ==============================================================================

HOST ?= https://app.thanhhieutiet.shop
WORKERS ?= 4

.PHONY: help perf security network down status scale logs

help:
	@echo "Available commands:"
	@echo "  make perf       Start cluster in Performance Benchmark mode (Range Header enabled)"
	@echo "  make security   Start cluster in WAF Security & Rate Limit test mode"
	@echo "  make network    Start cluster in Multi-ISP light routing test mode"
	@echo "  make scale      Scale workers: make scale WORKERS=8"
	@echo "  make down       Stop all containers"
	@echo "  make status     Check container status"
	@echo "  make logs       View Locust master logs"

perf:
	TEST_MODE=performance USE_RANGE_HEADER=true LOCUST_HOST=$(HOST) docker compose up -d --scale locust-worker=$(WORKERS)
	@echo "Started Performance Cluster at http://localhost:8089 and Grafana at http://localhost:3000"

security:
	TEST_MODE=security USE_RANGE_HEADER=false LOCUST_HOST=$(HOST) docker compose up -d --scale locust-worker=$(WORKERS)
	@echo "Started Security Cluster at http://localhost:8089 and Grafana at http://localhost:3000"

network:
	TEST_MODE=performance USE_RANGE_HEADER=true LOCUST_HOST=$(HOST) docker compose up -d --scale locust-worker=2
	@echo "Started Multi-ISP Routing Cluster at http://localhost:8089 and Grafana at http://localhost:3000"

scale:
	docker compose up -d --scale locust-worker=$(WORKERS)

down:
	docker compose down

status:
	docker compose ps

logs:
	docker compose logs -f locust-master
