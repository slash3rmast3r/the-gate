import { FormEvent, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, Application, IPGroup } from "../lib/api";
import { useI18n } from "../i18n/I18nProvider";
import CountryPicker from "../components/CountryPicker";

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
  defense_mode: "normal",
  http_flood: true,
  bot_protect: true,
  attacks: true,
  allow_group_id: "",
  deny_group_id: "",
  blocked_countries: "",
  extra_hosts: "",
};

export default function Applications() {
  const { t } = useI18n();
  const [apps, setApps] = useState<Application[]>([]);
  const [groups, setGroups] = useState<IPGroup[]>([]);
  const [stats, setStats] = useState({ applications_total: 0, applications_enabled: 0, active_decisions: 0 });
  const [form, setForm] = useState<Partial<Application>>(empty);
  const [editing, setEditing] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [ok, setOk] = useState("");
  const [syncing, setSyncing] = useState(false);
  const [showForm, setShowForm] = useState(false);

  async function load() {
    const [list, dash, ipGroups] = await Promise.all([api.listApps(), api.dashboard(), api.ipGroups()]);
    setApps(list);
    setStats(dash);
    setGroups(ipGroups);
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

  function closeForm() {
    setShowForm(false);
    setEditing(null);
    setError("");
  }

  return (
    <>
      <div className="topbar">
        <div>
          {showForm ? (
            <>
              <button className="btn" type="button" onClick={closeForm} style={{ marginBottom: 10 }}>
                ← {t("applications.backToList")}
              </button>
              <h1>{editing ? t("applications.edit") : t("applications.new")}</h1>
              <p>{t("applications.formHelp")}</p>
            </>
          ) : (
            <>
              <h1>{t("applications.title")}</h1>
              <p>{t("applications.description")}</p>
            </>
          )}
        </div>
        {!showForm && (
          <div className="row">
            <button
              className="btn"
              disabled={syncing}
              onClick={async () => {
                setError("");
                setOk("");
                setSyncing(true);
                try {
                  const res = await api.sync();
                  await load();
                  setOk(res?.message || t("applications.syncOk"));
                } catch (e: any) {
                  setError(e.message);
                } finally {
                  setSyncing(false);
                }
              }}
            >
              {syncing ? t("applications.syncing") : t("applications.syncTraefik")}
            </button>
            <button className="btn primary" onClick={startCreate}>
              {t("applications.add")}
            </button>
          </div>
        )}
      </div>

      {!showForm && (
        <div className="stats">
          <div className="stat">
            <em>{t("applications.totalApps")}</em>
            <strong>{stats.applications_total}</strong>
          </div>
          <div className="stat">
            <em>{t("applications.enabled")}</em>
            <strong>{stats.applications_enabled}</strong>
          </div>
          <div className="stat">
            <em>{t("applications.activeDecisions")}</em>
            <strong>{stats.active_decisions}</strong>
          </div>
        </div>
      )}

      {error && <div className="error" style={{ marginBottom: 16 }}>{error}</div>}
      {ok && <div className="ok" style={{ marginBottom: 16 }}>{ok}</div>}

      {showForm && (
        <form className="panel form" style={{ marginBottom: 20 }} onSubmit={submit}>
          <div className="row" style={{ justifyContent: "space-between", alignItems: "center" }}>
            <h3 style={{ margin: 0 }}>{editing ? t("applications.edit") : t("applications.new")}</h3>
            <button className="btn" type="button" onClick={closeForm}>
              ← {t("applications.backToList")}
            </button>
          </div>
          <label>
            {t("applications.name")}
            <input required value={form.name || ""} onChange={(e) => setForm({ ...form, name: e.target.value })} />
          </label>
          <label>
            {t("applications.publicHost")}
            <input
              required
              placeholder="app.example.com"
              value={form.host || ""}
              onChange={(e) => setForm({ ...form, host: e.target.value })}
            />
          </label>
          <label>
            {t("applications.upstreamUrl")}
            <input
              required
              placeholder="http://backend:8080"
              value={form.upstream_url || ""}
              onChange={(e) => setForm({ ...form, upstream_url: e.target.value })}
            />
          </label>
          <label>
            {t("applications.wafMode")}
            <select value={form.waf_mode || "full"} onChange={(e) => setForm({ ...form, waf_mode: e.target.value })}>
              <option value="full">{t("applications.waf.full")}</option>
              <option value="light">{t("applications.waf.light")}</option>
              <option value="crowdsec-only">{t("applications.waf.crowdsecOnly")}</option>
              <option value="off">{t("applications.waf.off")}</option>
            </select>
          </label>
          <label>
            {t("applications.rateLimit")}
            <select value={form.rate_limit || "soft"} onChange={(e) => setForm({ ...form, rate_limit: e.target.value })}>
              <option value="soft">{t("applications.rate.soft")}</option>
              <option value="strict">{t("applications.rate.strict")}</option>
              <option value="dos">{t("applications.rate.dos")}</option>
              <option value="off">{t("applications.rate.off")}</option>
            </select>
          </label>
          <label>
            {t("applications.accessAuth")}
            <select value={form.auth_mode || "none"} onChange={(e) => setForm({ ...form, auth_mode: e.target.value })}>
              <option value="none">{t("applications.auth.none")}</option>
              <option value="basic">{t("applications.auth.basic")}</option>
              <option value="authentik">{t("applications.auth.authentik")}</option>
            </select>
          </label>
          <label>
            {t("applications.allowGroup")}
            <select value={form.allow_group_id || ""} onChange={(e) => setForm({ ...form, allow_group_id: e.target.value })}>
              <option value="">{t("applications.none")}</option>
              {groups.filter((g) => g.kind === "allow").map((g) => (
                <option key={g.id} value={g.id}>{g.name}</option>
              ))}
            </select>
          </label>
          <label>
            {t("applications.denyGroup")}
            <select value={form.deny_group_id || ""} onChange={(e) => setForm({ ...form, deny_group_id: e.target.value })}>
              <option value="">{t("applications.none")}</option>
              {groups.filter((g) => g.kind === "deny").map((g) => (
                <option key={g.id} value={g.id}>{g.name}</option>
              ))}
            </select>
          </label>
          <div>
            <strong style={{ display: "block", marginBottom: 8 }}>{t("applications.blockedCountries")}</strong>
            <CountryPicker
              value={form.blocked_countries || ""}
              onChange={(csv) => setForm({ ...form, blocked_countries: csv })}
            />
          </div>
          <label>
            {t("applications.extraHosts")}
            <input
              placeholder={t("applications.extraHostsHint")}
              value={form.extra_hosts || ""}
              onChange={(e) => setForm({ ...form, extra_hosts: e.target.value.toLowerCase() })}
            />
          </label>
          <div className="panel" style={{ boxShadow: "none", display: "grid", gap: 10 }}>
            <strong>{t("applications.defenseToggles")}</strong>
            <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
              <input type="checkbox" checked={form.http_flood !== false} onChange={(e) => setForm({ ...form, http_flood: e.target.checked })} />
              {t("applications.httpFloodProtection")}
            </label>
            <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
              <input type="checkbox" checked={form.bot_protect !== false} onChange={(e) => setForm({ ...form, bot_protect: e.target.checked })} />
              {t("applications.botProtection")}
            </label>
            <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
              <input type="checkbox" checked={form.attacks !== false} onChange={(e) => setForm({ ...form, attacks: e.target.checked })} />
              {t("applications.applicationAttacks")}
            </label>
          </div>
          {form.auth_mode === "basic" && (
            <>
              <label>
                {t("applications.basicAuthUser")}
                <input
                  required
                  value={form.basic_auth_user || ""}
                  onChange={(e) => setForm({ ...form, basic_auth_user: e.target.value })}
                />
              </label>
              <label>
                {t("applications.basicAuthPassword")} {editing ? t("applications.leaveBlankToKeep") : ""}
                <input
                  type="password"
                  required={!editing}
                  value={form.basic_auth_pass || ""}
                  onChange={(e) => setForm({ ...form, basic_auth_pass: e.target.value })}
                />
              </label>
              <label>
                {t("applications.realm")}
                <input
                  value={form.basic_auth_realm || "The Gate"}
                  onChange={(e) => setForm({ ...form, basic_auth_realm: e.target.value })}
                />
              </label>
            </>
          )}
          <label>
            {t("applications.pathPrefix")}
            <input
              value={form.path_prefix || "/"}
              onChange={(e) => setForm({ ...form, path_prefix: e.target.value })}
              placeholder="/"
            />
          </label>
          <label>
            {t("applications.allowlistCidrs")}
            <input
              value={form.allow_cidrs || ""}
              onChange={(e) => setForm({ ...form, allow_cidrs: e.target.value })}
              placeholder="10.0.0.0/8, 192.168.1.0/24"
            />
          </label>
          <label>
            {t("applications.notes")}
            <textarea rows={2} value={form.notes || ""} onChange={(e) => setForm({ ...form, notes: e.target.value })} />
          </label>
          <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
            <input
              type="checkbox"
              checked={form.custom_errors !== false}
              onChange={(e) => setForm({ ...form, custom_errors: e.target.checked })}
            />
            {t("applications.customErrors")}
          </label>
          <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
            <input
              type="checkbox"
              checked={form.tls_enabled !== false}
              onChange={(e) => setForm({ ...form, tls_enabled: e.target.checked })}
            />
            {t("applications.letsEncryptTls")}
          </label>
          <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
            <input
              type="checkbox"
              checked={!!form.enabled}
              onChange={(e) => setForm({ ...form, enabled: e.target.checked })}
            />
            {t("applications.enabled")}
          </label>
          <div className="row">
            <button className="btn primary" type="submit">
              {t("common.save")}
            </button>
            <button className="btn" type="button" onClick={closeForm}>
              ← {t("applications.backToList")}
            </button>
          </div>
        </form>
      )}

      {!showForm && apps.length === 0 ? (
        <div className="empty">{t("applications.noApps")}</div>
      ) : !showForm ? (
        <div className="grid">
          {apps.map((a) => (
            <article className="card" key={a.id}>
              <header>
                <h3>{a.name}</h3>
                <span className={`badge ${a.enabled ? "" : "off"}`}>{a.enabled ? t("applications.online") : t("applications.disabled")}</span>
              </header>
              <div className="meta">{a.host}</div>
              <div className="meta">→ {a.upstream_url}</div>
              <div className="row">
                <span className="badge">{a.waf_mode}</span>
                <span className={`badge ${a.defense_mode === "under_attack" ? "off" : ""}`}>
                  {a.defense_mode === "under_attack" ? t("applications.defense.underAttack") : t(`applications.defense.${a.defense_mode || "normal"}`)}
                </span>
                {a.http_flood && <span className="badge">{t("applications.httpFloodChip")}</span>}
                {a.bot_protect && <span className="badge">{t("applications.botProtectChip")}</span>}
                {a.auth_mode && a.auth_mode !== "none" && <span className="badge">{t("applications.authChip")}</span>}
                {a.attacks && <span className="badge">{t("applications.attacksChip")}</span>}
                {a.rate_limit && <span className="badge">rl {a.rate_limit}</span>}
                {a.custom_errors && <span className="badge">{t("applications.errorsChip")}</span>}
                {a.blocked_countries && <span className="badge">GEO {a.blocked_countries}</span>}
              </div>
              {a.notes && <p style={{ margin: 0, color: "var(--muted)", fontSize: "0.9rem" }}>{a.notes}</p>}
              <div className="row">
                {(["normal", "audit", "under_attack"] as const).map((mode) => (
                  <button
                    key={mode}
                    className={`btn ${a.defense_mode === mode ? "accent" : ""}`}
                    onClick={() => api.setDefenseMode(a.id, mode).then(load).catch((e) => setError(e.message))}
                  >
                    {mode === "under_attack" ? t("applications.defense.underAttack") : t(`applications.defense.${mode}`)}
                  </button>
                ))}
              </div>
              <div className="row">
                <button className="btn" onClick={() => startEdit(a)}>
                  {t("applications.configure")}
                </button>
                <Link className="btn" to={`/apps/${a.id}/logs`}>
                  {t("applications.logs")}
                </Link>
                <button
                  className="btn"
                  onClick={() => api.toggleApp(a.id).then(load).catch((e) => setError(e.message))}
                >
                  {a.enabled ? t("applications.disable") : t("applications.enable")}
                </button>
                <button
                  className="btn danger"
                  onClick={() => {
                    if (confirm(`${t("applications.deleteConfirm")} ${a.name}?`)) {
                      api.deleteApp(a.id).then(load).catch((e) => setError(e.message));
                    }
                  }}
                >
                  {t("common.delete")}
                </button>
              </div>
            </article>
          ))}
        </div>
      ) : null}
    </>
  );
}