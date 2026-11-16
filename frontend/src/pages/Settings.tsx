import { useEffect, useState } from "react";
import { api, ErrorTheme, OpsSettings } from "../lib/api";
import { BRAND } from "../lib/brand";
import { useI18n } from "../i18n/I18nProvider";
import { useUiTheme } from "../theme/ThemeProvider";
import { UiThemeId } from "../theme/uiThemes";

const defaults: OpsSettings = {
  auto_update_enabled: false,
  auto_update_hours: 24,
  log_retention_days: 7,
  error_tip_message: "",
  error_theme: "forest",
  telegram: { enabled: false, bot_token: "", chat_id: "" },
  smtp: { enabled: false, host: "", port: 587, user: "", password: "", from: "", to: "" },
};

export default function Settings() {
  const { t, lang } = useI18n();
  const { uiThemeId, themes: uiThemes, setUiThemeId } = useUiTheme();
  const [health, setHealth] = useState<any>(null);
  const [dash, setDash] = useState<any>(null);
  const [prot, setProt] = useState<any>(null);
  const [ops, setOps] = useState<OpsSettings>(defaults);
  const [codes, setCodes] = useState<string[]>([]);
  const [pages, setPages] = useState<Record<string, string>>({});
  const [themes, setThemes] = useState<ErrorTheme[]>([]);
  const [selectedTheme, setSelectedTheme] = useState("forest");
  const [selectedCode, setSelectedCode] = useState("503");
  const [html, setHtml] = useState("");
  const [msg, setMsg] = useState("");
  const [busy, setBusy] = useState("");

  async function reloadPages() {
    const ep = await api.errorPages();
    setCodes(ep.codes || []);
    setPages(ep.pages || {});
    setThemes(ep.themes || []);
    setSelectedTheme(ep.theme || "forest");
    const code = selectedCode || ep.codes?.[0] || "503";
    setSelectedCode(code);
    setHtml(ep.pages?.[code] || "");
  }

  useEffect(() => {
    Promise.all([api.health(), api.dashboard(), api.protectionsStatus(), api.settings(), api.errorPages()]).then(([h, d, p, s, ep]) => {
      setHealth(h);
      setDash(d);
      setProt(p);
      setOps({ ...s, error_theme: s.error_theme || "forest" });
      setCodes(ep.codes || []);
      setPages(ep.pages || {});
      setThemes(ep.themes || []);
      setSelectedTheme(ep.theme || "forest");
      const code = ep.codes?.[0] || "503";
      setSelectedCode(code);
      setHtml(ep.pages?.[code] || "");
    });
  }, []);

  async function saveSettings() {
    setBusy("settings");
    try {
      setOps(await api.updateSettings(ops));
      setMsg(t("settings.settingsSaved"));
    } catch (e: any) {
      setMsg(e.message);
    } finally {
      setBusy("");
    }
  }

  return (
    <>
      <div className="topbar">
        <div>
          <h1>{t("settings.title")}</h1>
          <p>{t("settings.description")}</p>
        </div>
      </div>
      <div className="stats">
        <div className="stat"><em>{t("settings.product")}</em><strong style={{ fontSize: "1.1rem" }}>{dash?.product || "The Gate"}</strong></div>
        <div className="stat"><em>{t("settings.version")}</em><strong style={{ fontSize: "1.1rem" }}>{dash?.version || "..."}</strong></div>
        <div className="stat"><em>{t("settings.api")}</em><strong style={{ fontSize: "1.1rem" }}>{health?.status || "..."}</strong></div>
      </div>

      <div className="panel form" style={{ maxWidth: 820, marginBottom: 18 }}>
        <h3 style={{ margin: 0 }}>{t("settings.operationsSettings")}</h3>
        <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
          <input type="checkbox" checked={ops.auto_update_enabled} onChange={(e) => setOps({ ...ops, auto_update_enabled: e.target.checked })} />
          {t("settings.autoUpdateProtections")}
        </label>
        <label>{t("settings.autoUpdateInterval")}<input type="number" min={1} value={ops.auto_update_hours} onChange={(e) => setOps({ ...ops, auto_update_hours: Number(e.target.value) })} /></label>
        <label>{t("settings.logRetention")}<input type="number" min={1} value={ops.log_retention_days} onChange={(e) => setOps({ ...ops, log_retention_days: Number(e.target.value) })} /></label>
        <label>{t("settings.errorTip")}<input value={ops.error_tip_message || ""} onChange={(e) => setOps({ ...ops, error_tip_message: e.target.value })} /></label>
        <p style={{ margin: 0, color: "var(--muted)", fontSize: "0.85rem" }}>{t("settings.errorTipHelp")}</p>
        <button className="btn primary" disabled={busy === "settings"} onClick={saveSettings}>{busy === "settings" ? t("settings.saving") : t("settings.saveSettings")}</button>
      </div>

      <div className="panel" style={{ display: "grid", gap: 14, maxWidth: 820, marginBottom: 18 }}>
        <h3 style={{ margin: 0 }}>{t("settings.protectionFeeds")}</h3>
        <p style={{ margin: 0, color: "var(--muted)" }}>{t("settings.protectionFeedsDescription")}</p>
        <div className="row">
          <button className="btn accent" disabled={!!busy} onClick={() => { setBusy("crs"); api.updateCRS().then((r) => setMsg(r.message)).catch((e) => setMsg(e.message)).finally(() => setBusy("")); }}>
            {busy === "crs" ? t("settings.updatingCrs") : t("settings.updateCrs")}
          </button>
          <button className="btn accent" disabled={!!busy} onClick={() => { setBusy("cs"); api.updateCrowdSec().then((r) => setMsg(r.message)).catch((e) => setMsg(e.message)).finally(() => setBusy("")); }}>
            {busy === "cs" ? t("settings.updatingCrowdSec") : t("settings.updateCrowdSec")}
          </button>
        </div>
        {prot && (
          <ul style={{ margin: 0, paddingLeft: 18, color: "var(--muted)", lineHeight: 1.6 }}>
            <li>{t("settings.crsDir")}: <span className="mono">{prot.crs_dir}</span></li>
            <li>{t("settings.autoUpdate")}: {ops.auto_update_enabled ? `${t("settings.everyHours")} ${ops.auto_update_hours}h` : t("settings.off")}</li>
            <li>{t("settings.adminOidc")}: {prot.oidc_admin ? t("settings.enabled") : t("settings.localOnly")}</li>
            <li>{t("settings.appForwardAuth")}: <span className="mono">{prot.authentik_forward_url}</span></li>
          </ul>
        )}
      </div>

      <div className="panel form" style={{ maxWidth: 820, marginBottom: 18 }}>
        <h3 style={{ margin: 0 }}>{t("settings.notifications")}</h3>
        <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
          <input type="checkbox" checked={ops.telegram.enabled} onChange={(e) => setOps({ ...ops, telegram: { ...ops.telegram, enabled: e.target.checked } })} />
          {t("settings.telegramEnabled")}
        </label>
        <label>{t("settings.telegramBotToken")}<input value={ops.telegram.bot_token || ""} onChange={(e) => setOps({ ...ops, telegram: { ...ops.telegram, bot_token: e.target.value } })} /></label>
        <label>{t("settings.telegramChatId")}<input value={ops.telegram.chat_id || ""} onChange={(e) => setOps({ ...ops, telegram: { ...ops.telegram, chat_id: e.target.value } })} /></label>
        <label style={{ display: "flex", gap: 8, alignItems: "center" }}>
          <input type="checkbox" checked={ops.smtp.enabled} onChange={(e) => setOps({ ...ops, smtp: { ...ops.smtp, enabled: e.target.checked } })} />
          {t("settings.smtpEnabled")}
        </label>
        <label>{t("settings.smtpHost")}<input value={ops.smtp.host || ""} onChange={(e) => setOps({ ...ops, smtp: { ...ops.smtp, host: e.target.value } })} /></label>
        <label>{t("settings.port")}<input type="number" value={ops.smtp.port || 587} onChange={(e) => setOps({ ...ops, smtp: { ...ops.smtp, port: Number(e.target.value) } })} /></label>
        <label>{t("settings.user")}<input value={ops.smtp.user || ""} onChange={(e) => setOps({ ...ops, smtp: { ...ops.smtp, user: e.target.value } })} /></label>
        <label>{t("settings.password")}<input type="password" value={ops.smtp.password || ""} onChange={(e) => setOps({ ...ops, smtp: { ...ops.smtp, password: e.target.value } })} /></label>
        <label>{t("settings.from")}<input value={ops.smtp.from || ""} onChange={(e) => setOps({ ...ops, smtp: { ...ops.smtp, from: e.target.value } })} /></label>
        <label>{t("settings.to")}<input value={ops.smtp.to || ""} onChange={(e) => setOps({ ...ops, smtp: { ...ops.smtp, to: e.target.value } })} /></label>
        <div className="row">
          <button className="btn primary" onClick={saveSettings}>{t("settings.saveNotificationSettings")}</button>
          <button className="btn" onClick={() => api.testNotifications().then(() => setMsg(t("settings.notificationTestSent"))).catch((e) => setMsg(e.message))}>{t("settings.sendTest")}</button>
        </div>
      </div>

      <div className="panel" style={{ display: "grid", gap: 14, maxWidth: 820, marginBottom: 18 }}>
        <h3 style={{ margin: 0 }}>{t("settings.operations")}</h3>
        <div className="row">
          <button className="btn" onClick={() => api.sync().then(() => setMsg(t("settings.traefikSynced"))).catch((e) => setMsg(e.message))}>{t("settings.forceTraefikSync")}</button>
          <button className="btn" onClick={() => api.exportApps().then((apps) => { const blob = new Blob([JSON.stringify(apps, null, 2)], { type: "application/json" }); const a = document.createElement("a"); a.href = URL.createObjectURL(blob); a.download = "the-gate-apps.json"; a.click(); setMsg(`${t("settings.exportedApplications")} ${apps.length} ${t("settings.applicationsSuffix")}`); }).catch((e) => setMsg(e.message))}>{t("settings.exportApplications")}</button>
          <button className="btn" onClick={() => api.backup().then((blob) => { const a = document.createElement("a"); a.href = URL.createObjectURL(blob); a.download = "the-gate-backup.zip"; a.click(); setMsg(t("settings.backupDownloaded")); }).catch((e) => setMsg(e.message))}>{t("settings.backupZip")}</button>
        </div>
        {msg && <pre className="mono" style={{ whiteSpace: "pre-wrap", margin: 0 }}>{msg}</pre>}
      </div>

      <div className="panel form" style={{ maxWidth: 920, marginBottom: 18 }}>
        <h3 style={{ margin: 0 }}>{t("settings.panelTheme")}</h3>
        <p style={{ margin: 0, color: "var(--muted)" }}>{t("settings.panelThemeHelp")}</p>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill,minmax(140px,1fr))", gap: 10, marginTop: 10 }}>
          {uiThemes.map((theme) => {
            const active = uiThemeId === theme.id;
            const label = lang === "it" ? theme.name_it : theme.name_en;
            return (
              <button
                key={theme.id}
                type="button"
                className={`btn ${active ? "accent" : ""}`}
                style={{ display: "grid", gap: 8, textAlign: "left" }}
                onClick={() => {
                  setUiThemeId(theme.id as UiThemeId);
                  setMsg(`${t("settings.panelThemeApplied")} ${label}`);
                }}
              >
                <span
                  style={{
                    display: "block",
                    height: 28,
                    borderRadius: 8,
                    background: `linear-gradient(135deg, ${theme.preview}, ${theme.vars["--bg"]})`,
                    border: "1px solid var(--line)",
                  }}
                />
                <span style={{ fontSize: "0.85rem" }}>{label}</span>
              </button>
            );
          })}
        </div>
      </div>

      <div className="panel form" style={{ maxWidth: 920, marginBottom: 18 }}>
        <h3 style={{ margin: 0 }}>{t("settings.errorPages")}</h3>
        <p style={{ margin: 0, color: "var(--muted)" }}>{t("settings.errorPagesDescription")}</p>
        <div>
          <strong style={{ display: "block", marginBottom: 8 }}>{t("settings.colorTheme")}</strong>
          <p style={{ margin: "0 0 10px", color: "var(--muted)", fontSize: "0.9rem" }}>{t("settings.colorThemeHelp")}</p>
          <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill,minmax(140px,1fr))", gap: 10 }}>
            {themes.map((theme) => {
              const active = selectedTheme === theme.id;
              const label = lang === "it" ? theme.name_it : theme.name_en;
              return (
                <button
                  key={theme.id}
                  type="button"
                  className={`btn ${active ? "accent" : ""}`}
                  style={{
                    display: "grid",
                    gap: 8,
                    textAlign: "left",
                    borderColor: active ? "var(--brand-2)" : undefined,
                  }}
                  onClick={() => setSelectedTheme(theme.id)}
                >
                  <span
                    style={{
                      display: "block",
                      height: 28,
                      borderRadius: 8,
                      background: `linear-gradient(135deg, ${theme.preview}, ${theme.background})`,
                      border: "1px solid var(--line)",
                    }}
                  />
                  <span style={{ fontSize: "0.85rem" }}>{label}</span>
                </button>
              );
            })}
          </div>
          <div className="row" style={{ marginTop: 12 }}>
            <button
              className="btn primary"
              disabled={busy === "theme"}
              onClick={() => {
                setBusy("theme");
                api
                  .applyErrorTheme(selectedTheme)
                  .then(async () => {
                    setOps({ ...ops, error_theme: selectedTheme });
                    await reloadPages();
                    setMsg(`${t("settings.themeApplied")} ${selectedTheme}`);
                  })
                  .catch((e) => setMsg(e.message))
                  .finally(() => setBusy(""));
              }}
            >
              {busy === "theme" ? t("settings.applyingTheme") : t("settings.applyTheme")}
            </button>
          </div>
        </div>
        <label>
          {t("settings.code")}
          <select value={selectedCode} onChange={(e) => { const code = e.target.value; setSelectedCode(code); setHtml(pages[code] || ""); }}>
            {codes.map((code) => <option key={code} value={code}>{code}</option>)}
          </select>
        </label>
        <textarea rows={14} className="mono" value={html} onChange={(e) => setHtml(e.target.value)} />
        <button className="btn primary" onClick={() => api.updateErrorPage(selectedCode, html).then(() => { setPages({ ...pages, [selectedCode]: html }); setMsg(`${t("settings.saved")} ${selectedCode}.`); }).catch((e) => setMsg(e.message))}>{t("settings.saveTemplate")}</button>
      </div>

      <div className="panel" style={{ maxWidth: 640 }}>
        <h3 style={{ margin: "0 0 8px" }}>{t("settings.licenseCopyright")}</h3>
        <p style={{ margin: 0, color: "var(--muted)", lineHeight: 1.6 }}>
          <strong>{BRAND.product}</strong> - Copyright © 2026 <strong>{BRAND.author}</strong> - <strong>{BRAND.company}</strong>.
          <br />
          {t("settings.website")}: <a href={BRAND.website} target="_blank" rel="noreferrer">{BRAND.website}</a>
          <br />
          {t("settings.licenseText")} <span className="mono">LICENSE</span> {t("settings.licenseAnd")} <span className="mono">NOTICE</span> {t("settings.projectRoot")}
        </p>
      </div>
    </>
  );
}