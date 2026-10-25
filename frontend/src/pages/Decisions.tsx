import { useEffect, useState } from "react";
import { api } from "../lib/api";

export default function Decisions() {
  const [rows, setRows] = useState<any[]>([]);
  const [warning, setWarning] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    api
      .decisions()
      .then((d) => {
        setRows(d.decisions || []);
        setWarning(d.warning || "");
      })
      .catch((e) => setError(e.message));
  }, []);

  return (
    <>
      <div className="topbar">
        <div>
          <h1>Decisions</h1>
          <p>Live CrowdSec decisions — bans and captchas currently enforced at the edge.</p>
        </div>
      </div>
      {error && <div className="error">{error}</div>}
      {warning && <div className="error" style={{ background: "rgba(180,83,9,.08)", color: "var(--warn)" }}>{warning}</div>}
      <div className="panel">
        {rows.length === 0 ? (
          <div className="empty">No active decisions.</div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Type</th>
                <th>Scope</th>
                <th>Value</th>
                <th>Reason</th>
                <th>Origin</th>
                <th>Until</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((d, i) => (
                <tr key={i}>
                  <td>{d.type || d.Type}</td>
                  <td>{d.scope || d.Scope}</td>
                  <td className="mono">{d.value || d.Value}</td>
                  <td>{d.scenario || d.Scenario || d.reason}</td>
                  <td>{d.origin || d.Origin}</td>
                  <td className="mono">{d.until || d.Until || "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </>
  );
}