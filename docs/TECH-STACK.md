# Technology stack

Primary selection criterion: **runtime and operational speed** for the control plane (fast cold start, efficient concurrency, small container images) while keeping **worker isolation** for heavy builds.

## Control plane services

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Orchestrator, Validate, File (server-side) | **Go** | Strong concurrency, fast startup, compact static binaries, good fit for queue-driven HTTP/gRPC services |
| UI | **TypeScript** with **React** and **Vite** | Rich ecosystem for dashboards, component testing, and tooling |

All application source for these components lives under `./src` per [REPO-LAYOUT.md](REPO-LAYOUT.md).

## Ephemeral workers (Build, Test, Deploy)

Workers are **not** required to share the same language as the control plane:

- **Build**: one **Docker image per language/toolchain** (Java, Python, Node/TS, Rust, plus two additional languages—see [CALLOUTS.md](../CALLOUTS.md)). The Orchestrator schedules a container (Compose or Kubernetes Job) with the correct image and passes run metadata and File node endpoints (or tokens).
- **Test**: images bundling browsers, k6, or contract-test tools as needed; same scheduling model.
- **Deploy**: images with Terraform, Helm, Ansible CLIs and cloud SDKs as needed; scope grows post-MVP.

A **thin Go agent** inside each worker image (or a shared sidecar pattern) can standardize registration, heartbeats, and log shipping to the Orchestrator; alternatively workers speak a minimal HTTP API. The implementation picks one pattern in Phase 2 and applies it consistently.

## Six programming languages (build targets)

Baseline set aligned with the product prompt:

1. Java  
2. Python  
3. TypeScript / JavaScript  
4. Rust  
5. **TBD** (candidates: Go, C#)  
6. **TBD**  

Final selection is recorded in [CALLOUTS.md](../CALLOUTS.md) once decided.

## Queue

**Recommended default: Redis with Redis Streams** (or Redis as a queue backend with consumer groups).

- Low latency, simple ops in Docker Compose, straightforward Helm charts on Kubernetes.  
- Fits “enqueue next stage when previous completes” and multiple worker types (stream keys or prefixes per stage).  

**Alternative: NATS JetStream** if you prefer log-style persistence and stronger streaming semantics across clusters. Either choice should be abstracted behind a small internal interface so swapping is possible.

## Database

**PostgreSQL** for relational data: runs, stages, users/roles, audit, and references to blobs stored on File nodes or object storage. **Credential secrets** are encrypted **in the application** before insert; the DB stores ciphertext and metadata only.

## APIs

| Boundary | Style |
|----------|--------|
| Browser to Orchestrator (UI) | **REST** + JSON; optional later: WebSocket/SSE for live stage updates |
| GitHub / external CI to Orchestrator | **REST** + JSON (webhook or explicit kickoff—see [DATA-AND-API.md](DATA-AND-API.md)) |
| Internal service to service | **gRPC** (recommended for Orchestrator ↔ Validate/File agents) **or** REST; choose one for v1 and document generated types |

Public external docs should describe stable REST endpoints; internal protobuf definitions can live under `./src` when gRPC is adopted.

## IaC and clouds (extensibility)

Configuration-driven lists for:

- **IaC**: Terraform, Helm, Ansible (MVP may implement one; see [ROADMAP.md](ROADMAP.md)).  
- **Clouds**: AWS, GCP, Azure (credentials and targets supplied per environment).

New providers add **worker image variants** and **config schema** rather than forks of the Orchestrator.

## Testing technologies (Phases 2–3)

- **Unit**: language-native frameworks (Go `testing`, Vitest/Jest for UI).  
- **Component / E2E**: Playwright or Cypress against the running stack; API contract tests against the Orchestrator.  

Details belong in the test layout under `./test` (see [REPO-LAYOUT.md](REPO-LAYOUT.md)).
