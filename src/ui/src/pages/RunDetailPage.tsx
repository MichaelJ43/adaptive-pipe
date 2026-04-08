import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { apiFetch } from "../api";

type Stage = { name: string; state: string };

type RunDetail = {
  run_id?: string;
  build_number: number;
  status: string;
  stages?: Stage[];
  eta_seconds?: Record<string, number>;
};

export default function RunDetailPage() {
  const { org, repo, runId } = useParams<{ org: string; repo: string; runId: string }>();
  const [run, setRun] = useState<RunDetail | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    if (!runId) return;
    let cancel = false;
    async function tick() {
      const res = await apiFetch(`/api/v1/runs/${encodeURIComponent(runId)}`);
      if (cancel) return;
      if (!res.ok) {
        setErr(await res.text());
        return;
      }
      setRun(await res.json());
      setErr(null);
    }
    void tick();
    const id = setInterval(tick, 2000);
    return () => {
      cancel = true;
      clearInterval(id);
    };
  }, [runId]);

  function cls(st: string) {
    if (st === "pending" || st === "skipped") return "state-pending";
    if (st === "running") return "state-running";
    if (st === "succeeded") return "state-succeeded";
    if (st === "failed") return "state-failed";
    return "";
  }

  if (!org || !repo || !runId) return null;

  return (
    <div>
      <p className="muted">
        <Link to={`/repos/${encodeURIComponent(org)}/${encodeURIComponent(repo)}`}>← Runs</Link>
      </p>
      {!run && !err && <p>Loading…</p>}
      {err && <p style={{ color: "#b91c1c" }}>{err}</p>}
      {run && (
        <div className="card">
          <h1>
            Build #{run.build_number}{" "}
            <span className={cls(run.status)}>{run.status}</span>
          </h1>
          <p className="muted">Run ID: {runId}</p>
          <h2>Stages</h2>
          <ul className="stages">
            {(run.stages || []).map((s) => (
              <li key={s.name}>
                <span>{s.name}</span>
                <span>
                  <span className={cls(s.state)}>{s.state}</span>
                  {run.eta_seconds && run.eta_seconds[s.name] != null && s.state === "pending" && (
                    <span className="muted" style={{ marginLeft: "0.5rem" }}>
                      ETA ~{run.eta_seconds[s.name].toFixed(1)}s
                    </span>
                  )}
                </span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
