import { useEffect, useState } from "react";
import { api } from "../lib/api";

export default function Audit() {
  const [rows, setRows] = useState<any[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    api
      .audit()
      .then(setRows)
      .catch((e) => setError(e.message));
  }, []);

  return (
    <>
      <div className="topbar">
        <div>
          <h1>Audit</h1>
          <p>Immutable control-plane actions — who changed what, and when.</p>
        </div>
      </div>
      {error && <div className="error">{error}</div>}
      <div className="panel">
        {rows.length === 0 ? (
          <div className="empty">No audit events yet.</div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>When</th>
                <th>Actor</th>
                <th>Action</th>
                <th>Target</th>
                <th>Detail</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((r) => (
                <tr key={r.id}>
                  <td className="mono">{r.created_at}</td>
                  <td>{r.actor}</td>
                  <td>{r.action}</td>
                  <td className="mono">{r.target}</td>
                  <td>{r.detail}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </>
  );
}