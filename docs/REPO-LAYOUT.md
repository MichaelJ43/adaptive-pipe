# Repository layout

Conventions for source, tests, containers, and deployment artifacts. Aligned with `project_prompt.txt` (git-enabled project, `./src` and `./test`, Docker Compose with a path to Kubernetes).

## Top level

```
adaptive-pipe/
  src/                 # All application source code
  test/                # All automated tests (unit, integration, component, UI)
  docs/                # Planning and technical documentation
  deploy/              # Optional: k8s/helm or compose overrides (recommended)
  docker-compose.yml   # Primary local/stack orchestration (Phase 2)
  README.md            # Quick start and pointers to docs/
  .gitignore
  project_prompt.txt   # Original product requirements
  CALLOUTS.md          # Short clarifications and optional follow-ups
```

Supporting files that are not source (lint config, CI workflow definitions, `Makefile`, etc.) live at the **root** or under **named directories** (for example `.github/workflows/`, `deploy/`).

## `src/` (application code)

Organize by **deployable** or **library** boundaries, for example:

```
src/
  orchestrator/        # HTTP API, queue producers/consumers, persistence
  validate/
  file/
  ui/                  # React/Vite app (or separate folder if built as its own image)
  pkg/                 # Shared Go libraries (optional)
  worker/              # Shared worker agent or stubs (optional)
```

Exact package names are chosen in Phase 2; the rule is **no application logic outside `src/`** except generated code if you add a `gen/` or `api/` folder later (document if added).

Worker **Dockerfile**s may live under `src/build/images/java/`, `deploy/docker/`, or similar—pick one convention and keep all language images discoverable from [TECH-STACK.md](TECH-STACK.md).

## `test/` (all tests)

```
test/
  unit/                # Fast, isolated unit tests (mirror packages under src/ or co-locate in src with *_test.go—pick one style per language)
  integration/         # Services + DB + queue in containers
  component/           # Whole-application or multi-service scenarios
  ui/                  # Browser-based tests (Playwright/Cypress) targeting running stack
```

The project prompt asks for **unit**, **UI**, and **component** tests; all live under `test/` **or** follow language idioms (Go tests beside code) with **clear** mapping in the root `README.md` once implementation exists.

## Git

Initialize a git repository at the repo root when starting Phase 2 if not already present:

```bash
git init
```

Use `.gitignore` for build outputs, `node_modules`, secrets, and local Compose overrides.

## Docker Compose (Phase 2)

`docker-compose.yml` at the root should define at minimum:

- UI, Orchestrator, Validate, File, PostgreSQL, Redis (queue).  
- Optional profiles or separate compose files for **worker** scaling.  

Images build from Dockerfiles referenced by each service; version pins for databases and Redis should be explicit.

## Kubernetes (future)

Recommended layout:

```
deploy/
  k8s/
    base/              # Deployments, Services, ConfigMaps
    overlays/          # dev, staging, prod
```

Or Helm chart under `deploy/helm/adaptive-pipe/`. The same container images built for Compose should be deployable to K8s without source changes—only configuration and replica counts differ.

## Documentation

- **Planning / architecture**: `docs/` (this set), including [SECURITY-AND-OPERATIONS.md](SECURITY-AND-OPERATIONS.md).  
- **End-user runbooks**: expand in Phase 4 per [ROADMAP.md](ROADMAP.md).
