# Data model and API sketch

This document captures **conceptual** entities, retention rules, and API behavior for Phase 2 implementation. Field names and migrations will be finalized in code.

## Core entities

### Repository context

- **`github_org`**, **`github_repo`**: identify the GitHub (or GHE) repository.  
- **`build_number`**: monotonic integer **per** `(org, repo)` for display and correlation.  
- **`commit_sha`**, optional **`version_tag`**: attached to the build when available.

### Pipeline run

- **`run`**: one execution from kickoff through terminal state.  
  - Foreign keys: org/repo identifiers, `build_number`, commit metadata.  
  - `status`: queued, running, succeeded, failed, canceled, etc.  
  - `current_stage` and ordered **stage records** (below).  
  - **Sticky assignments**: which File node (and other pool members) this run is bound to.

### Stage

- Ordered stages: **validate → file → build → test → deploy** (exact names fixed in config).  
- Each stage has: `state` (pending, skipped, running, succeeded, failed), timestamps, optional **ETA** fields derived from history, and links to **log** storage.  
- **Skipped stages** remain visible in API responses so the UI can render the row without progress.

### Logs and history

- Log bodies may be **chunks in DB** (small) or **object references** (File node or S3-compatible storage) for large streams; Phase 2 chooses based on volume.  
- Historical queries for “completed” runs read from the DB (and blob refs if used).

### Credentials

- **`credential`**: logical secret (name, type, scope such as org/repo or global).  
- Ciphertext + key id + nonce (if applicable); **never** store plaintext.  
- **Admin-only** create/update/delete; audit who changed what and when.

### Users and roles

- Minimal v1: local users or OIDC (see [CALLOUTS.md](../CALLOUTS.md)).  
- Role **`admin`** required for credential mutations.

## Retention: “10 builds” of material

Product requirement: retain enough **material** so any of the **last 10 builds** for a repo can be **re-initiated at any stage** (given prerequisites).

**Proposed definition (implement exactly in Phase 2):**

- For each `(org, repo)`, consider the **10 most recent runs by `build_number`** (or by created time—pick one and keep consistent).  
- **Material** includes:  
  - Source snapshot references (commit + cached tarball/tree id on File node).  
  - Artifact references produced by build.  
  - Queue messages or job tokens tied to those runs (if any remain).  
- Runs **older** than that window may have material garbage-collected; **DB rows** may be archived or summarized per policy (callout: legal/compliance).

Exact GC policy and whether “10” means “last 10 regardless of status” vs “last 10 completed” should be confirmed (see [CALLOUTS.md](../CALLOUTS.md)); default recommendation: **last 10 runs by number including failed/canceled**, so retries behave predictably.

## Orchestrator: kickoff

- **Behavior**: validate payload, create `run` + initial queue entries, return **immediately** once the run is durable.  
- **HTTP status**: product prompt specifies **200** when kickoff is initiated; use **200** with a JSON body containing `run_id` / `build_number` (and optionally a `status_url`). If you later prefer **202 Accepted** for async semantics, treat that as a documented breaking change.  
- **Idempotency**: optional header or client-supplied idempotency key to avoid duplicate runs for the same webhook delivery (recommended in implementation).

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

**Inbound**

- Kickoff may come from a **GitHub webhook** (push, workflow_dispatch) or from the UI using the same API shape.  
- Payload should carry org, repo, commit, and optional **installation** or token reference (handled via credential store, not logged).

**Outbound (status back to GitHub)**

- MVP options: **Commit Statuses API** or **Checks API**. Checks are richer for PRs; pick one for v1 and document required OAuth/app permissions.  
- Orchestrator updates status when stages transition or on terminal failure.

## Read APIs (UI-oriented)

Illustrative resources (REST):

- `GET /repos/{org}/{repo}/builds` — list recent runs with `build_number`, commit, status.  
- `GET /repos/{org}/{repo}/builds/{build_number}` — detail with ordered stages, ETAs, log pointers.  
- `POST /repos/{org}/{repo}/builds` — manual kickoff (authenticated).  
- `POST /repos/{org}/{repo}/builds/{build_number}/stages/{stage}/retry` — kick off from a stage when material exists (exact rules in Phase 2).  
- `GET/POST /admin/credentials` — admin only.

Authentication: session cookie or bearer token; align with UI login flow.

## Queue messages (conceptual)

Each message references `run_id`, `stage`, and dispatch hints (image, resource class). Consumers acknowledge completion; Orchestrator updates DB and enqueues the next stage. Dead-letter handling and visibility timeouts are implementation details in Phase 2.
