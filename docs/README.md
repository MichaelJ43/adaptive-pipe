# adaptive-pipe documentation

Planning and technical reference for **adaptive-pipe**, an all-encompassing, containerized build pipeline that coordinates validation, source caching, builds, tests, and deployments. The system is designed to run under Docker Compose first and to be adaptable to Kubernetes with autoscaling for worker-style nodes.

## Product summary

**What it does**

- Accepts pipeline runs (for example from GitHub CI or the web UI) keyed by organization, repository, and commit.
- Moves each run through ordered stages: validate, file fetch/cache, build, test, deploy—with state persisted for history, logs, and dashboards.
- Keeps ephemeral compute (build, test, deploy) separate from long-lived control-plane services (orchestrator, validate, file nodes, UI, database).
- Targets extensible support for multiple programming languages, IaC tools, and cloud providers via configuration and pluggable worker images.

**Who uses it**

- **Developers and release engineers** monitoring runs and stages, kicking off builds, and managing credentials (admin).
- **External automation** (for example GitHub Actions) calling the orchestrator APIs and receiving status back.

## Document index

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | Services, communication rules, lifecycle, kickoff flow, Kubernetes notes |
| [TECH-STACK.md](TECH-STACK.md) | Languages, runtimes, queue, database, APIs |
| [DATA-AND-API.md](DATA-AND-API.md) | Data model sketch, retention, orchestrator and GitHub-oriented contracts |
| [REPO-LAYOUT.md](REPO-LAYOUT.md) | Source and test layout, Compose, future K8s layout |
| [ROADMAP.md](ROADMAP.md) | Phases 2–5 milestones and MVP vs deferred scope |

Project-wide prompts and callouts live at the repository root: `project_prompt.txt` and [CALLOUTS.md](../CALLOUTS.md).

## Phase 1 status

Phase 1 (planning) is satisfied by the documents above. Implementation, testing, user guides, and release automation follow in Phases 2–5 per [ROADMAP.md](ROADMAP.md).
