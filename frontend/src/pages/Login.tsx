import { FormEvent, useEffect, useState } from "react";
import { api } from "../lib/api";
import { BRAND } from "../lib/brand";
import { useI18n } from "../i18n/I18nProvider";
import { Lang } from "../i18n/translations";

export default function Login({ onSuccess }: { onSuccess: (u: string) => void }) {
  const { lang, setLang, t } = useI18n();
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [opts, setOpts] = useState<{ local: boolean; oidc: boolean }>({ local: true, oidc: false });

  useEffect(() => {
    api.authOptions().then(setOpts).catch(() => undefined);
  }, []);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      const res = await api.login(username, password);
      onSuccess(res.username);
    } catch (err: any) {
      setError(err.message || t("login.failed"));
    } finally {
      setBusy(false);
    }
  }

  const chooseLang = (next: Lang) => {
    if (next !== lang) setLang(next);
  };

  return (
    <div className="login-page">
      <form className="login-card form" onSubmit={submit}>
        <div className="mark">The Gate</div>
        <div className="sub">{t("login.tagline")}</div>
        <div className="row" aria-label={t("login.language")} style={{ justifyContent: "flex-end" }}>
          <button className={`btn ${lang === "en" ? "accent" : ""}`} onClick={() => chooseLang("en")} type="button">
            EN
          </button>
          <button className={`btn ${lang === "it" ? "accent" : ""}`} onClick={() => chooseLang("it")} type="button">
            IT
          </button>
        </div>
        {error && <div className="error">{error}</div>}
        {opts.local && (
          <>
            <label>
              {t("login.username")}
              <input value={username} onChange={(e) => setUsername(e.target.value)} autoComplete="username" />
            </label>
            <label>
              {t("login.password")}
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete="current-password"
              />
            </label>
            <button className="btn primary" disabled={busy} type="submit">
              {busy ? t("login.signingIn") : t("login.enter")}
            </button>
          </>
        )}
        {opts.oidc && (
          <a className="btn accent" href="/api/auth/oidc/login" style={{ textAlign: "center", display: "block" }}>
            {t("login.oidc")}
          </a>
        )}
        {!opts.local && !opts.oidc && <div className="error">{t("login.noMethod")}</div>}
        <p className="sub" style={{ marginTop: 18, marginBottom: 0, fontSize: "0.78rem" }}>
          {BRAND.shortCredit}
          <br />
          <a href={BRAND.website} target="_blank" rel="noreferrer" style={{ color: "var(--brand-2)" }}>
            {BRAND.websiteHost}
          </a>
        </p>
      </form>
    </div>
  );
}