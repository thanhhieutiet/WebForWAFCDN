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

### Phase 2 – Business Logic (Shop Engine) [PENDING]
- [ ] Domain entities & rules (Money VND, User, Catalog, Cart, Coupon, Order, Review, Support).
- [ ] Usecases & Port definitions (Auth, Catalog, Cart, Order, Review, Support, Admin).
- [ ] SQLite schema migrations runner & embedded SQL.
- [ ] SQLite repositories & TxManager.
- [ ] Mock Payment Gateway with delay/failure simulation.
- [ ] HTTP handlers, DTO validation, presenter error mapping.
- [ ] Database seeder (5 categories, 40 products, 4 coupons, admin/customer demo accounts).
- [ ] Unit tests (domain & usecase >= 80% coverage) & integration tests.

### Phase 3 – UI, Tutorial & Lab Console [PENDING]
- [ ] OpenApi client generation (`npm run gen:api`).
- [ ] Design tokens (Slate, Teal, semantic lab colors, self-hosted Inter & JetBrains Mono).
- [ ] Single fetch wrapper (detect origin vs edge block, cache status, age, encoding).
- [ ] Lab Console (`Ctrl+Shift+L` drawer, request history, cURL copy, JSON export).
- [ ] Storefront UI (Home, Catalog, Product details, Cart, Checkout with Idempotency-Key, Orders, Support).
- [ ] Admin dashboard UI (KPI cards, revenue chart lazy Recharts, order management).
- [ ] Tutorial UI (`/tutorial`, scenarios list, P0 scenarios C1-C10, C16-C18, W1-W8, Runner panel: single, N runs, burst histogram).

### Phase 4 – Hardening & Documentation [PENDING]
- [ ] Full OpenAPI 3.0.3 spec (`api/openapi.yaml`) & contract tests (`kin-openapi`).
- [ ] Payload hygiene test suite (running `probes.json` across all input vectors).
- [ ] Smoke test tool (`cmd/smoke`).
- [ ] Documentation: `README.md`, `docs/ARCHITECTURE.md`, `docs/TEST_PLAN.md`.

### Phase 5 – Advanced & Production Readiness [PENDING]
- [ ] HTTPS in Nginx (cert volume mounting).
- [ ] P1 scenarios (Geo access, SSL origin, Image optimization, etc.).
- [ ] GitHub Actions CI workflow.
