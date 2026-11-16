import { useEffect, useState } from "react";
import { api } from "../lib/api";
import { useI18n } from "../i18n/I18nProvider";

export default function Decisions() {
  const { t } = useI18n();
  const [rows, setRows] = useState<any[]>([]);
  const [warning, setWarning] = useState("");
  const [error, setError] = useState("");

  function load() {
    api
      .decisions()
      .then((d) => {
        setRows(d.decisions || []);
        setWarning(d.warning || "");
      })
      .catch((e) => setError(e.message));
  }

  useEffect(() => {
    load();
  }, []);

  return (
    <>
      <div className="topbar">
        <div>
          <h1>{t("decisions.title")}</h1>
          <p>{t("decisions.description")}</p>
        </div>
      </div>
      {error && <div className="error">{error}</div>}
      {warning && <div className="error" style={{ background: "rgba(180,83,9,.08)", color: "var(--warn)" }}>{warning}</div>}
      <div className="panel">
        {rows.length === 0 ? (
          <div className="empty">{t("decisions.noActive")}</div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>{t("decisions.type")}</th>
                <th>{t("decisions.scope")}</th>
                <th>{t("decisions.value")}</th>
                <th>{t("decisions.reason")}</th>
                <th>{t("decisions.origin")}</th>
                <th>{t("decisions.until")}</th>
                <th></th>
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
                  <td>
                    <button
                      className="btn"
                      onClick={() =>
                        api
                          .deleteDecision(d.scope || d.Scope, d.value || d.Value)
                          .then(load)
                          .catch((e) => setError(e.message))
                      }
                    >
                      {t("decisions.unblock")}
                    </button>
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