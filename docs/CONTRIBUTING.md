# Contributing to Adaptive Pipe

This project is structured for **continued contribution**: clear package boundaries, tests at multiple layers, and documentation that matches behavior.

## Repository layout

| Path | Purpose |
|------|---------|
| [`src/cmd/`](../src/cmd/) | `orchestrator`, `validate`, `filesvc` entrypoints |
| [`src/internal/`](../src/internal/) | Shared domain logic: `db`, `httpapi`, `dispatcher`, `clients`, `auth`, `eta`, `gc` |
| [`src/ui/`](../src/ui/) | React + Vite console; `e2e/` holds Playwright specs |
| [`src/integration/`](../src/integration/) | Go tests with `-tags=integration` (need Compose stack) |
| [`deploy/migrations/`](../deploy/migrations/) | Reference SQL (embedded copy under `src/internal/db/migrations/`) |
| [`docs/`](../docs/) | Architecture and contracts |
| [`test/component/`](../test/component/) | Shell smoke against a running orchestrator |

Application code lives under **`src/`** per product guidelines; automated tests also use **`test/`** for non-Go artifacts.

## Local development

1. Install **Go 1.22+** and **Node 20+** (or rely on Docker only).
2. Copy environment patterns from `docker-compose.yml` for `DATABASE_URL`, `REDIS_URL`, `VALIDATE_URL`, `FILE_URL`, `JWT_SECRET`.
3. Run `go test ./...` from `src/` before pushing.
4. For UI: `cd src/ui && npm install && npm run dev` with orchestrator on `:8080`.

## Pull request checklist

- **Tests**: add or update unit tests near changed packages; run integration + Playwright when touching dispatch, HTTP, or tenancy boundaries.
- **Tenancy**: every query and queue message must carry `tenant_id` where applicable—never trust org/repo from the client without binding to the authenticated tenant.
- **Docs**: update [`docs/DATA-AND-API.md`](DATA-AND-API.md) or [`docs/ARCHITECTURE.md`](ARCHITECTURE.md) when behavior or contracts change.
- **Migrations**: append a new numbered SQL file under `deploy/migrations/` and mirror into `src/internal/db/migrations/` (or generate from a single source in a follow-up).

## Adding a build language or cloud target

1. Add a worker Docker image (or extend `dispatcher` stage handlers) per [docs/TECH-STACK.md](TECH-STACK.md).
2. Register the image name and defaults in orchestrator configuration (future: DB-backed registry).
3. Document operator steps in [docs/README.md](README.md) or a runbook when the path is stable.

## Code review values

- Prefer **small, reversible** changes.
- Keep log and artifact **throughput** off the orchestrator hot path for large payloads ([ARCHITECTURE.md](ARCHITECTURE.md)).
- Make retention / GC operations **idempotent** and **scoped per tenant + repo**.
