# Repository layout

Conventions for source, tests, containers, and deployment artifacts.

## Top level

```
adaptive-pipe/
  src/                    # Go services + React UI (see below)
  test/                   # Cross-cutting non-Go tests (e.g. component shell scripts)
  deploy/                 # nginx config, reference migrations
  docs/                   # Architecture and contribution guides
  .github/workflows/      # CI (build + test only; no deploy)
  docker-compose.yml      # Full local stack
  Dockerfile.*            # Per-service images
  Makefile                # Convenience targets
  README.md               # SaaS-oriented quick start (live stack)
  project_prompt.txt
  CALLOUTS.md
```

## `src/` (application code)

```
src/
  cmd/orchestrator/       # API + embedded dispatcher + DB seed
  cmd/validate/           # Shift-left stub HTTP service
  cmd/filesvc/            # File/workspace stub (local volume MVP)
  internal/
    auth/                 # JWT
    clients/              # Validate + File HTTP clients
    config/
    db/                   # pgx store, migrations (embedded SQL)
    dispatcher/           # Redis BRPOP worker loop
    eta/                  # Moving average ETAs
    gc/                   # Retention selection helpers
    httpapi/              # Chi router + handlers
  integration/            # go test -tags=integration
  ui/                     # Vite + React + Playwright e2e
```

## `test/` (other tests)

```
test/
  component/              # curl smoke (expects running orchestrator)
```

Go unit tests live next to packages under `src/internal/...`.

## Docker Compose (Phase 2+)

Publishes **5432, 6379, 8080–8082, 3000** for local dev and CI host-based integration tests.

## Kubernetes (future)

See [ARCHITECTURE.md](ARCHITECTURE.md); suggested `deploy/k8s/` when you add manifests.

## Documentation

- **Operators & users**: [README.md](../README.md)
- **Contributors**: [CONTRIBUTING.md](CONTRIBUTING.md)
