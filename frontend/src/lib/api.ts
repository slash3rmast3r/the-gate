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
  defense_mode: "normal" | "audit" | "under_attack" | string;
  http_flood: boolean;
  bot_protect: boolean;
  attacks: boolean;
  defense_snapshot?: string;
  allow_group_id: string;
  deny_group_id: string;
  blocked_countries: string;
  extra_hosts: string;
  antibot_mode: string;
  dnsbl_enabled: boolean;
  www_redirect: string;
  request_headers: string;
  response_headers: string;
  websocket_sticky: boolean;
  created_at?: string;
  updated_at?: string;
};

export type SecurityRule = {
  id: string;
  name: string;
  enabled: boolean;
  priority: number;
  scope: "global" | "app" | string;
  app_id: string;
  field: string;
  header_name: string;
  operator: string;
  value: string;
  action: "block" | "allow" | "log" | string;
  notes: string;
  hit_count?: number;
};

export type CertificateInfo = {
  resolver: string;
  main: string;
  sans: string[];
  domains: string[];
  issuer: string;
  not_before: string;
  not_after: string;
  days_left: number;
  app_id?: string;
  app_name?: string;
  app_host?: string;
};

export type IPGroup = {
  id: string;
  name: string;
  kind: "allow" | "deny";
  cidrs: string;
  notes: string;
  hit_count?: number;
};

export type OpsSettings = {
  auto_update_enabled: boolean;
  auto_update_hours: number;
  log_retention_days: number;
  error_tip_message: string;
  error_theme?: string;
  telegram: { enabled: boolean; bot_token: string; chat_id: string };
  smtp: { enabled: boolean; host: string; port: number; user: string; password: string; from: string; to: string };
  bad_behavior: {
    enabled: boolean;
    threshold: number;
    window_sec: number;
    min_status: number;
    ban_duration: string;
  };
  dnsbl_zones: string;
};

export type ErrorTheme = {
  id: string;
  name_en: string;
  name_it: string;
  ink: string;
  muted: string;
  brand: string;
  background: string;
  preview: string;
};

export type TrafficSummary = {
  window_minutes: number;
  requests: number;
  errors: number;
  blocked: number;
  applications: Array<{
    application_id: string;
    name: string;
    host: string;
    requests: number;
    errors: number;
    blocked: number;
    unique_clients: number;
    top_country: string;
  }>;
  clients: Array<{ ip: string; country: string; requests: number; errors: number; blocked: number }>;
  countries: Record<string, number>;
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
  bulkUpdateApps: (ids: string[], set: Record<string, unknown>) =>
    req<{ ok: boolean; updated: number; ids: string[] }>("/api/applications/bulk", {
      method: "POST",
      body: JSON.stringify({ ids, set }),
    }),
  toggleApp: (id: string) => req<Application>(`/api/applications/${id}/toggle`, { method: "POST" }),
  setDefenseMode: (id: string, mode: "normal" | "audit" | "under_attack") =>
    req<Application>(`/api/applications/${id}/defense`, { method: "POST", body: JSON.stringify({ mode }) }),
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
  deleteDecision: (scope: string, value: string) =>
    req<{ ok: boolean }>(`/api/decisions?scope=${encodeURIComponent(scope || "")}&value=${encodeURIComponent(value)}`, {
      method: "DELETE",
    }),
  audit: () => req<any[]>("/api/audit"),
  sync: () => req<{ ok: boolean; message?: string; written?: number }>("/api/sync", { method: "POST" }),
  exportApps: () => req<Application[]>("/api/export"),
  backup: async () => {
    const res = await fetch("/api/backup", { method: "POST", credentials: "include" });
    if (!res.ok) throw new Error((await res.text()) || res.statusText);
    return res.blob();
  },
  settings: () => req<OpsSettings>("/api/settings"),
  updateSettings: (body: OpsSettings) => req<OpsSettings>("/api/settings", { method: "PUT", body: JSON.stringify(body) }),
  ipGroups: () => req<IPGroup[]>("/api/ip-groups"),
  createIPGroup: (body: Partial<IPGroup>) => req<IPGroup>("/api/ip-groups", { method: "POST", body: JSON.stringify(body) }),
  updateIPGroup: (id: string, body: Partial<IPGroup>) =>
    req<IPGroup>(`/api/ip-groups/${id}`, { method: "PUT", body: JSON.stringify(body) }),
  deleteIPGroup: (id: string) => req<{ ok: boolean }>(`/api/ip-groups/${id}`, { method: "DELETE" }),
  securityRules: () => req<SecurityRule[]>("/api/security-rules"),
  createSecurityRule: (body: Partial<SecurityRule>) =>
    req<SecurityRule>("/api/security-rules", { method: "POST", body: JSON.stringify(body) }),
  updateSecurityRule: (id: string, body: Partial<SecurityRule>) =>
    req<SecurityRule>(`/api/security-rules/${id}`, { method: "PUT", body: JSON.stringify(body) }),
  deleteSecurityRule: (id: string) => req<{ ok: boolean }>(`/api/security-rules/${id}`, { method: "DELETE" }),
  certificates: () =>
    req<{ certificates: CertificateInfo[]; storage?: string; note?: string }>("/api/certificates"),
  renewCertificate: (domain: string) =>
    req<{ ok: boolean; domain: string; removed: number; message: string }>("/api/certificates/renew", {
      method: "POST",
      body: JSON.stringify({ domain }),
    }),
  traffic: (minutes = 60) => req<TrafficSummary>(`/api/traffic?minutes=${minutes}`),
  testNotifications: () => req<{ ok: boolean; errors?: string[] }>("/api/notifications/test", { method: "POST" }),
  errorPages: () =>
    req<{ codes: string[]; pages: Record<string, string>; tip_message: string; theme?: string; themes?: ErrorTheme[] }>(
      "/api/errorpages"
    ),
  updateErrorPage: (code: string, html: string) =>
    req<{ ok: boolean; code: string }>(`/api/errorpages/${code}`, { method: "PUT", body: JSON.stringify({ html }) }),
  applyErrorTheme: (theme: string) =>
    req<{ ok: boolean; theme: string }>("/api/errorpages/theme", { method: "POST", body: JSON.stringify({ theme }) }),
  updateCRS: () => req<{ message: string }>("/api/protections/update-crs", { method: "POST" }),
  updateCrowdSec: () => req<{ message: string }>("/api/protections/update-crowdsec", { method: "POST" }),
  protectionsStatus: () => req<any>("/api/protections/status"),
};