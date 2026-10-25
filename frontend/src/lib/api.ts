export type Application = {
  id: string;
  name: string;
  host: string;
  upstream_url: string;
  enabled: boolean;
  waf_mode: string;
  rate_limit: string;
  tls_enabled: boolean;
  path_prefix: string;
  allow_cidrs: string;
  auth_mode: string;
  basic_auth_user: string;
  basic_auth_pass?: string;
  basic_auth_realm: string;
  custom_errors: boolean;
  notes: string;
  created_at?: string;
  updated_at?: string;
};

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    credentials: "include",
    headers: { "Content-Type": "application/json", ...(init?.headers || {}) },
    ...init,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || res.statusText);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export const api = {
  health: () => req<{ status: string }>("/api/health"),
  authOptions: () =>
    req<{ local: boolean; oidc: boolean; auth_mode: string }>("/api/auth/options"),
  login: (username: string, password: string) =>
    req<{ username: string }>("/api/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    }),
  logout: () => req("/api/logout", { method: "POST" }),
  me: () => req<{ username: string }>("/api/me"),
  dashboard: () =>
    req<{
      applications_total: number;
      applications_enabled: number;
      active_decisions: number;
      product: string;
      version: string;
    }>("/api/dashboard"),
  listApps: () => req<Application[]>("/api/applications"),
  getApp: (id: string) => req<Application>(`/api/applications/${id}`),
  createApp: (body: Partial<Application>) =>
    req<Application>("/api/applications", { method: "POST", body: JSON.stringify(body) }),
  updateApp: (id: string, body: Partial<Application>) =>
    req<Application>(`/api/applications/${id}`, { method: "PUT", body: JSON.stringify(body) }),
  deleteApp: (id: string) => req(`/api/applications/${id}`, { method: "DELETE" }),
  toggleApp: (id: string) => req<Application>(`/api/applications/${id}/toggle`, { method: "POST" }),
  appLogs: (id: string, opts?: { kind?: string; limit?: number; q?: string }) => {
    const p = new URLSearchParams();
    if (opts?.kind) p.set("kind", opts.kind);
    if (opts?.limit) p.set("limit", String(opts.limit));
    if (opts?.q) p.set("q", opts.q);
    const qs = p.toString();
    return req<{ host: string; kind: string; count: number; entries: any[] }>(
      `/api/applications/${id}/logs${qs ? `?${qs}` : ""}`
    );
  },
  decisions: () => req<{ decisions: any[]; warning?: string }>("/api/decisions"),
  audit: () => req<any[]>("/api/audit"),
  sync: () => req("/api/sync", { method: "POST" }),
  exportApps: () => req<Application[]>("/api/export"),
  updateCRS: () => req<{ message: string }>("/api/protections/update-crs", { method: "POST" }),
  updateCrowdSec: () => req<{ message: string }>("/api/protections/update-crowdsec", { method: "POST" }),
  protectionsStatus: () => req<any>("/api/protections/status"),
};