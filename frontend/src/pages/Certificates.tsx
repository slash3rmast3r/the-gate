import { useEffect, useState } from "react";
import { api, CertificateInfo } from "../lib/api";
import { useI18n } from "../i18n/I18nProvider";

export default function Certificates() {
  const { t } = useI18n();
  const [certs, setCerts] = useState<CertificateInfo[]>([]);
  const [note, setNote] = useState("");
  const [error, setError] = useState("");
  const [ok, setOk] = useState("");
  const [busy, setBusy] = useState("");

  async function load() {
    const res = await api.certificates();
    setCerts(res.certificates || []);
    setNote(res.note || "");
  }

  useEffect(() => {
    load().catch((e) => setError(e.message));
  }, []);

  return (
    <>
      <div className="topbar">
        <div>
          <h1>{t("certs.title")}</h1>
          <p>{t("certs.description")}</p>
        </div>
        <button
          className="btn"
          type="button"
          onClick={() => {
            setError("");
            setOk("");
            load().catch((e) => setError(e.message));
          }}
        >
          {t("common.refresh")}
        </button>
      </div>
      {error && <div className="error">{error}</div>}
      {ok && <div className="ok">{ok}</div>}
      {note && <p style={{ color: "var(--muted)", marginTop: 0 }}>{note}</p>}
      <p style={{ color: "var(--muted)" }}>{t("certs.extraHostsHelp")}</p>

      {certs.length === 0 ? (
        <div className="empty">{t("certs.empty")}</div>
      ) : (
        <div className="grid">
          {certs.map((c) => (
            <article className="card" key={`${c.resolver}-${c.main}`}>
              <header>
                <h3>{c.main}</h3>
                <span className={`badge ${c.days_left < 21 ? "off" : ""}`}>
                  {c.days_left} {t("certs.daysLeft")}
                </span>
              </header>
              <div className="meta">
                {c.app_name ? `${c.app_name} · ${c.app_host}` : t("certs.unlinked")}
              </div>
              {c.sans?.length > 0 && (
                <div className="meta">SAN: {c.sans.join(", ")}</div>
              )}
              <div className="meta">
                {t("certs.issuer")}: {c.issuer || "—"}
              </div>
              <div className="meta">
                {t("certs.expires")}: {c.not_after ? new Date(c.not_after).toLocaleString() : "—"}
              </div>
              <div className="row">
                <button
                  className="btn"
                  type="button"
                  disabled={busy === c.main}
                  onClick={async () => {
                    if (!confirm(t("certs.renewConfirm"))) return;
                    setBusy(c.main);
                    setError("");
                    setOk("");
                    try {
                      const res = await api.renewCertificate(c.main);
                      setOk(res.message || t("certs.renewOk"));
                      await load();
                    } catch (e: any) {
                      setError(e.message);
                    } finally {
                      setBusy("");
                    }
                  }}
                >
                  {busy === c.main ? t("certs.renewing") : t("certs.forceRenew")}
                </button>
              </div>
            </article>
          ))}
        </div>
      )}
    </>
  );
}
