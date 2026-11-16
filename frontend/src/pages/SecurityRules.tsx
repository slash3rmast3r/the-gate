import { FormEvent, useEffect, useState } from "react";
import { api, Application, SecurityRule } from "../lib/api";
import { useI18n } from "../i18n/I18nProvider";

const blank: Partial<SecurityRule> = {
  name: "",
  enabled: true,
  priority: 100,
  scope: "app",
  app_id: "",
  field: "uri",
  header_name: "",
  operator: "contains",
  value: "",
  action: "block",
  notes: "",
};

export default function SecurityRules() {
  const { t } = useI18n();
  const [rules, setRules] = useState<SecurityRule[]>([]);
  const [apps, setApps] = useState<Application[]>([]);
  const [form, setForm] = useState<Partial<SecurityRule>>(blank);
  const [editing, setEditing] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [ok, setOk] = useState("");

  async function load() {
    const [r, a] = await Promise.all([api.securityRules(), api.listApps()]);
    setRules(r);
    setApps(a);
  }

  useEffect(() => {
    load().catch((e) => setError(e.message));
  }, []);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setOk("");
    try {
      const payload = { ...form };
      if (payload.scope === "global") payload.app_id = "";
      if (editing) await api.updateSecurityRule(editing, payload);
      else await api.createSecurityRule(payload);
      setEditing(null);
      setForm(blank);
      setOk(t("rules.saved"));
      await load();
    } catch (err: any) {
      setError(err.message);
    }
  }

  function appName(id: string) {
    return apps.find((a) => a.id === id)?.name || id || "—";
  }

  return (
    <>
      <div className="topbar">
        <div>
          <h1>{t("rules.title")}</h1>
          <p>{t("rules.description")}</p>
        </div>
      </div>
      {error && <div className="error">{error}</div>}
      {ok && <div className="ok">{ok}</div>}

      <form className="panel form" onSubmit={submit} style={{ marginBottom: 18, maxWidth: 720 }}>
        <h3 style={{ margin: 0 }}>{editing ? t("rules.edit") : t("rules.new")}</h3>
        <label>
          {t("rules.name")}
          <input required value={form.name || ""} onChange={(e) => setForm({ ...form, name: e.target.value })} />
        </label>
        <div className="row">
          <label style={{ flex: 1 }}>
            {t("rules.scope")}
            <select
              value={form.scope || "app"}
              onChange={(e) => setForm({ ...form, scope: e.target.value as "global" | "app" })}
            >
              <option value="app">{t("rules.scopeApp")}</option>
              <option value="global">{t("rules.scopeGlobal")}</option>
            </select>
          </label>
          <label style={{ flex: 1 }}>
            {t("rules.priority")}
            <input
              type="number"
              min={1}
              value={form.priority ?? 100}
              onChange={(e) => setForm({ ...form, priority: Number(e.target.value) || 100 })}
            />
          </label>
        </div>
        {form.scope !== "global" && (
          <label>
            {t("rules.application")}
            <select
              required
              value={form.app_id || ""}
              onChange={(e) => setForm({ ...form, app_id: e.target.value })}
            >
              <option value="">{t("rules.pickApp")}</option>
              {apps.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name} ({a.host})
                </option>
              ))}
            </select>
          </label>
        )}
        <div className="row">
          <label style={{ flex: 1 }}>
            {t("rules.field")}
            <select value={form.field || "uri"} onChange={(e) => setForm({ ...form, field: e.target.value })}>
              <option value="uri">URI</option>
              <option value="path">Path</option>
              <option value="query">Query</option>
              <option value="method">Method</option>
              <option value="host">Host</option>
              <option value="ip">IP</option>
              <option value="user_agent">User-Agent</option>
              <option value="header">Header</option>
            </select>
          </label>
          <label style={{ flex: 1 }}>
            {t("rules.operator")}
            <select value={form.operator || "contains"} onChange={(e) => setForm({ ...form, operator: e.target.value })}>
              <option value="contains">{t("rules.op.contains")}</option>
              <option value="eq">{t("rules.op.eq")}</option>
              <option value="starts_with">{t("rules.op.starts")}</option>
              <option value="ends_with">{t("rules.op.ends")}</option>
              <option value="regex">{t("rules.op.regex")}</option>
              <option value="in">{t("rules.op.in")}</option>
            </select>
          </label>
        </div>
        {form.field === "header" && (
          <label>
            {t("rules.headerName")}
            <input
              required
              placeholder="X-Custom-Header"
              value={form.header_name || ""}
              onChange={(e) => setForm({ ...form, header_name: e.target.value })}
            />
          </label>
        )}
        <label>
          {t("rules.value")}
          <input
            required
            placeholder={form.operator === "in" ? "GET,HEAD" : "/wp-admin"}
            value={form.value || ""}
            onChange={(e) => setForm({ ...form, value: e.target.value })}
          />
        </label>
        <label>
          {t("rules.action")}
          <select value={form.action || "block"} onChange={(e) => setForm({ ...form, action: e.target.value as any })}>
            <option value="block">{t("rules.actions.block")}</option>
            <option value="allow">{t("rules.actions.allow")}</option>
            <option value="log">{t("rules.actions.log")}</option>
          </select>
        </label>
        <label>
          {t("rules.notes")}
          <input value={form.notes || ""} onChange={(e) => setForm({ ...form, notes: e.target.value })} />
        </label>
        <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
          <input
            type="checkbox"
            checked={form.enabled !== false}
            onChange={(e) => setForm({ ...form, enabled: e.target.checked })}
          />
          {t("rules.enabled")}
        </label>
        <div className="row">
          <button className="btn primary" type="submit">
            {t("common.save")}
          </button>
          {editing && (
            <button
              className="btn"
              type="button"
              onClick={() => {
                setEditing(null);
                setForm(blank);
              }}
            >
              {t("common.cancel")}
            </button>
          )}
        </div>
      </form>

      <div className="panel" style={{ overflowX: "auto" }}>
        {rules.length === 0 ? (
          <div className="empty">{t("rules.empty")}</div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>{t("rules.priority")}</th>
                <th>{t("rules.name")}</th>
                <th>{t("rules.scope")}</th>
                <th>{t("rules.match")}</th>
                <th>{t("rules.action")}</th>
                <th>{t("rules.hits")}</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {rules.map((r) => (
                <tr key={r.id}>
                  <td>{r.priority}</td>
                  <td>
                    {r.name}
                    {!r.enabled && <span className="badge off" style={{ marginLeft: 8 }}>{t("rules.disabled")}</span>}
                  </td>
                  <td>{r.scope === "global" ? t("rules.scopeGlobal") : appName(r.app_id)}</td>
                  <td className="mono">
                    {r.field}
                    {r.field === "header" ? `[${r.header_name}]` : ""} {r.operator} {r.value}
                  </td>
                  <td>{r.action}</td>
                  <td>{r.hit_count || 0}</td>
                  <td>
                    <div className="row">
                      <button
                        className="btn"
                        type="button"
                        onClick={() => {
                          setEditing(r.id);
                          setForm({ ...r });
                        }}
                      >
                        {t("common.edit")}
                      </button>
                      <button
                        className="btn danger"
                        type="button"
                        onClick={async () => {
                          if (!confirm(`${t("common.delete")} ${r.name}?`)) return;
                          try {
                            await api.deleteSecurityRule(r.id);
                            await load();
                          } catch (e: any) {
                            setError(e.message);
                          }
                        }}
                      >
                        {t("common.delete")}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </>
  );
}
