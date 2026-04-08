# Callouts

Most product and engineering decisions are captured in [project_prompt.txt](project_prompt.txt) and under [docs/](docs/). This file only holds items that are worth keeping visible outside those documents.

---

## Tenancy

**Decision:** adaptive-pipe is a **multi-tenant SaaS**—one installation, many customers, hard boundaries between tenants (data, credentials, builds, warm-pool settings). See [docs/DATA-AND-API.md](docs/DATA-AND-API.md) and [docs/SECURITY-AND-OPERATIONS.md](docs/SECURITY-AND-OPERATIONS.md) for how `tenant_id` and routing show up in the model.

**Still separate:** whether *your* application on AWS is single-tenant or multi-tenant is entirely about what you deploy; it is not the same as pipeline-product tenancy.

---

## Optional follow-ups (not blocking MVP)

- **Backup and restore**: Define RPO/RTO when Postgres and File storage move beyond local dev.
- **Legal / compliance**: Log and artifact retention beyond the “10 builds” material window if regulations apply.
