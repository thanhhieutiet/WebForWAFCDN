# LabShop Task Board & Multi-Agent Tracking

## Roles
- **Manager Agent**: Overall orchestration, code review, acceptance criteria validation, cross-phase consistency.
- **Backend Agent (`backend_dev`)**: Go Clean Architecture, domain models, usecases & ports, SQLite repositories, Lab & Business handlers, genassets, OpenAPI spec.
- **Frontend Agent (`frontend_dev`)**: Vite + React + TS, Tailwind CSS, shadcn/ui components, generated OpenAPI client, React Router, Lab Console (`Ctrl+Shift+L`), scenario runners, UI/UX tokens.
- **QA & Troubleshooter Agent (`qa_troubleshoot`)**: Architecture tests (`archtest`), payload hygiene tests, `go test -race`, linter (`golangci-lint`), Vitest, Docker build verification, troubleshooting regressions.

---

## Phases & Status

### Phase 0 – Bootstrap [COMPLETED]
- [x] Git init, `.gitattributes` (LF), `.gitignore`, `docs/DECISIONS.md`, `docs/TASK_BOARD.md`.
- [x] Go module initialization (`labshop`), folder skeleton (cmd, internal/domain, internal/usecase, internal/adapter, internal/infrastructure).
- [x] Core config & slog JSON logger.
- [x] Common HTTP middleware (`X-Request-Id`, `X-Origin-Instance`, `X-Origin-Time`, access log, panic recover).
- [x] Architecture test (`internal/archtest`) validating dependency rules.
- [x] `/api/health` endpoint.
- [x] Frontend project scaffolding (`frontend/` with Vite, TS, Tailwind CSS, shadcn/ui base, ESLint, Vitest).
- [x] Dockerfile (multi-stage: build Go, build Vite/node, target api, target nginx).
- [x] `docker-compose.yml`, `nginx/nginx.conf.template`, `.env.example`.
- [x] Scripts: `scripts/dev.ps1`, `scripts/deploy.sh`.
- [x] Verification: `go vet`, `go test -race`, Vite build, `docker compose build`.

### Phase 1 – Lab Endpoints & Static Assets [COMPLETED]
- [x] Port `HitCounter` in `internal/usecase/port` & thread-safe memory adapter in `internal/adapter/counter/memory`.
- [x] Lab usecase (`internal/usecase/lab`) and HTTP handlers (`internal/adapter/http/handler/lab.go`).
- [x] Lab endpoints: `/api/time`, `/api/cache/{mode}`, `/api/headers`, `/api/echo`, `/api/status/{code}`, `/api/slow`, `/api/payload`, `/api/stats`, `/api/upload`, `/ws/echo`, `/robots.txt`, `/sitemap.xml`, initial `/openapi.json`.
- [x] Asset generator: `cmd/genassets/main.go` producing product images (p01-p40), banner-large.jpg, sample pdf/zip/bin files, static css/js.
- [x] Static serving verification in Go when `SERVE_STATIC=true` and in Nginx via `/static/` and `/contents/`.
- [x] Unit tests, race detection, and handler tests for all lab endpoints.

### Phase 2 – Business Logic (Shop Engine) [COMPLETED]
- [x] Domain entities, value objects & business rules (`money`, `user`, `catalog`, `cart`, `coupon`, `order`, `review`, `support`, `domain errors`).
- [x] Usecases & Port definitions with fake repository unit tests (TDD, >=80% coverage).
- [x] SQLite schema migrations runner (embedded SQL) & pure-Go `modernc.org/sqlite` repositories with `TxManager`.
- [x] Argon2id password hasher & JWT token issuer adapter.
- [x] Mock Payment Gateway with configurable delay & failure rate.
- [x] HTTP handlers, strict JSON DTOs (`DisallowUnknownFields`), presenter mapping errors to status codes.
- [x] Database seeder (5 categories, 40 products with images, 4 coupons, demo admin & customer).
- [x] Integration & End-to-End flow verification (Register -> Cart -> Order -> Pay -> Ship/Deliver -> Review -> Ticket).
- [x] Architecture compliance (`internal/archtest`), `go vet`, and `go test -v ./...` 100% passing.

### Phase 3 – UI, Tutorial & Lab Console [COMPLETED]
- [x] Backend: `ScenarioRepository` port, embedded `scenarios.json` (P0 scenarios C1-C10, C16-C18, W1-W8, Shop), usecase & `GET /api/tutorial/scenarios`.
- [x] Frontend single fetch wrapper: intercept all calls, classify Origin vs Edge Block, extract cache HIT/MISS/Age/encoding, stream to Lab Console.
- [x] Lab Console Drawer: fixed at bottom (`Ctrl+Shift+L` toggle), request history (50 items), inspect headers, copy cURL, export JSON.
- [x] Environment Bar: active host & routing label (`direct`, `waap`, `cdn`, `full`).
- [x] Storefront: Home, Catalog (filters, search, sort, pagination), Product details (quantity stepper, reviews), Cart (optimistic updates, coupon), Checkout (`Idempotency-Key`), Orders (status timeline), Support (ticket thread).
- [x] Admin Dashboard: KPI cards, lazy-loaded Recharts revenue graph, order status transitions, tickets management.
- [x] Tutorial & Runner: 4-route network diagram, scenarios list, scenario steps, Runner panel (Run once, Run N with repeated ID color highlighting, Burst histogram, cURL generator).
- [x] Docs page: self-rendered endpoint explorer from `/openapi.json`.
- [x] Build & Docker update: `docker compose up -d --build`.

### Phase 4 – Hardening & Documentation [COMPLETED]
- [x] Full OpenAPI 3.0.3 spec (`api/openapi.yaml`) & contract tests (`kin-openapi`).
- [x] Payload hygiene test suite (running `probes.json` across all input vectors).
- [x] Smoke test tool (`cmd/smoke/main.go`).
- [x] Documentation: `README.md`, `docs/ARCHITECTURE.md`, `docs/TEST_PLAN.md`.

### Phase 5 – Advanced & Production Readiness [COMPLETED]
- [x] Nginx HTTPS port 443 configuration with Let's Encrypt certificate mounting (`/etc/letsencrypt`) and fallback self-signed cert.
- [x] Multi-environment Docker Compose volume mount for host SSL certificates.
- [x] Extension P1 scenarios (W9 Geo access, W10 IP list, W11 Body limit, W12 SSL origin, W13 Load balancer, C11 Image opt, C12 Range request, C13 HTTP/2, C14 CDN firewall, C15 Multi-CDN routing).
- [x] GitHub Actions CI workflow (`.github/workflows/ci.yml`) testing vet, archtest, race detector, frontend build, and docker compose build.
- [x] EC2 production deployment guide & checklist in `docs/DEPLOY_EC2.md`.

