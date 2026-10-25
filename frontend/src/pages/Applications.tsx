import { FormEvent, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, Application } from "../lib/api";

const empty: Partial<Application> = {
  name: "",
  host: "",
  upstream_url: "http://",
  enabled: true,
  waf_mode: "full",
  rate_limit: "soft",
  tls_enabled: true,
  path_prefix: "/",
  allow_cidrs: "",
  auth_mode: "none",
  basic_auth_user: "",
  basic_auth_pass: "",
  basic_auth_realm: "The Gate",
  custom_errors: true,
  notes: "",
};

export default function Applications() {
  const [apps, setApps] = useState<Application[]>([]);
  const [stats, setStats] = useState({ applications_total: 0, applications_enabled: 0, active_decisions: 0 });
  const [form, setForm] = useState<Partial<Application>>(empty);
  const [editing, setEditing] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [showForm, setShowForm] = useState(false);

  async function load() {
    const [list, dash] = await Promise.all([api.listApps(), api.dashboard()]);
    setApps(list);
    setStats(dash);
  }

  useEffect(() => {
    load().catch((e) => setError(e.message));
  }, []);

  function startCreate() {
    setEditing(null);
    setForm(empty);
    setShowForm(true);
  }

  function startEdit(a: Application) {
    setEditing(a.id);
    setForm({ ...a, basic_auth_pass: "" });
    setShowForm(true);
  }

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    try {
      const payload = { ...form };
      if (editing && !payload.basic_auth_pass) {
        delete payload.basic_auth_pass;
      }
      if (editing) await api.updateApp(editing, payload);
      else await api.createApp(payload);
      setShowForm(false);
      await load();
    } catch (err: any) {
      setError(err.message);
    }
  }

  return (
    <>
      <div className="topbar">
        <div>
          <h1>Applications</h1>
          <p>Endpoints protected and routed through The Gate — WAF, auth, rate limits, custom errors.</p>
        </div>
        <div className="row">
          <button className="btn" onClick={() => api.sync().then(load).catch((e) => setError(e.message))}>
            Sync Traefik
          </button>
          <button className="btn primary" onClick={startCreate}>
            Add application
          </button>
        </div>
      </div>

      <div className="stats">
        <div className="stat">
          <em>Total apps</em>
          <strong>{stats.applications_total}</strong>
        </div>
        <div className="stat">
          <em>Enabled</em>
          <strong>{stats.applications_enabled}</strong>
        </div>
        <div className="stat">
          <em>Active decisions</em>
          <strong>{stats.active_decisions}</strong>
        </div>
      </div>

      {error && <div className="error" style={{ marginBottom: 16 }}>{error}</div>}

      {showForm && (
        <form className="panel form" style={{ marginBottom: 20 }} onSubmit={submit}>
          <h3 style={{ margin: 0 }}>{editing ? "Edit application" : "New application"}</h3>
          <label>
            Name
            <input required value={form.name || ""} onChange={(e) => setForm({ ...form, name: e.target.value })} />
          </label>
          <label>
            Public host
            <input
              required
              placeholder="app.example.com"
              value={form.host || ""}
              onChange={(e) => setForm({ ...form, host: e.target.value })}
            />
          </label>
          <label>
            Upstream URL
            <input
              required
              placeholder="http://backend:8080"
              value={form.upstream_url || ""}
              onChange={(e) => setForm({ ...form, upstream_url: e.target.value })}
            />
          </label>
          <label>
            WAF mode
            <select value={form.waf_mode || "full"} onChange={(e) => setForm({ ...form, waf_mode: e.target.value })}>
              <option value="full">Full (Coraza CRS + CrowdSec)</option>
              <option value="light">Light (CrowdSec + rate limit)</option>
              <option value="crowdsec-only">CrowdSec only</option>
              <option value="off">Off (proxy only)</option>
            </select>
          </label>
          <label>
            Rate limit / anti-DoS
            <select value={form.rate_limit || "soft"} onChange={(e) => setForm({ ...form, rate_limit: e.target.value })}>
              <option value="soft">Soft (100 rps)</option>
              <option value="strict">Strict (20 rps)</option>
              <option value="dos">Anti-DoS (5 rps)</option>
              <option value="off">Off</option>
            </select>
          </label>
          <label>
            Access auth
            <select value={form.auth_mode || "none"} onChange={(e) => setForm({ ...form, auth_mode: e.target.value })}>
              <option value="none">None</option>
              <option value="basic">Basic Auth</option>
              <option value="authentik">Authentik SSO (forward-auth)</option>
            </select>
          </label>
          {form.auth_mode === "basic" && (
            <>
              <label>
                Basic auth user
                <input
                  required
                  value={form.basic_auth_user || ""}
                  onChange={(e) => setForm({ ...form, basic_auth_user: e.target.value })}
                />
              </label>
              <label>
                Basic auth password {editing ? "(leave blank to keep)" : ""}
                <input
                  type="password"
                  required={!editing}
                  value={form.basic_auth_pass || ""}
                  onChange={(e) => setForm({ ...form, basic_auth_pass: e.target.value })}
                />
              </label>
              <label>
                Realm
                <input
                  value={form.basic_auth_realm || "The Gate"}
                  onChange={(e) => setForm({ ...form, basic_auth_realm: e.target.value })}
                />
              </label>
            </>
          )}
          <label>
            Path prefix
            <input
              value={form.path_prefix || "/"}
              onChange={(e) => setForm({ ...form, path_prefix: e.target.value })}
              placeholder="/"
            />
          </label>
          <label>
            Allowlist CIDRs (optional)
            <input
              value={form.allow_cidrs || ""}
              onChange={(e) => setForm({ ...form, allow_cidrs: e.target.value })}
              placeholder="10.0.0.0/8, 192.168.1.0/24"
            />
          </label>
          <label>
            Notes
            <textarea rows={2} value={form.notes || ""} onChange={(e) => setForm({ ...form, notes: e.target.value })} />
          </label>
          <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
            <input
              type="checkbox"
              checked={form.custom_errors !== false}
              onChange={(e) => setForm({ ...form, custom_errors: e.target.checked })}
            />
            Custom error pages (401/403/404/429/5xx)
          </label>
          <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
            <input
              type="checkbox"
              checked={form.tls_enabled !== false}
              onChange={(e) => setForm({ ...form, tls_enabled: e.target.checked })}
            />
            Let’s Encrypt TLS
          </label>
          <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
            <input
              type="checkbox"
              checked={!!form.enabled}
              onChange={(e) => setForm({ ...form, enabled: e.target.checked })}
            />
            Enabled
          </label>
          <div className="row">
            <button className="btn primary" type="submit">
              Save
            </button>
            <button className="btn" type="button" onClick={() => setShowForm(false)}>
              Cancel
            </button>
          </div>
        </form>
      )}

      {apps.length === 0 ? (
        <div className="empty">No applications yet. Add your first endpoint to open The Gate.</div>
      ) : (
        <div className="grid">
          {apps.map((a) => (
            <article className="card" key={a.id}>
              <header>
                <h3>{a.name}</h3>
                <span className={`badge ${a.enabled ? "" : "off"}`}>{a.enabled ? "online" : "disabled"}</span>
              </header>
              <div className="meta">{a.host}</div>
              <div className="meta">→ {a.upstream_url}</div>
              <div className="row">
                <span className="badge">{a.waf_mode}</span>
                {a.rate_limit && <span className="badge">rl {a.rate_limit}</span>}
                {a.auth_mode && a.auth_mode !== "none" && <span className="badge">{a.auth_mode}</span>}
                {a.custom_errors && <span className="badge">errors</span>}
              </div>
              {a.notes && <p style={{ margin: 0, color: "var(--muted)", fontSize: "0.9rem" }}>{a.notes}</p>}
              <div className="row">
                <button className="btn" onClick={() => startEdit(a)}>
                  Configure
                </button>
                <Link className="btn" to={`/apps/${a.id}/logs`}>
                  Logs
                </Link>
                <button
                  className="btn"
                  onClick={() => api.toggleApp(a.id).then(load).catch((e) => setError(e.message))}
                >
                  {a.enabled ? "Disable" : "Enable"}
                </button>
                <button
                  className="btn danger"
                  onClick={() => {
                    if (confirm(`Delete ${a.name}?`)) {
                      api.deleteApp(a.id).then(load).catch((e) => setError(e.message));
                    }
                  }}
                >
                  Delete
                </button>
              </div>
            </article>
          ))}
        </div>
      )}
    </>
  );
}