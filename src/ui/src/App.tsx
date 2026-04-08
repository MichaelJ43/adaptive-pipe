import type { ReactNode } from "react";
import { Navigate, Route, Routes, Link, useNavigate } from "react-router-dom";
import LoginPage from "./pages/LoginPage";
import DashboardPage from "./pages/DashboardPage";
import RepoRunsPage from "./pages/RepoRunsPage";
import RunDetailPage from "./pages/RunDetailPage";
import SettingsPage from "./pages/SettingsPage";
import { getToken } from "./api";

function Layout({ children }: { children: ReactNode }) {
  const nav = useNavigate();
  const authed = !!getToken();
  return (
    <div className="layout">
      <header className="header">
        <Link to="/" className="brand">
          Adaptive Pipe
        </Link>
        <nav className="nav">
          {authed && (
            <>
              <Link to="/">Dashboard</Link>
              <Link to="/settings">Platform settings</Link>
              <button
                type="button"
                className="linkbtn"
                onClick={() => {
                  localStorage.removeItem("ap_token");
                  nav("/login");
                }}
              >
                Sign out
              </button>
            </>
          )}
        </nav>
      </header>
      <main className="main">{children}</main>
    </div>
  );
}

function RequireAuth({ children }: { children: ReactNode }) {
  if (!getToken()) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route
          path="/"
          element={
            <RequireAuth>
              <DashboardPage />
            </RequireAuth>
          }
        />
        <Route
          path="/repos/:org/:repo"
          element={
            <RequireAuth>
              <RepoRunsPage />
            </RequireAuth>
          }
        />
        <Route
          path="/repos/:org/:repo/runs/:runId"
          element={
            <RequireAuth>
              <RunDetailPage />
            </RequireAuth>
          }
        />
        <Route
          path="/settings"
          element={
            <RequireAuth>
              <SettingsPage />
            </RequireAuth>
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Layout>
  );
}
