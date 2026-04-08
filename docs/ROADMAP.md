# Roadmap: Phases 2–5

Maps the process in `project_prompt.txt` to concrete outcomes. Settled MVP vs future scope is defined there and in [TECH-STACK.md](TECH-STACK.md); optional notes remain in [CALLOUTS.md](../CALLOUTS.md).

## Phase 2: Implementation with TDD

**Goal:** Minimal vertical slice: Compose stack, Redis + container Postgres, Orchestrator as REST hub, **AWS + Terraform** deploy path stub or happy path, **github.com** webhook, local File volumes.

Suggested milestones:

1. **Repository bootstrap** — `src/`, `test/`, root `docker-compose.yml`, PostgreSQL + Redis, `.gitignore`, root `README.md` stub.  
2. **Orchestrator core (TDD first)** — domain types for `run` and `stage`, kickoff API returning **200 OK** (or other agreed 2xx) with `run_id` / `build_number`, enqueue first stage, persist to DB, **idempotent** webhook handling.  
3. **Validate + File stubs** — REST calls from Orchestrator; File uses **local volume** MVP.  
4. **Build** — at least **one** language image end-to-end (expand toward all six incrementally).  
5. **Test + Deploy workers** — Test worker stub; Deploy worker with **Terraform/AWS** MVP.  
6. **UI skeleton** — org/repo lists, build detail with stages, **simple moving average** ETAs (pluggable function).  
7. **Platform settings** — persist **warm pool** sizes for Build/Test/Deploy; Orchestrator respects targets (best-effort in Compose).  
8. **Auth** — **local username/password** MVP with **tenant membership**; hooks for future OIDC/SSO. **Tenant admin** gate for credentials. **Tenant isolation** tests (no cross-tenant reads) from day one.

**Explicitly easy to extend later:** Helm and Ansible worker images, GCP/Azure credentials, GitHub Enterprise base URL, object storage for File, external Postgres—without redesigning the orchestration model.

## Phase 3: Verification and test pyramid

**Goal:** Unit, integration, component, and UI tests as required by the product prompt.

- **Unit** and **integration** (Orchestrator + DB + Redis + workers).  
- **Component**: full Compose stack, end-to-end run from webhook or API.  
- **UI**: Playwright/Cypress; cover warm pool settings, kickoff, skipped stages.

Add coverage for retention/GC rules (**10 builds per org/repo**, any status) and sticky File affinity failures.

## Phase 4: Supplemental documentation

**Goal:** Operator and developer “how to run and extend.”

- GitHub App/webhook setup (**github.com** MVP; note future on-prem).  
- AWS/Terraform backend and credential model.  
- How to add a **language** image or a **future** cloud/IaC target.  
- Moving Postgres and File storage to **production-grade** external services.  
- Troubleshooting: Redis backlog, warm pool tuning, webhook duplicates (idempotency).

## Phase 5: Release pipeline

**Goal:** Versioning, build, test, and publish artifacts for multiple platforms.

- Semantic versioning (tags + changelog).  
- CI: lint, unit, integration, image build, optional registry push.  
- Document architectures (`linux/amd64`, `linux/arm64` as needed).  
- Optional: signed images, SBOM—if your organization requires them.

## MVP vs deferred

| MVP (Phase 2–3) | Deferred (designed as extensions) |
|-----------------|-----------------------------------|
| AWS + Terraform deploy | GCP, Azure; Helm, Ansible |
| github.com + commit webhook | GitHub Enterprise / on-prem API URL |
| Container Postgres + local File volumes | Managed/external DB; S3/GCS/Azure Blob |
| Local username/password | OIDC, SSO, additional identity providers |
| Redis queue | Same; interface allows swap if ever needed |
| Simple moving average ETAs | Richer estimators; ML (if ever) |
| Explicit config per environment | Automatic IaC/cloud detection (feature-flagged) |
| Tenant-scoped Postgres + Redis keys/streams | Dedicated DB per enterprise customer (only if a future tier requires it) |

Performance testing in the Test node: start with **smoke** API/UI checks; add **k6** or full perf gates when prioritized.
