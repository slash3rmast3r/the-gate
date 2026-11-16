import { useEffect, useState } from "react";
import { api } from "../lib/api";
import { useI18n } from "../i18n/I18nProvider";

export default function Audit() {
  const { t } = useI18n();
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
          <h1>{t("audit.title")}</h1>
          <p>{t("audit.description")}</p>
        </div>
      </div>
      {error && <div className="error">{error}</div>}
      <div className="panel">
        {rows.length === 0 ? (
          <div className="empty">{t("audit.noEvents")}</div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>{t("audit.when")}</th>
                <th>{t("audit.actor")}</th>
                <th>{t("audit.action")}</th>
                <th>{t("audit.target")}</th>
                <th>{t("audit.detail")}</th>
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