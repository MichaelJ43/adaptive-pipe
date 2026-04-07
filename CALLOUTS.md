# Callouts: additions, removals, modifications, and open questions

Review notes for adaptive-pipe planning. Resolve items here before or during Phase 2 to avoid rework.

---

## Suggested additions

- **Threat model (short)** — Document who can call kickoff APIs, how GitHub tokens are scoped, and where secrets never appear (logs, traces, error payloads).  
- **Observability** — Structured logs with `run_id`, `build_number`, `stage`; optional OpenTelemetry trace id propagated from Orchestrator through workers.  
- **Idempotency** — Webhook and kickoff endpoints should accept an idempotency key or dedupe on `(org, repo, commit, delivery_id)` to survive GitHub retries.  
- **Rate limiting** — Per-IP and per-token limits on public-facing Orchestrator routes.  
- **Backup and restore** — For Postgres and File node storage: RPO/RTO expectations once you leave dev-only mode.  
- **Feature flags** — Gate “cloud prereq validation” and “auto IaC detection” so MVP stays shippable.

---

## Suggested removals or deferrals

- **Vague “etc.” on IaC** — [project_prompt.txt](project_prompt.txt) lists Terraform, Helm, Ansible, “etc.” Defer unspecified tools until there is demand; MVP: **one** (for example Terraform only).  
- **Full multi-cloud intelligence early** — AWS, GCP, Azure support is a roadmap goal; defer automatic provider inference to post-MVP (see [docs/ROADMAP.md](docs/ROADMAP.md)).  
- **Ambitious performance testing in first Test node iteration** — Start with smoke-level checks; add k6 and full perf gates later.

---

## Suggested modifications to requirements or docs

- **Kickoff HTTP status** — Prompt says **200** when kickoff is initiated; [DATA-AND-API.md](docs/DATA-AND-API.md) follows that. Some APIs use **202** for async acceptance; if you standardize on 202 later, document it as an API version change.  
- **“Always running” vs scale-to-zero for Build/Test** — Prompt says ephemeral and “begin spinning up immediately.” Clarify whether you want a **warm pool** of idle workers for latency vs strict **scale-from-zero** (KEDA-friendly). The architecture supports both; ops defaults should be explicit.  
- **Database “only non-multi-node”** — For production, consider **managed Postgres** (single endpoint) or **HA** (Patroni, cloud RDS Multi-AZ) without changing the logical “one database” rule. Update docs when you pick HA.  
- **“10 builds” retention** — Confirm whether the window is **last 10 runs by build number regardless of status** (recommended for retries) vs **last 10 successful only**. Also confirm whether **DB rows** for older runs are kept indefinitely while **files** are GC’d.

---

## Open questions

1. **Languages 5 and 6** — Java, Python, TS/JS, Rust are fixed in the prompt. Which two additional languages: **Go + C#**, **Go + PHP**, **C# + Kotlin**, or others?  
2. **Authentication** — Local username/password only for v1, or **OIDC** (Entra ID, Google, GitHub OAuth) from day one?  
3. **Tenancy** — Single-tenant installation per deployment vs multi-tenant SaaS (affects DB schema and credential scoping).  
4. **GitHub scope** — GitHub.com only at MVP, or **GitHub Enterprise Server** from the start (different webhook and API base URLs)?  
5. **GitHub status API** — **Commit Statuses** vs **Checks API** for reporting back to GitHub CI?  
6. **File node storage** — Local volume only, or **S3/GCS/Azure Blob** as backing store behind the File service for K8s durability?  
7. **Internal RPC** — Confirm **gRPC** vs **REST** for Orchestrator ↔ Validate/File for v1 (see [docs/TECH-STACK.md](docs/TECH-STACK.md)).  
8. **Queue** — Confirm **Redis Streams** vs **NATS JetStream** for production preference.  
9. **Deploy MVP** — Smallest first target: one cloud (which?) and one mechanism (Terraform apply in isolated workspace, Helm upgrade, etc.)?  
10. **ETA algorithm** — Simple moving average of last *n* successful stage durations vs exponential weighting; minimum sample size before showing an estimate?

---

## Items to sync back into `project_prompt.txt` (optional)

If you adopt any modification above (for example 202 vs 200, or retention definition), consider updating `project_prompt.txt` so it stays the single written source of product intent—or add a short “Decisions” subsection pointing at `docs/DATA-AND-API.md` and this file.
