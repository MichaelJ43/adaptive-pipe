const base = import.meta.env.VITE_API_BASE || "";

export function apiUrl(path: string): string {
  if (path.startsWith("http")) return path;
  return `${base}${path}`;
}

export function getToken(): string | null {
  return localStorage.getItem("ap_token");
}

export async function apiFetch(
  path: string,
  init: RequestInit = {}
): Promise<Response> {
  const headers = new Headers(init.headers);
  const t = getToken();
  if (t) headers.set("Authorization", `Bearer ${t}`);
  if (!headers.has("Content-Type") && init.body && typeof init.body === "string") {
    headers.set("Content-Type", "application/json");
  }
  return fetch(apiUrl(path), { ...init, headers });
}

export type LoginBody = {
  tenant_slug: string;
  username: string;
  password: string;
};

export async function login(body: LoginBody): Promise<string> {
  const res = await fetch(apiUrl("/api/v1/auth/login"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(await res.text());
  const j = (await res.json()) as { token: string };
  return j.token;
}
