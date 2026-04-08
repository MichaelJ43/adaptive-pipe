# Security and operations

Cross-cutting requirements agreed for adaptive-pipe. Implementation detail lands in Phase 2+; this document is the planning baseline.

## Threat model (short)

- **Multi-tenant SaaS**: Every read and write must be **authorized for exactly one `tenant_id`** derived from the session, token, or webhook routing—not from client-supplied org/repo alone. Prevent **IDOR** (cross-tenant run or credential access) in API handlers, workers, and admin tools.
- **Kickoff and webhooks**: Treat unauthenticated or weakly authenticated ingress as untrusted. Validate payloads, signatures (GitHub webhook secret **per tenant/installation**), and never trust client-supplied file paths or shell fragments.
- **Credentials**: GitHub tokens, cloud keys, and UI-entered secrets must **never** appear in application logs, distributed traces, or error responses returned to non-admin callers. Store ciphertext only; restrict decryption to the narrow services that need it.
- **Admin actions**: Credential create/update/delete and platform settings (warm pool, dangerous flags) require an **admin**-authenticated session.
- **Network**: In Kubernetes, use NetworkPolicies (or equivalent) so workers cannot reach arbitrary internal services—only Orchestrator and File as per [ARCHITECTURE.md](ARCHITECTURE.md).

## Observability

- **Structured logs**: Include `tenant_id`, `run_id`, `build_number`, `stage`, and `github_org` / `github_repo` where applicable. Use a stable JSON field naming convention.
- **Correlation**: Propagate a **request or trace id** from the Orchestrator into worker logs (headers or job env) so a single run can be traced across containers.
- **OpenTelemetry** (optional in MVP, easy to add): trace spans around kickoff, queue publish/consume, and stage boundaries.

## Idempotency

- **Webhooks**: GitHub may retry deliveries. Support deduplication via delivery id, idempotency key header, or a natural key such as `(tenant_id, org, repo, commit, event_id)` so duplicate hooks do not create duplicate runs.
- **Manual kickoff**: Optional idempotency key for UI/API clients that may double-submit.

## Rate limiting

- Apply **per-IP** and **per-token** (or per-installation / per-tenant) limits on public Orchestrator routes, especially webhook and unauthenticated endpoints, to reduce abuse and accidental storms.

## Feature flags

- Gate **cloud prereq validation** in Validate and **automatic IaC/cloud detection** in Deploy so MVP ships with **AWS + Terraform** while the codebase stays ready for Helm, Ansible, GCP, and Azure without large refactors.

## Backup and restore

- **MVP**: Document that containerized Postgres and local File volumes are **ephemeral-friendly**; recreating the stack may lose history unless volumes are backed up.
- **Future**: When using external managed Postgres or object storage, document backup frequency, RPO/RTO, and restore drills in runbooks (Phase 4).
