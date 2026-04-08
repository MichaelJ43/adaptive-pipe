# adaptive-pipe documentation

Planning and technical reference for **adaptive-pipe**, an all-encompassing, containerized build pipeline that coordinates validation, source caching, builds, tests, and deployments. The system is designed to run under Docker Compose first and to be adaptable to Kubernetes with autoscaling for worker-style nodes.

## Product summary

**What it does**

- **Multi-tenant SaaS**: many customers on one installation, isolated by **`tenant_id`**. Accepts pipeline runs from **GitHub.com** (commit webhook **MVP**) or the web UI, keyed by tenant, organization, repository, and commit.
- Moves each run through ordered stages: validate, file fetch/cache, build, test, deploy—with state persisted for history, logs, and dashboards.
- Uses **Redis** for queues, **PostgreSQL** (containerized **MVP**) for authoritative state, and **REST/JSON** between services.
- **MVP deploy path**: **AWS** with **Terraform**; **Helm, Ansible, GCP, and Azure** are future extensions via config and worker images.
- **Six build languages**: Java, Python, TypeScript/JavaScript, Rust, Go, C#.
- **Warm pools** for Build/Test/Deploy are configurable from **platform settings** in the UI; **scale-from-zero** applies when pool size is zero.

**Who uses it**

- **Developers and release engineers** monitoring runs and stages, kicking off builds, tuning warm pools, and managing credentials (admin).
- **GitHub** automation receiving run status via an implementation-chosen Checks/Status API (documented at release).

## Document index

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | Services, comms, warm pools, sticky affinity, Compose/K8s |
| [TECH-STACK.md](TECH-STACK.md) | Languages, Redis, Postgres MVP, REST, AWS+Terraform MVP |
| [DATA-AND-API.md](DATA-AND-API.md) | Data model, retention (10 per org/repo), kickoff 2xx/200, GitHub |
| [SECURITY-AND-OPERATIONS.md](SECURITY-AND-OPERATIONS.md) | Threat model, logging, idempotency, rate limits, flags, backups |
| [REPO-LAYOUT.md](REPO-LAYOUT.md) | Source and test layout, Compose, future K8s layout |
| [ROADMAP.md](ROADMAP.md) | Phases 2–5 milestones and MVP vs deferred scope |

Repository root: `project_prompt.txt` (authoritative product text) and [CALLOUTS.md](../CALLOUTS.md) (short clarifications and optional follow-ups).

## Phase 1 status

Phase 1 (planning) is satisfied by the documents above. Implementation follows Phases 2–5 per [ROADMAP.md](ROADMAP.md).
