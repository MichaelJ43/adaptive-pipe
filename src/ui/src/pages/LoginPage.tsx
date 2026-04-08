import { useState } from "react";
import { useNavigate, Navigate } from "react-router-dom";
import { getToken, login } from "../api";

export default function LoginPage() {
  const nav = useNavigate();
  const [tenant, setTenant] = useState("demo");
  const [user, setUser] = useState("admin");
  const [password, setPassword] = useState("admin123");
  const [err, setErr] = useState<string | null>(null);

  if (getToken()) return <Navigate to="/" replace />;

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErr(null);
    try {
      const token = await login({
        tenant_slug: tenant,
        username: user,
        password,
      });
      localStorage.setItem("ap_token", token);
      nav("/");
    } catch (ex) {
      setErr(ex instanceof Error ? ex.message : "Login failed");
    }
  }

  return (
    <div className="card" style={{ maxWidth: 400 }}>
      <h1>Sign in</h1>
      <p className="muted">
        Live SaaS console. Use tenant <strong>demo</strong> and the seeded admin user
        (default password <code>admin123</code> unless overridden).
      </p>
      <form onSubmit={onSubmit}>
        <label htmlFor="tenant">Tenant slug</label>
        <input
          id="tenant"
          value={tenant}
          onChange={(e) => setTenant(e.target.value)}
          autoComplete="username"
        />
        <label htmlFor="user">Username</label>
        <input
          id="user"
          value={user}
          onChange={(e) => setUser(e.target.value)}
        />
        <label htmlFor="pw">Password</label>
        <input
          id="pw"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          autoComplete="current-password"
        />
        {err && <p style={{ color: "#b91c1c" }}>{err}</p>}
        <button type="submit" className="primary">
          Sign in
        </button>
      </form>
    </div>
  );
}
