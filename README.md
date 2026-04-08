# Adaptive Pipe

**Adaptive Pipe** is a multi-tenant **SaaS-style continuous delivery control plane**: one deployment serves many customers, each with isolated pipelines, credentials, and warm-pool settings. The product is **live** in the sense that you can run the full stack locally or in your cluster today—Compose spins up the orchestrator, validate and file services, workers (embedded in the orchestrator for the MVP slice), PostgreSQL, Redis, and the web console.

This repository implements the architecture described in [`docs/`](docs/): REST between services, Redis job queues, tenant-scoped data, GitHub webhooks (github.com MVP), and an AWS + Terraform shaped deploy stage (stubbed for fast feedback; extend worker images for real applies).

## Who this is for

- **Teams** who want a hosted pipeline control plane with org/repo grouping, sequential stages (validate → file → build → test → deploy), and historical ETAs.
- **Contributors** extending language images, clouds, or auth—see [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md).

## Quick start (recommended)

Prerequisites: **Docker** and **Docker Compose v2**.

```bash
docker compose up --build
```

Open **http://localhost:3000** for the UI (nginx proxies `/api/` to the orchestrator). The orchestrator API is also on **http://localhost:8080**.

### First sign-in (seeded tenant)

After the first orchestrator boot, a demo tenant and admin user are created if the database is empty:

| Field        | Value        |
|-------------|--------------|
| Tenant slug | `demo`       |
| Username    | `admin`      |
| Password    | `admin123`   |

Override the seed password with `SEED_DEMO_ADMIN_PASSWORD` on the orchestrator (see `docker-compose` environment). Set a strong `JWT_SECRET` for anything beyond local sandboxes.

### GitHub webhook (github.com MVP)

Configure a repository webhook (push or commit-oriented events) to:

`POST http://<your-orchestrator-host>:8080/api/v1/tenants/<tenant-slug>/webhooks/github`

For local Docker with port publish:

`http://localhost:8080/api/v1/tenants/demo/webhooks/github`

GitHub sends `X-GitHub-Delivery`; the orchestrator deduplicates on `(tenant, delivery)` to survive retries. Payload shape follows GitHub `push` events (`after`, `repository.owner.login`, `repository.name`).

### Platform settings (warm pools)

Signed-in **tenant admins** can `PATCH /api/v1/platform/settings` (or use **Platform settings** in the UI) to persist **Build / Test / Deploy** warm pool targets. The embedded dispatcher respects these targets in later iterations; today the values are stored and surfaced for operations consistency with the architecture docs.

### Developer workflow (UI hot reload)

```bash
# Terminal 1 — dependencies only, or full stack without UI container
docker compose up postgres redis validate filesvc orchestrator

# Terminal 2 — UI against local API (Vite proxies /api → :8080)
cd src/ui && npm install && npm run dev
```

Visit http://localhost:5173

## Tests & CI

| Layer        | Command |
|-------------|---------|
| Go unit      | `cd src && go test ./...` (or use the Go 1.22+ toolchain) |
| Go integration | `docker compose up -d --build` then `cd src && INTEGRATION=1 go test -tags=integration ./integration/...` |
| UI E2E     | With stack up: `cd src/ui && npx playwright install && npm run test:e2e` |
| Component curl | `BASE_URL=http://127.0.0.1:8080 sh test/component/smoke.sh` |

GitHub Actions (`.github/workflows/ci.yml`) runs unit tests, UI production build, `docker compose config`, brings the stack up, runs webhook smoke, Go integration tests against published ports, and Playwright.

## Documentation map

- [docs/README.md](docs/README.md) — documentation index and product summary
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — topology, risks (log throughput, sticky File, GC)
- [docs/DATA-AND-API.md](docs/DATA-AND-API.md) — tenancy, retention, API sketch
- [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) — how to extend the codebase safely

## License

Specify your license in a `LICENSE` file when you publish the product publicly.
