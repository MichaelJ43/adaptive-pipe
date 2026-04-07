# Roadmap: Phases 2–5

Maps the process in `project_prompt.txt` to concrete outcomes. Phase 1 (planning) is complete once the `docs/` set and [CALLOUTS.md](../CALLOUTS.md) exist and decisions there are resolved for implementation.

## Phase 2: Implementation with TDD

**Goal:** Minimal vertical slice that proves the architecture: Compose stack, real persistence, one happy-path pipeline, Orchestrator as the hub.

Suggested milestones:

1. **Repository bootstrap** — `src/`, `test/`, root `docker-compose.yml`, PostgreSQL + Redis, `.gitignore`, root `README.md` stub.  
2. **Orchestrator core (TDD first)** — domain types for `run` and `stage`, kickoff API returning **200** with `run_id` / `build_number`, enqueue first stage, persist to DB.  
3. **Validate + File stubs** — Orchestrator calls Validate and File services (can return success with fixed delays initially).  
4. **One Build language** — single Dockerfile (for example Python or Go) scheduled as ephemeral job; reports completion to Orchestrator.  
5. **UI skeleton** — list builds by org/repo, detail page with ordered stages (static or mock ETAs OK).  
6. **Auth placeholder** — login screen and role check hook; wire **admin** gate for credential API before storing real secrets.

Defer: full six languages, all clouds, intelligent IaC detection, performance tests.

## Phase 3: Verification and test pyramid

**Goal:** Confidence via automated tests as requested: unit, component, UI.

- Run **unit** tests in CI and locally.  
- **Integration** tests: Orchestrator + DB + Redis + one worker image.  
- **Component** tests: full Compose stack, API-level or thin UI checks for end-to-end run creation.  
- **UI** tests: Playwright/Cypress against staging stack; cover kickoff flow and stage display (skipped vs active).

Add coverage for GitHub webhook/idempotency and failure paths (Validate fail, Build fail).

## Phase 4: Supplemental documentation

**Goal:** “What it does” and “how to use it” for operators and developers.

- Installation and configuration (environment variables, secrets, GitHub App setup).  
- How to add a **language** image and register it in config.  
- How to add a **cloud** or **IaC** target post-MVP.  
- Troubleshooting (queue backlog, sticky File node loss, credential rotation).

## Phase 5: Release pipeline

**Goal:** Versioning, build, test, and publish artifacts for multiple platforms.

- Semantic versioning (tags + changelog).  
- CI pipeline: lint, unit, integration, image build, optional push to registry.  
- Release artifacts: container images for each service and worker flavors; document supported architectures (for example `linux/amd64`, `linux/arm64`).  
- Optional: signed images, SBOM generation—add if required by your org.

## MVP vs deferred (feature flags)

| MVP (Phase 2–3) | Deferred (post-MVP) |
|-----------------|---------------------|
| One cloud + one IaC path manually configured | “Intelligent” detection of IaC and cloud |
| Commit Statuses **or** Checks (one chosen) | Full GitHub Enterprise parity |
| Single Postgres instance | HA / read replicas |
| ETA from simple moving average | ML or per-repo tuning |

Track deferred items in [CALLOUTS.md](../CALLOUTS.md) until promoted into scope.
