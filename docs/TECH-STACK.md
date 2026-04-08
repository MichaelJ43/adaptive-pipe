# Technology stack

Primary selection criterion: **runtime and operational speed** for the control plane (fast cold start, efficient concurrency, small container images) while keeping **worker isolation** for heavy builds.

## Control plane services

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Orchestrator, Validate, File (server-side) | **Go** | Strong concurrency, fast startup, compact static binaries, good fit for queue-driven REST services |
| UI | **TypeScript** with **React** and **Vite** | Rich ecosystem for dashboards, platform settings (warm pools), component testing, and tooling |

All application source for these components lives under `./src` per [REPO-LAYOUT.md](REPO-LAYOUT.md).

## Ephemeral workers (Build, Test, Deploy)

Workers are **not** required to share the same language as the control plane:

- **Build**: one **Docker image per language/toolchain** for **Java, Python, TypeScript/JavaScript, Rust, Go, and C#**. The Orchestrator schedules a container (Compose or Kubernetes Job) with the correct image and passes run metadata and File node endpoints (or tokens).
- **Test**: images bundling browsers, contract-test tools, etc.; same scheduling model. Performance-heavy testing (for example k6) can start as smoke-level and expand later.
- **Deploy**: **MVP** image includes **Terraform** and AWS tooling. **Helm, Ansible, GCP, and Azure** are **future goals**—add worker variants and config schema rather than forking the Orchestrator (see [ROADMAP.md](ROADMAP.md)).

A **thin Go agent** inside each worker image (or a shared sidecar pattern) can standardize registration, heartbeats, and log shipping to the Orchestrator; alternatively workers speak a minimal REST API. The implementation picks one pattern in Phase 2 and applies it consistently.

**Warm pool and scale-from-zero**: Long-lived **control plane** stays up. Build/Test/Deploy use a **configurable warm pool per tenant** (counts in **platform settings** in the UI for that customer) of idle workers; when pool size is zero, workers **scale from zero** on demand. See [ARCHITECTURE.md](ARCHITECTURE.md).

## Six programming languages (build targets)

1. Java  
2. Python  
3. TypeScript / JavaScript  
4. Rust  
5. Go  
6. C#  

## Queue

**Redis** (Redis Streams or consumer groups) is the chosen queue for MVP.

- Low latency, simple operations in Docker Compose, straightforward charts on Kubernetes.  
- Fits “enqueue next stage when previous completes” and multiple worker types (stream keys or prefixes per stage).  

Abstract queue access behind a small internal interface if you ever need a second implementation.

## Database

**PostgreSQL** for relational data: runs, stages, users/roles, audit, and references to blobs on the File node.

- **MVP**: Run as a **small containerized** instance in Compose (and equivalent in dev K8s).  
- **Future**: Swap connection settings (or operators) for **external / HA / managed** Postgres without changing domain models—document migration in Phase 4 runbooks.

**Credential secrets** are encrypted **in the application** before insert; the DB stores ciphertext and metadata only.

## APIs

| Boundary | Style |
|----------|--------|
| Browser to Orchestrator (UI) | **REST** + JSON; optional later: WebSocket/SSE for live stage updates |
| GitHub / external CI to Orchestrator | **REST** + JSON (webhook on commit for MVP—see [DATA-AND-API.md](DATA-AND-API.md)) |
| Internal service to service | **REST** + JSON (Orchestrator ↔ Validate, File, workers) |

## Authentication

- **MVP**: **Local username and password** (hashed passwords, session or token as implemented in Phase 2).  
- **Future**: OIDC/SSO and other methods should plug in without rewriting core run/credential flows.

## IaC and clouds (MVP vs future)

| Area | MVP | Future goals (extension-friendly) |
|------|-----|-----------------------------------|
| Cloud | **AWS** | GCP, Azure |
| IaC | **Terraform** | Helm, Ansible, additional tooling |
| Deploy “intelligence” | Explicit config per environment | Optional auto-detection (behind feature flags per [SECURITY-AND-OPERATIONS.md](SECURITY-AND-OPERATIONS.md)) |

New targets add **worker image variants** and **configuration**, not forks of the Orchestrator.

## File storage

- **MVP**: **Local volumes** attached to File service containers.  
- **Future**: S3, GCS, or Azure Blob behind the same File API abstraction.

## Stage ETA (UI)

- **MVP**: **Simple moving average** of recent stage durations (per repo or global—implementation choice).  
- Keep the estimator behind a **small pure function or interface** so it can be swapped (exponential decay, minimum sample size, etc.) without UI churn.

## Testing technologies (Phases 2–3)

- **Unit**: language-native frameworks (Go `testing`, Vitest/Jest for UI).  
- **Component / E2E**: Playwright or Cypress against the running stack; API contract tests against the Orchestrator.  

Details belong in the test layout under `./test` (see [REPO-LAYOUT.md](REPO-LAYOUT.md)).

## Related documents

- [SECURITY-AND-OPERATIONS.md](SECURITY-AND-OPERATIONS.md) — threat model, logging, idempotency, rate limits, flags, backups.
