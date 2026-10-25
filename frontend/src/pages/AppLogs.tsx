import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { api, Application } from "../lib/api";

type LogEntry = {
  time?: string;
  kind?: string;
  host?: string;
  client?: string;
  method?: string;
  path?: string;
  status?: number;
  duration?: string;
  message?: string;
  raw?: string;
};

export default function AppLogs() {
  const { id } = useParams<{ id: string }>();
  const [app, setApp] = useState<Application | null>(null);
  const [kind, setKind] = useState<"access" | "error">("access");
  const [q, setQ] = useState("");
  const [entries, setEntries] = useState<LogEntry[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  async function load(nextKind = kind, nextQ = q) {
    if (!id) return;
    setLoading(true);
    setError("");
    try {
      const a = app || (await api.getApp(id));
      if (!app) setApp(a);
      const res = await api.appLogs(id, { kind: nextKind, limit: 200, q: nextQ });
      setEntries(res.entries || []);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
    const t = setInterval(() => load(), 8000);
    return () => clearInterval(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  return (
    <>
      <div className="topbar">
        <div>
          <h1>Logs — {app?.name || "…"}</h1>
          <p>
            Traefik access / error filtered by host <span className="mono">{app?.host || id}</span>. Auto-refresh 8s.
          </p>
        </div>
        <div className="row">
          <Link className="btn" to="/">
            ← Applications
          </Link>
          <button className="btn" onClick={() => load()} disabled={loading}>
            Refresh
          </button>
        </div>
      </div>

      <div className="panel" style={{ marginBottom: 16, display: "flex", gap: 12, flexWrap: "wrap", alignItems: "end" }}>
        <label style={{ display: "grid", gap: 6 }}>
          Kind
          <select
            value={kind}
            onChange={(e) => {
              const k = e.target.value as "access" | "error";
              setKind(k);
              load(k, q);
            }}
          >
            <option value="access">Access (all requests)</option>
            <option value="error">Error (Traefik log + host match)</option>
          </select>
        </label>
        <label style={{ display: "grid", gap: 6, flex: 1, minWidth: 180 }}>
          Filter
          <input
            value={q}
            placeholder="path, IP, status…"
            onChange={(e) => setQ(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") load(kind, q);
            }}
          />
        </label>
        <button className="btn primary" onClick={() => load(kind, q)}>
          Apply
        </button>
      </div>

      {error && <div className="error">{error}</div>}

      <div className="panel">
        {loading && entries.length === 0 ? (
          <div className="empty">Loading logs…</div>
        ) : entries.length === 0 ? (
          <div className="empty">
            No matching entries yet. Generate traffic to the app host, or switch kind.
            {kind === "access" && " Access log is JSON from Traefik (format enabled in traefik.yml)."}
          </div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Time</th>
                {kind === "access" ? (
                  <>
                    <th>Client</th>
                    <th>Method</th>
                    <th>Path</th>
                    <th>Status</th>
                    <th>Duration</th>
                  </>
                ) : (
                  <th>Message</th>
                )}
              </tr>
            </thead>
            <tbody>
              {entries.map((e, i) => (
                <tr key={i}>
                  <td className="mono">{e.time || "—"}</td>
                  {kind === "access" ? (
                    <>
                      <td className="mono">{e.client || "—"}</td>
                      <td>{e.method || "—"}</td>
                      <td className="mono" style={{ maxWidth: 320, wordBreak: "break-all" }}>
                        {e.path || "—"}
                      </td>
                      <td>
                        <span className={`badge ${e.status && e.status >= 400 ? "off" : ""}`}>{e.status || "—"}</span>
                      </td>
                      <td className="mono">{e.duration || "—"}</td>
                    </>
                  ) : (
                    <td className="mono" style={{ whiteSpace: "pre-wrap" }}>
                      {e.message || e.raw}
                    </td>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </>
  );
}