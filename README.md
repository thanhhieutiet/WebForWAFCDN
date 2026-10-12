# LabShop — Origin Server & Security Lab for CMC WAAP & Multi-CDN

[![Go Version](https://img.shields.io/badge/go-1.26+-00ADD8?style=flat&logo=go)](https://golang.org)
[![OpenAPI 3.0.3](https://img.shields.io/badge/openapi-3.0.3-6BA539?style=flat&logo=openapi-initiative)](./api/openapi.yaml)
[![Architecture](https://img.shields.io/badge/architecture-Clean%20Go-blue?style=flat)](./docs/ARCHITECTURE.md)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)

**LabShop** is a production-grade e-commerce reference application and interactive security laboratory specifically designed to evaluate, test, and demonstrate **CMC WAAP (Web Application & API Protection)** and **CMC Multi-CDN** solutions.

It couples a complete shop engine (catalog, cart, coupons, idempotent checkout, payment gateway, reviews, ticketing, admin KPI) with synthetic lab endpoints to test edge caching, revalidation, compression, rate limiting, and attack vectors.

---

## Table of Contents

- [Topology & 4 Routing Hostnames](#topology--4-routing-hostnames)
- [Architecture Overview](#architecture-overview)
- [Quick Start & Local Development](#quick-start--local-development)
- [Running with Docker Compose](#running-with-docker-compose)
- [Automated Smoke Testing](#automated-smoke-testing)
- [EC2 Production Deployment](#ec2-production-deployment)
- [Environment Variables](#environment-variables)
- [Lab Console & Interactive Tutorial](#lab-console--interactive-tutorial)
- [API Documentation](#api-documentation)

---

## Topology & 4 Routing Hostnames

LabShop is designed to be mapped to **4 distinct hostnames / DNS routes** pointing to the same backend application, allowing students and security engineers to compare baseline origin behavior against protected edge layers:

```
                          ┌─────────────────────────────┐
                          │   Client / Browser / CLI    │
                          └──────────────┬──────────────┘
                                         │
        ┌───────────────────┬────────────┴─────────────┬────────────────────┐
        │ 1. Direct         │ 2. WAAP Only             │ 3. CDN Only        │ 4. Full (Multi-CDN + WAAP)
        ▼                   ▼                          ▼                    ▼
 ┌───────────────┐   ┌───────────────┐          ┌───────────────┐    ┌───────────────┐
 │ direct.<host> │   │  waap.<host>  │          │  cdn.<host>   │    │  full.<host>  │
 └───────┬───────┘   └───────┬───────┘          └───────┬───────┘    └───────┬───────┘
         │                   │                          │                    │
         │           ┌───────▼───────┐          ┌───────▼───────┐    ┌───────▼───────┐
         │           │   CMC WAAP    │          │ CMC Multi-CDN │    │ CMC Multi-CDN │
         │           │(Inspection)   │          │ (Edge Cache)  │    │ (Edge Cache)  │
         │           └───────┬───────┘          └───────┬───────┘    └───────┬───────┘
         │                   │                          │                    │
         │                   │                          │            ┌───────▼───────┐
         │                   │                          │            │   CMC WAAP    │
         │                   │                          │            │(Inspection)   │
         │                   │                          │            └───────┬───────┘
         │                   │                          │                    │
         └───────────────────┴────────────┬─────────────┴────────────────────┘
                                          ▼
                                ┌───────────────────┐
                                │  LabShop Origin   │
                                │   (EC2 / Docker)  │
                                └───────────────────┘
```

| Hostname | Routing Path | Primary Use Case & Testing Objectives |
| :--- | :--- | :--- |
| `direct.<domain>` | **Client ➔ Origin** | **Baseline Control**: Direct origin access bypassing edge layers. Used to measure true origin latency, inspect raw HTTP headers, test un-cached requests, and verify backend state. |
| `waap.<domain>` | **Client ➔ CMC WAAP ➔ Origin** | **WAF / WAAP Security Testing**: Evaluates OWASP Top 10 rule enforcement (SQLi, XSS, Path Traversal), API schema validation, rate limiting, and Bot mitigation without cache interference. |
| `cdn.<domain>` | **Client ➔ Multi-CDN ➔ Origin** | **Multi-CDN & Caching Testing**: Verifies cache HIT/MISS ratio, `Cache-Control` max-age, ETag 304 revalidation, Gzip/Brotli edge compression, failover, and fast purges. |
| `full.<domain>` | **Client ➔ Multi-CDN ➔ WAAP ➔ Origin** | **Production Defense-in-Depth**: Complete end-to-end topology combining edge acceleration, CDN shielding, and cloud WAAP security inspection. |

---

## Architecture Overview

The backend is built in pure **Go** adhering strictly to **Clean Architecture** principles:
- **`internal/domain`**: Zero external dependencies (Go standard library only). No JSON tags, no database drivers. Contains core entities (`Money`, `User`, `Product`, `Order`, `Review`, `Ticket`).
- **`internal/usecase`**: Pure application business rules and abstract Port definitions (`UserRepository`, `CatalogRepository`, `TokenIssuer`, `HitCounter`, `TxManager`).
- **`internal/adapter`**: Out-of-the-box port implementations:
  - SQLite repositories using `modernc.org/sqlite` (pure Go, CGO-free).
  - Password hashing with Argon2id and JWT issuance.
  - Chi HTTP router (`github.com/go-chi/chi/v5`) with strict DTO deserialization.
  - Thread-safe in-memory hit counter.
- **`internal/infrastructure`**: Database migration runner, seeder, slog JSON logger, configuration loader.

---

## Quick Start & Local Development

### Prerequisites
- **Go**: 1.26 or newer
- **Node.js**: 20+ with npm (for frontend development)
- **Git**

### 1. Clone & Setup Environment
```bash
git clone https://github.com/cmc/labshop.git
cd "labshop"

# Copy environment variables
cp .env.example .env
```

### 2. Generate Static Assets
Generate synthetic test assets (product images, banners, PDF docs, ZIP archives):
```bash
go run ./cmd/genassets/main.go
```

### 3. Run the Backend Server
```bash
# Enable static serving locally without Nginx
go run ./cmd/server/main.go
```
The server will start listening at `http://localhost:8080`.

### 4. Run Frontend (Vite + React)
```bash
cd frontend
npm install
npm run dev
```
Open `http://localhost:5173` in your browser.

---

## Running with Docker Compose

To test the complete stack with Nginx acting as the local reverse proxy and static asset engine:

```bash
# 1. Build and launch containers
docker compose up -d --build

# 2. Check container status
docker compose ps

# 3. View logs
docker compose logs -f
```

The application will be accessible on standard web ports:
- **Web Storefront & Lab Console**: `http://localhost`
- **Health Check**: `http://localhost/api/health`
- **OpenAPI Specification**: `http://localhost/openapi.json`

---

## Automated Smoke Testing

LabShop includes an automated **Smoke Test CLI** (`cmd/smoke`) that verifies 18 critical origin, caching, security, and business flow endpoints sequentially:

```bash
# Run against local server
go run ./cmd/smoke/main.go -base-url=http://localhost

# Or against remote domain
go run ./cmd/smoke/main.go -base-url=https://direct.labshop.vn
```

### Sample Output:
```text
+------+-------------------------------------+---------------------------------+--------+------------+--------+
| Step | Test Name                           | Endpoint                        | Status | Duration   | Result |
+------+-------------------------------------+---------------------------------+--------+------------+--------+
| 1    | Health check                        | /api/health                     | 200    | 1.52ms     | PASS   |
| 2    | Time (no Cache-Control)             | /api/time                       | 200    | 0.81ms     | PASS   |
| 3    | Cache public header                 | /api/cache/public               | 200    | 0.68ms     | PASS   |
| 4    | Cache ETag + 304                    | /api/cache/etag                 | 304    | 1.10ms     | PASS   |
| 5    | Headers redact secret               | /api/headers                    | 200    | 0.56ms     | PASS   |
| 6    | List categories (5)                 | /api/categories                 | 200    | 0.53ms     | PASS   |
| 7    | List products                       | /api/products                   | 200    | 1.05ms     | PASS   |
| 8    | Tutorial scenarios                  | /api/tutorial/scenarios         | 200    | 0.52ms     | PASS   |
| 9    | Customer register & login           | /api/auth/login                 | 200    | 46.35ms    | PASS   |
| 10   | Get auth me profile                 | /api/auth/me                    | 200    | 0.64ms     | PASS   |
| 11   | Cart checkout (idempotent)          | /api/orders                     | 201    | 6.34ms     | PASS   |
| 12   | Pay order mock gateway              | /api/orders/:id/pay             | 200    | 104.55ms   | PASS   |
| 13   | Admin stats                         | /api/admin/stats                | 200    | 21.80ms    | PASS   |
| 14   | OpenAPI spec JSON                   | /openapi.json                   | 200    | 0.58ms     | PASS   |
| 15   | Robots.txt                          | /robots.txt                     | 200    | 0.30ms     | PASS   |
| 16   | Sitemap.xml                         | /sitemap.xml                    | 200    | 0.53ms     | PASS   |
| 17   | Static sample PDF                   | /static/docs/sample.pdf         | 200    | 12.41ms    | PASS   |
| 18   | Contents sample PDF                 | /contents/docs/sample.pdf       | 200    | 1.00ms     | PASS   |
+------+-------------------------------------+---------------------------------+--------+------------+--------+
Summary: Total: 18 | Passed: 18 | Failed: 0 | Total Duration: 201.47ms
[SUCCESS] All 18 smoke tests passed successfully.
```

---

## EC2 Production Deployment

### 1. Provision EC2 Instance
- **Instance Type**: `t3.small` or `t3.medium` (Ubuntu 24.04 LTS or Amazon Linux 2023)
- **Security Group Inbound Rules**:
  - `TCP 80` (HTTP) — Open to `0.0.0.0/0` (or restricted to CMC WAAP & Multi-CDN POP IP ranges)
  - `TCP 443` (HTTPS) — Open to `0.0.0.0/0`
  - `TCP 22` (SSH) — Restricted to administrator IP

### 2. Deploy via Script
```bash
# On the EC2 host:
sudo apt-get update && sudo apt-get install -y docker.io docker-compose-v2 git

git clone https://github.com/cmc/labshop.git /opt/labshop
cd /opt/labshop
cp .env.example .env

# Run deployment script (builds containers, verifies health, runs smoke tests)
chmod +x scripts/deploy.sh
./scripts/deploy.sh
```

---

## Environment Variables

| Variable | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `APP_ENV` | string | `development` | Environment mode (`development`, `staging`, `production`). |
| `HTTP_ADDR` | string | `:8080` | Bind address for Go HTTP server. |
| `DB_PATH` | string | `/data/labshop.db` | Absolute path to SQLite database file. |
| `JWT_SECRET` | string | *(min 32 chars)* | Secret key for signing and validating JWT tokens. |
| `ADMIN_TOKEN` | string | *(secret)* | Static token required for `/api/stats/reset` (`X-Admin-Token`). |
| `SEED` | boolean | `true` | When true, auto-seeds default categories, products, coupons, and users. |
| `SEED_ADMIN_EMAIL` | string | `admin@labshop.local` | Default administrator account email. |
| `SEED_ADMIN_PASSWORD`| string | `Admin@LabShop2026!` | Default administrator account password. |
| `SEED_CUSTOMER_EMAIL`| string | `customer@labshop.local` | Default customer account email. |
| `SEED_CUSTOMER_PASSWORD`| string | `Customer@LabShop2026!` | Default customer account password. |
| `TUTORIAL_SHOW_DEMO_CREDS`| boolean | `true` | Displays demo login credentials on login & tutorial pages. |
| `MAX_UPLOAD_MB` | int | `10` | Maximum file upload size in MB for `/api/upload`. |
| `MAX_SLOW_MS` | int | `60000` | Upper bound for artificial delay on `/api/slow?ms=`. |
| `MAX_PAYLOAD_KB` | int | `5120` | Upper bound for synthetic response size on `/api/payload?kb=`. |
| `CATALOG_CACHE_SECONDS` | int | `60` | Origin `max-age` Cache-Control header for catalog & scenarios. |
| `ENABLE_DOCS` | boolean | `true` | Enables OpenAPI JSON endpoint and Swagger exploration. |
| `ENABLE_WS` | boolean | `true` | Enables WebSocket echo test at `/ws/echo`. |
| `CORS_ALLOWED_ORIGINS`| string | `*` | Allowed CORS origins. |
| `SERVE_STATIC` | boolean | `false` | When true, Go directly serves static files (used in local dev without Nginx). |
| `STATIC_DIR` | string | `/srv/static` | Filesystem directory containing static assets. |
| `STORE_UPLOADS` | boolean | `false` | Persists uploaded files to disk when true. |
| `MOCK_PAYMENT_FAIL_RATE` | float | `0.0` | Probability of simulated payment failure (0.0 to 1.0). |
| `INSTANCE_ID` | string | `labshop-origin-01` | Origin instance identifier stamped in `X-Origin-Instance`. |
| `LOG_LEVEL` | string | `info` | Logging level (`debug`, `info`, `warn`, `error`). |

---

## Lab Console & Interactive Tutorial

The frontend contains an integrated **Lab Console**:
- **Hot-key Toggle**: Press `Ctrl + Shift + L` to open the bottom drawer.
- **Request Inspection**: Inspect raw HTTP requests, response headers (`Age`, `X-Cache`, `CF-Cache-Status`), origin IP, and latency.
- **Scenario Runner**: Run automated multi-CDN and WAAP lab scenarios (Burst tests, cache key permutation, bypass checks).
- **Curl Generator**: Instantly generate cURL commands with custom headers and proxies.

---

## API Documentation

- **OpenAPI 3.0.3 Specification**: [`api/openapi.yaml`](./api/openapi.yaml)
- **JSON Specification**: [`api/openapi.json`](./api/openapi.json)
- **Contract Tests**: Verified via [`test/contract/openapi_test.go`](./test/contract/openapi_test.go) with `kin-openapi`.
- **Docs Page**: Built-in visual API explorer at `/docs`.
