import { FormEvent, useEffect, useState } from "react";
import { api, IPGroup } from "../lib/api";
import { useI18n } from "../i18n/I18nProvider";

const blank: Partial<IPGroup> = { name: "", kind: "allow", cidrs: "", notes: "" };

export default function AllowDeny() {
  const { t } = useI18n();
  const [groups, setGroups] = useState<IPGroup[]>([]);
  const [form, setForm] = useState<Partial<IPGroup>>(blank);
  const [editing, setEditing] = useState<string | null>(null);
  const [error, setError] = useState("");

  async function load() {
    setGroups(await api.ipGroups());
  }

  useEffect(() => {
    load().catch((e) => setError(e.message));
  }, []);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setError("");
    try {
      if (editing) await api.updateIPGroup(editing, form);
      else await api.createIPGroup(form);
      setEditing(null);
      setForm(blank);
      await load();
    } catch (err: any) {
      setError(err.message);
    }
  }

  return (
    <>
      <div className="topbar">
        <div>
          <h1>{t("allowDeny.title")}</h1>
          <p>{t("allowDeny.description")}</p>
        </div>
      </div>
      {error && <div className="error">{error}</div>}
      <form className="panel form" onSubmit={submit} style={{ marginBottom: 18 }}>
        <h3 style={{ margin: 0 }}>{editing ? t("allowDeny.edit") : t("allowDeny.new")}</h3>
        <label>
          {t("allowDeny.name")}
          <input required value={form.name || ""} onChange={(e) => setForm({ ...form, name: e.target.value })} />
        </label>
        <label>
          {t("allowDeny.kind")}
          <select value={form.kind || "allow"} onChange={(e) => setForm({ ...form, kind: e.target.value as "allow" | "deny" })}>
            <option value="allow">{t("allowDeny.allow")}</option>
            <option value="deny">{t("allowDeny.deny")}</option>
          </select>
        </label>
        <label>
          {t("allowDeny.ipsCidrs")}
          <textarea
            required
            rows={4}
            placeholder="203.0.113.10&#10;198.51.100.0/24"
            value={form.cidrs || ""}
            onChange={(e) => setForm({ ...form, cidrs: e.target.value })}
          />
        </label>
        <label>
          {t("allowDeny.notes")}
          <input value={form.notes || ""} onChange={(e) => setForm({ ...form, notes: e.target.value })} />
        </label>
        <div className="row">
          <button className="btn primary" type="submit">{t("allowDeny.saveGroup")}</button>
          {editing && <button className="btn" type="button" onClick={() => { setEditing(null); setForm(blank); }}>{t("common.cancel")}</button>}
        </div>
      </form>

      <div className="grid">
        {groups.map((g) => (
          <article className="card" key={g.id}>
            <header>
              <h3>{g.name}</h3>
              <span className={`badge ${g.kind === "deny" ? "off" : ""}`}>{t(`allowDeny.${g.kind}`)}</span>
            </header>
            <pre className="mono" style={{ whiteSpace: "pre-wrap", margin: 0 }}>{g.cidrs}</pre>
            <div className="meta">{t("allowDeny.hitCount")}: <strong>{g.hit_count ?? 0}</strong></div>
            {g.notes && <p style={{ color: "var(--muted)", margin: 0 }}>{g.notes}</p>}
            <div className="row">
              <button className="btn" onClick={() => { setEditing(g.id); setForm(g); }}>{t("allowDeny.editAction")}</button>
              <button
                className="btn danger"
                onClick={() => {
                  if (confirm(`${t("allowDeny.deleteConfirm")} ${g.name}?`)) api.deleteIPGroup(g.id).then(load).catch((e) => setError(e.message));
                }}
              >
                {t("common.delete")}
              </button>
            </div>
          </article>
        ))}
      </div>
      {groups.length === 0 && <div className="empty">{t("allowDeny.noGroups")}</div>}
    </>
  );
}
