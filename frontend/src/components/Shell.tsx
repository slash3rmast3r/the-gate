import { NavLink } from "react-router-dom";
import { ReactNode } from "react";
import { BRAND } from "../lib/brand";
import { Lang } from "../i18n/translations";
import { useI18n } from "../i18n/I18nProvider";

export default function Shell({
  user,
  onLogout,
  children,
}: {
  user: string;
  onLogout: () => void;
  children: ReactNode;
}) {
  const { lang, setLang, t } = useI18n();
  const chooseLang = (next: Lang) => {
    if (next !== lang) setLang(next);
  };

  return (
    <div className="shell">
      <aside className="side">
        <div className="brand">
          <strong>The Gate</strong>
          <span>{t("shell.tagline")}</span>
        </div>
        <nav className="nav">
          <NavLink to="/" end>
            {t("shell.nav.applications")}
          </NavLink>
          <NavLink to="/allow-deny">{t("shell.nav.allowDeny")}</NavLink>
          <NavLink to="/rules">{t("shell.nav.rules")}</NavLink>
          <NavLink to="/certificates">{t("shell.nav.certificates")}</NavLink>
          <NavLink to="/traffic">{t("shell.nav.traffic")}</NavLink>
          <NavLink to="/decisions">{t("shell.nav.decisions")}</NavLink>
          <NavLink to="/audit">{t("shell.nav.audit")}</NavLink>
          <NavLink to="/settings">{t("shell.nav.settings")}</NavLink>
        </nav>
        <div style={{ marginTop: "auto", fontSize: "0.85rem", opacity: 0.75 }}>
          {t("shell.signedInAs")} <b>{user}</b>
          <div style={{ marginTop: 10, display: "flex", gap: 6, alignItems: "center" }} aria-label={t("shell.language")}>
            <button className={`btn ${lang === "en" ? "accent" : ""}`} onClick={() => chooseLang("en")} type="button">
              EN
            </button>
            <button className={`btn ${lang === "it" ? "accent" : ""}`} onClick={() => chooseLang("it")} type="button">
              IT
            </button>
          </div>
          <div style={{ marginTop: 10 }}>
            <button className="btn" onClick={onLogout} style={{ width: "100%" }}>
              {t("shell.signOut")}
            </button>
          </div>
          <div style={{ marginTop: 14, fontSize: "0.72rem", lineHeight: 1.4, opacity: 0.85 }}>
            © 2026 {BRAND.author}
            <br />
            {BRAND.company} · {BRAND.license}
            <br />
            <a href={BRAND.website} target="_blank" rel="noreferrer" style={{ color: "inherit", textDecoration: "underline" }}>
              {BRAND.websiteHost}
            </a>
          </div>
        </div>
      </aside>
      <main className="main">{children}</main>
    </div>
  );
}