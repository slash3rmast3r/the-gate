import { useEffect, useState } from "react";
import { api, TrafficSummary } from "../lib/api";
import { useI18n } from "../i18n/I18nProvider";

export default function Traffic() {
  const { t } = useI18n();
  const [minutes, setMinutes] = useState(60);
  const [data, setData] = useState<TrafficSummary | null>(null);
  const [error, setError] = useState("");

  async function load(next = minutes) {
    setError("");
    try {
      setData(await api.traffic(next));
    } catch (e: any) {
      setError(e.message);
    }
  }

  useEffect(() => {
    load();
    const t = setInterval(() => load(), 10000);
    return () => clearInterval(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <>
      <div className="topbar">
        <div>
          <h1>{t("traffic.title")}</h1>
          <p>{t("traffic.description")}</p>
        </div>
        <div className="row">
          <select
            value={minutes}
            onChange={(e) => {
              const n = Number(e.target.value);
              setMinutes(n);
              load(n);
            }}
          >
            <option value={15}>15m</option>
            <option value={60}>1h</option>
            <option value={360}>6h</option>
            <option value={1440}>24h</option>
          </select>
          <button className="btn" onClick={() => load()}>{t("common.refresh")}</button>
        </div>
      </div>
      {error && <div className="error">{error}</div>}
      <div className="stats">
        <div className="stat"><em>{t("traffic.requests")}</em><strong>{data?.requests ?? "..."}</strong></div>
        <div className="stat"><em>{t("traffic.blocked")}</em><strong>{data?.blocked ?? "..."}</strong></div>
        <div className="stat"><em>{t("traffic.errors5xx")}</em><strong>{data?.errors ?? "..."}</strong></div>
      </div>
      <div className="panel" style={{ marginBottom: 18 }}>
        <h3 style={{ marginTop: 0 }}>{t("traffic.applications")}</h3>
        {!data || (data.applications || []).length === 0 ? (
          <div className="empty">{t("traffic.noAccessLogs")}</div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>{t("traffic.application")}</th>
                <th>{t("traffic.requests")}</th>
                <th>{t("traffic.blocked")}</th>
                <th>{t("traffic.errors")}</th>
                <th>{t("traffic.clients")}</th>
                <th>{t("traffic.topCountry")}</th>
              </tr>
            </thead>
            <tbody>
              {(data.applications || []).map((a) => (
                <tr key={a.application_id}>
                  <td>{a.name}<div className="mono">{a.host}</div></td>
                  <td>{a.requests}</td>
                  <td>{a.blocked}</td>
                  <td>{a.errors}</td>
                  <td>{a.unique_clients}</td>
                  <td>{a.top_country}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
      <div className="panel" style={{ marginBottom: 18 }}>
        <h3 style={{ marginTop: 0 }}>{t("traffic.countries")}</h3>
        {!data || Object.keys(data.countries || {}).length === 0 ? (
          <div className="empty">{t("traffic.noGeoData")}</div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>{t("traffic.country")}</th>
                <th>{t("traffic.requests")}</th>
                <th>{t("traffic.share")}</th>
              </tr>
            </thead>
            <tbody>
              {Object.entries(data.countries || {})
                .sort((a, b) => b[1] - a[1])
                .map(([country, count]) => {
                  const pct = data.requests ? Math.round((count * 1000) / data.requests) / 10 : 0;
                  return (
                    <tr key={country}>
                      <td>{country}</td>
                      <td>{count}</td>
                      <td>
                        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                          <div style={{ flex: 1, height: 8, background: "var(--line)", borderRadius: 99, overflow: "hidden" }}>
                            <div style={{ width: `${Math.min(pct, 100)}%`, height: "100%", background: "var(--brand-2)" }} />
                          </div>
                          <span className="mono">{pct}%</span>
                        </div>
                      </td>
                    </tr>
                  );
                })}
            </tbody>
          </table>
        )}
      </div>
      <div className="panel">
        <h3 style={{ marginTop: 0 }}>{t("traffic.topClients")}</h3>
        {!data || (data.clients || []).length === 0 ? (
          <div className="empty">{t("traffic.noClientData")}</div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>{t("traffic.ip")}</th>
                <th>{t("traffic.country")}</th>
                <th>{t("traffic.requests")}</th>
                <th>{t("traffic.blocked")}</th>
                <th>{t("traffic.errors")}</th>
              </tr>
            </thead>
            <tbody>
              {(data.clients || []).map((c) => (
                <tr key={c.ip}>
                  <td className="mono">{c.ip}</td>
                  <td>{c.country}</td>
                  <td>{c.requests}</td>
                  <td>{c.blocked}</td>
                  <td>{c.errors}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </>
  );
}
