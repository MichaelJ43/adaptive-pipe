import { useEffect, useState } from "react";
import { apiFetch } from "../api";

export default function SettingsPage() {
  const [b, setB] = useState(0);
  const [t, setT] = useState(0);
  const [d, setD] = useState(0);
  const [err, setErr] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    (async () => {
      const res = await apiFetch("/api/v1/platform/settings");
      if (!res.ok) {
        setErr(await res.text());
        return;
      }
      const j = await res.json();
      setB(j.build_warm_pool ?? 0);
      setT(j.test_warm_pool ?? 0);
      setD(j.deploy_warm_pool ?? 0);
    })();
  }, []);

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setErr(null);
    setSaved(false);
    const res = await apiFetch("/api/v1/platform/settings", {
      method: "PATCH",
      body: JSON.stringify({
        build_warm_pool: b,
        test_warm_pool: t,
        deploy_warm_pool: d,
      }),
    });
    if (!res.ok) {
      setErr(await res.text());
      return;
    }
    setSaved(true);
  }

  return (
    <div>
      <h1>Platform settings</h1>
      <p className="muted">
        Warm pool targets per tenant (Build / Test / Deploy). Zero means scale-from-zero; higher
        values reserve idle capacity when your runtime supports it.
      </p>
      <div className="card">
        <form onSubmit={save}>
          <label htmlFor="bw">Build warm pool</label>
          <input
            id="bw"
            type="number"
            min={0}
            value={b}
            onChange={(e) => setB(Number(e.target.value))}
          />
          <label htmlFor="tw">Test warm pool</label>
          <input
            id="tw"
            type="number"
            min={0}
            value={t}
            onChange={(e) => setT(Number(e.target.value))}
          />
          <label htmlFor="dw">Deploy warm pool</label>
          <input
            id="dw"
            type="number"
            min={0}
            value={d}
            onChange={(e) => setD(Number(e.target.value))}
          />
          {err && <p style={{ color: "#b91c1c" }}>{err}</p>}
          {saved && <p style={{ color: "#16a34a" }}>Saved.</p>}
          <button type="submit" className="primary">
            Save (admin)
          </button>
        </form>
      </div>
    </div>
  );
}
