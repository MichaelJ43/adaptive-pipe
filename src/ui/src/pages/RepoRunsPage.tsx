import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { apiFetch } from "../api";

type RunRow = {
  run_id: string;
  build_number: number;
  status: string;
};

export default function RepoRunsPage() {
  const { org, repo } = useParams<{ org: string; repo: string }>();
  const [runs, setRuns] = useState<RunRow[] | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const [commit, setCommit] = useState("abc123deadbeef");
  const [busy, setBusy] = useState(false);

  async function load() {
    if (!org || !repo) return;
    setErr(null);
    const res = await apiFetch(
      `/api/v1/repos/${encodeURIComponent(org)}/${encodeURIComponent(repo)}/runs`
    );
    if (!res.ok) {
      setErr(await res.text());
      return;
    }
    setRuns(await res.json());
  }

  useEffect(() => {
    void load();
  }, [org, repo]);

  async function kickoff(e: React.FormEvent) {
    e.preventDefault();
    if (!org || !repo) return;
    setBusy(true);
    setErr(null);
    try {
      const res = await apiFetch(`/api/v1/runs`, {
        method: "POST",
        body: JSON.stringify({
          github_org: org,
          github_repo: repo,
          commit_sha: commit,
        }),
      });
      if (!res.ok) throw new Error(await res.text());
      await load();
    } catch (ex) {
      setErr(ex instanceof Error ? ex.message : "Failed");
    } finally {
      setBusy(false);
    }
  }

  if (!org || !repo) return null;

  return (
    <div>
      <h1>
        {org}/{repo}
      </h1>
      <p className="muted">
        <Link to="/">Dashboard</Link>
      </p>

      <div className="card">
        <h2>Kick off build</h2>
        <form onSubmit={kickoff}>
          <label htmlFor="c">Commit SHA</label>
          <input id="c" value={commit} onChange={(e) => setCommit(e.target.value)} />
          <button type="submit" className="primary" disabled={busy}>
            Start pipeline
          </button>
        </form>
      </div>

      {err && <p style={{ color: "#b91c1c" }}>{err}</p>}

      <div className="card">
        <h2>Recent runs</h2>
        {runs === null && <p>Loading…</p>}
        {runs && runs.length === 0 && <p className="muted">No runs yet.</p>}
        {runs && runs.length > 0 && (
          <ul className="stages">
            {runs.map((r) => (
              <li key={r.run_id}>
                <Link to={`/repos/${encodeURIComponent(org)}/${encodeURIComponent(repo)}/runs/${r.run_id}`}>
                  Build #{r.build_number}
                </Link>
                <span>{r.status}</span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
