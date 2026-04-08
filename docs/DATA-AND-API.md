# Data model and API sketch

This document captures **conceptual** entities, retention rules, and API behavior for Phase 2 implementation. Field names and migrations will be finalized in code.

## Tenancy (SaaS)

- **Model**: One adaptive-pipe deployment serves **many tenants** (customers). All durable entities that belong to a customer carry a **`tenant_id`** (or equivalent). APIs and webhooks must **resolve the tenant** before reading or writing state (for example tenant slug in the webhook URL, GitHub App installation mapping, or authenticated session tenant claim).
- **Isolation**: Queries, queue dispatch, and File paths are always **scoped by `tenant_id`** so one tenant cannot read or mutate another’s runs, credentials, or artifacts (see [SECURITY-AND-OPERATIONS.md](SECURITY-AND-OPERATIONS.md)).
- **`admin`**: means **admin within the tenant** unless you add a separate **platform super-admin** role for operator tasks.

## Core entities

### Repository context

- **`tenant_id`**: owning customer for this repo context.  
- **`github_org`**, **`github_repo`**: identify the GitHub.com repository (**MVP**) within that tenant; future: configurable API base for GitHub Enterprise Server / on-prem.  
- **`build_number`**: monotonic integer **per** `(tenant_id, github_org, github_repo)` for display and correlation.  
- **`commit_sha`**, optional **`version_tag`**: attached to the build when available.

### Pipeline run

- **`run`**: one execution from kickoff through terminal state.  
  - Foreign keys: **`tenant_id`**, org/repo identifiers, `build_number`, commit metadata.  
  - `status`: queued, running, succeeded, failed, canceled, etc.  
  - `current_stage` and ordered **stage records** (below).  
  - **Sticky assignments**: which File node (and other pool members) this run is bound to.

### Stage

- Ordered stages: **validate → file → build → test → deploy** (exact names fixed in config).  
- Each stage has: `state` (pending, skipped, running, succeeded, failed), timestamps, optional **ETA** fields derived from history (**simple moving average** for MVP—replaceable implementation), and links to **log** storage.  
- **Skipped stages** remain visible in API responses so the UI can render the row without progress.

### Logs and history

- Log bodies may be **chunks in DB** (small) or **object references** (File node; future: object storage) for large streams; Phase 2 chooses based on volume. For high concurrency, **prefer references / segments on the File layer** so the Orchestrator is not a log proxy (see [ARCHITECTURE.md](ARCHITECTURE.md)).  
- Historical queries for “completed” runs read from the DB (and blob refs if used).

### Credentials

- **`credential`**: logical secret (name, type, scope such as org/repo or tenant-global). Always includes **`tenant_id`**.  
- Ciphertext + key id + nonce (if applicable); **never** store plaintext.  
- **Admin-only** create/update/delete; audit who changed what and when.

### Users and roles

- **MVP**: **local username/password** users **bound to a `tenant_id`** (signup or invite flow defines membership).  
- **Future**: Additional auth methods (OIDC, SSO) without changing run/credential core models; tenant claim comes from identity provider or session.  
- Role **`admin`** (tenant-scoped) required for credential mutations and that tenant’s platform settings (for example warm pool sizes).

### Platform settings

- **Warm pool sizes**: persisted per **`tenant_id`** for Build, Test, and Deploy (each customer tunes their own spool). Zero means **scale-from-zero** until demand creates workers.  
- Optional **platform operator** settings (global caps, feature flags) may live in config or a super-admin surface; align with [SECURITY-AND-OPERATIONS.md](SECURITY-AND-OPERATIONS.md).

## Retention: “10 builds” of material

Product requirement: retain enough **material** so any of the **last 10 builds** can be **re-initiated at any stage** (given prerequisites).

**Definition (implement in Phase 2):**

- Scope: **per** `(tenant_id, github_org, github_repo)` — e.g. tenant A `org1/repo1` retains its own 10, tenant A `org1/repo2` another 10; tenant B has independent windows.  
- **Which 10**: the **10 most recent runs by `build_number`** for that triple, **regardless of status** (success, failure, canceled all count).  
- **Material** includes:  
  - Source snapshot references (commit + cached tree on File node).  
  - Artifact references from build.  
  - Queue-related tokens or messages needed to resume (if any).  
- Older runs: material may be **garbage-collected**; long-term DB retention for analytics is a separate policy (see [CALLOUTS.md](../CALLOUTS.md) optional follow-ups).

### Garbage collection (11th-and-older material) — risk and pattern

