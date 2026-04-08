import { useState } from "react";
import { useNavigate } from "react-router-dom";

export default function DashboardPage() {
  const nav = useNavigate();
  const [org, setOrg] = useState("acme");
  const [repo, setRepo] = useState("widget");

  function go(e: React.FormEvent) {
    e.preventDefault();
    nav(`/repos/${encodeURIComponent(org)}/${encodeURIComponent(repo)}`);
  }

  return (
    <div>
      <h1>Dashboard</h1>
      <p className="muted">
        Open a GitHub org/repo pipeline view. Kick off builds from the repo page or via
        GitHub webhook (see docs).
      </p>
      <div className="card">
        <h2>Go to repository</h2>
        <form onSubmit={go}>
          <label htmlFor="org">GitHub org</label>
          <input id="org" value={org} onChange={(e) => setOrg(e.target.value)} />
          <label htmlFor="repo">Repository</label>
          <input id="repo" value={repo} onChange={(e) => setRepo(e.target.value)} />
          <button type="submit" className="primary">
            View runs
          </button>
        </form>
      </div>
    </div>
  );
}