Dropping material for builds **outside** the last 10 per `(tenant_id, org, repo)` must avoid **orphan files** on File nodes and **dangling references** in Postgres.

**Recommended approach (Phase 2):**

1. **Selection**: Compute the set of `run_id`s (or build numbers) that **remain** in the retention window vs those **eligible for purge** for a given repo. Do not purge any run that is **still active** (queued/running).  
2. **Serialize per scope**: Run GC jobs **per** `(tenant_id, github_org, github_repo)` (or shard) with a **lock** or lease so two workers never purge the same repo concurrently.  
3. **Two-phase delete**:  
   - **Phase A (metadata)**: Mark eligible runs as `material_purged` (or delete child rows in a transaction) only after File delete succeeds **or** after recording File paths to delete in an **outbox** table consumed by a File janitor.  
   - **Phase B (files)**: File service deletes paths **listed** for that purge; on success, finalize DB state; on failure, **retry** janitor and surface alerts for stuck paths.  
4. **Reconciliation**: Periodic job compares File disk usage or listings against DB references and removes **unreferenced** blobs (with a grace period) to catch partial failures.  
5. **Queue/Redis**: Ensure messages or tokens tied to purged runs are **acked/removed** or expired so old work cannot be claimed after material is gone.

Exact ordering (DB-first vs file-first) should favor **no user-visible reference to deleted bytes**: either delete files first then DB, or tombstone in DB then delete files and clear tombstone—pick one ordering and test both crash mid-flight.

## Orchestrator: kickoff

- **Behavior**: validate payload, create `run` + initial queue entries, return **immediately** once acceptance is **durable** (written to DB / outbox as implemented).  
- **HTTP status**: return the most appropriate **2xx success** code; **200 OK** is correct when the response body confirms the kickoff was accepted and recorded. Use another 2xx only if a future API contract explicitly needs it.  
- **Idempotency**: **required** for production-grade webhooks—support dedupe via GitHub delivery id, `Idempotency-Key`, or `(tenant_id, org, repo, commit, delivery_id)` (see [SECURITY-AND-OPERATIONS.md](SECURITY-AND-OPERATIONS.md)).

### Example kickoff request body (illustrative)

```json
{
  "github_org": "acme",
  "github_repo": "widget",
  "commit_sha": "abc123...",
  "ref": "refs/heads/main",
  "requested_by": "github_actions",
  "metadata": {
    "workflow_run_id": "12345"
  }
}
```

### Example kickoff response

```json
{
  "run_id": "uuid",
  "build_number": 42,
  "status": "queued"
}
```

## GitHub integration

**Inbound (MVP)**

- **github.com** only at first. A **commit-triggered webhook** initiates runs (push or a commit-oriented event—exact event list fixed in implementation).  
- **Tenant routing**: each tenant’s webhook URL or GitHub App installation must map unambiguously to a **`tenant_id`** before enqueueing work.  
- Payload carries org, repo, commit; use webhook **signature** verification. Secrets reference the credential store **for that tenant**; never log tokens.

**Future**

- Configurable **API base URL** for GitHub Enterprise Server / on-prem.

**Outbound (status back to GitHub)**

- Implementation chooses **Checks API**, **Commit Statuses**, or another supported mechanism; document the chosen approach, required GitHub App/OAuth scopes, and payload mapping in Phase 4. The product requirement is that **GitHub CI can observe progress**—exact API is an implementation detail stable behind documented behavior.

## Read APIs (UI-oriented)

Illustrative resources (REST)—all scoped by **authenticated tenant** (path prefix, header, or session; pick one style and keep it consistent):

- `GET /repos/{org}/{repo}/builds` — list recent runs with `build_number`, commit, status.  
- `GET /repos/{org}/{repo}/builds/{build_number}` — detail with ordered stages, ETAs, log pointers.  
- `POST /repos/{org}/{repo}/builds` — manual kickoff (authenticated).  
- `POST /repos/{org}/{repo}/builds/{build_number}/stages/{stage}/retry` — kick off from a stage when material exists (exact rules in Phase 2).  
- `GET/POST /admin/credentials` — tenant **admin** only.  
- `GET/PATCH /platform/settings` (or equivalent) — that tenant’s warm pool sizes and flags.

Authentication: session cookie or bearer token **including tenant membership**; align with UI login flow.

## Queue messages (conceptual)

Each message references `tenant_id`, `run_id`, `stage`, and dispatch hints (image, resource class). **Redis** backs the queue. Consumers acknowledge completion; Orchestrator updates DB and enqueues the next stage. Dead-letter handling and visibility timeouts are implementation details in Phase 2.
