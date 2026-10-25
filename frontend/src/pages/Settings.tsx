import { useEffect, useState } from "react";
import { api } from "../lib/api";
import { BRAND } from "../lib/brand";

export default function Settings() {
  const [health, setHealth] = useState<any>(null);
  const [dash, setDash] = useState<any>(null);
  const [prot, setProt] = useState<any>(null);
  const [msg, setMsg] = useState("");
  const [busy, setBusy] = useState("");

  useEffect(() => {
    Promise.all([api.health(), api.dashboard(), api.protectionsStatus()]).then(([h, d, p]) => {
      setHealth(h);
      setDash(d);
      setProt(p);
    });
  }, []);

  return (
    <>
      <div className="topbar">
        <div>
          <h1>Settings</h1>
          <p>Stack health, protection feeds, sync, and exports.</p>
        </div>
      </div>
      <div className="stats">
        <div className="stat">
          <em>Product</em>
          <strong style={{ fontSize: "1.1rem" }}>{dash?.product || "The Gate"}</strong>
        </div>
        <div className="stat">
          <em>Version</em>
          <strong style={{ fontSize: "1.1rem" }}>{dash?.version || "—"}</strong>
        </div>
        <div className="stat">
          <em>API</em>
          <strong style={{ fontSize: "1.1rem" }}>{health?.status || "…"}</strong>
        </div>
      </div>

      <div className="panel" style={{ display: "grid", gap: 14, maxWidth: 640, marginBottom: 18 }}>
        <h3 style={{ margin: 0 }}>Protection feeds</h3>
        <p style={{ margin: 0, color: "var(--muted)" }}>
          Update free rule sources on demand. Optional auto-update via{" "}
          <span className="mono">GATE_AUTO_UPDATE_PROTECTIONS=true</span>.
        </p>
        <div className="row">
          <button
            className="btn accent"
            disabled={!!busy}
            onClick={() => {
              setBusy("crs");
              api
                .updateCRS()
                .then((r) => setMsg(r.message))
                .catch((e) => setMsg(e.message))
                .finally(() => setBusy(""));
            }}
          >
            {busy === "crs" ? "Updating CRS…" : "Update OWASP CRS"}
          </button>
          <button
            className="btn accent"
            disabled={!!busy}
            onClick={() => {
              setBusy("cs");
              api
                .updateCrowdSec()
                .then((r) => setMsg(r.message))
                .catch((e) => setMsg(e.message))
                .finally(() => setBusy(""));
            }}
          >
            {busy === "cs" ? "Updating CrowdSec…" : "Update CrowdSec Hub"}
          </button>
        </div>
        {prot && (
          <ul style={{ margin: 0, paddingLeft: 18, color: "var(--muted)", lineHeight: 1.6 }}>
            <li>CRS dir: <span className="mono">{prot.crs_dir}</span></li>
            <li>Auto-update: {prot.auto_update ? `every ${prot.auto_update_hours}h` : "off"}</li>
            <li>Admin OIDC: {prot.oidc_admin ? "enabled" : "local only"}</li>
            <li>App forward-auth: <span className="mono">{prot.authentik_forward_url}</span></li>
            <li>
              CrowdSec Hub from UI needs optional compose file{" "}
              <span className="mono">docker-compose.hub-update.yml</span> (docker.sock)
            </li>
          </ul>
        )}
      </div>

      <div className="panel" style={{ display: "grid", gap: 14, maxWidth: 640 }}>
        <h3 style={{ margin: 0 }}>Operations</h3>
        <div className="row">
          <button
            className="btn"
            onClick={() =>
              api
                .sync()
                .then(() => setMsg("Traefik config synced."))
                .catch((e) => setMsg(e.message))
            }
          >
            Force Traefik sync
          </button>
          <button
            className="btn"
            onClick={() =>
              api
                .exportApps()
                .then((apps) => {
                  const blob = new Blob([JSON.stringify(apps, null, 2)], { type: "application/json" });
                  const a = document.createElement("a");
                  a.href = URL.createObjectURL(blob);
                  a.download = "the-gate-apps.json";
                  a.click();
                  setMsg(`Exported ${apps.length} application(s).`);
                })
                .catch((e) => setMsg(e.message))
            }
          >
            Export applications
          </button>
        </div>
        {msg && <pre className="mono" style={{ whiteSpace: "pre-wrap", margin: 0 }}>{msg}</pre>}
        <ul style={{ margin: 0, paddingLeft: 18, color: "var(--muted)", lineHeight: 1.6 }}>
          <li>Customize error HTML in <span className="mono">errorpages/</span></li>
          <li>Authentik SSO for admin: set <span className="mono">GATE_AUTH_MODE</span> + OIDC_* in .env</li>
          <li>Per-app Authentik: Auth mode → Authentik SSO (needs outpost Traefik)</li>
        </ul>
      </div>

      <div className="panel" style={{ marginTop: 18, maxWidth: 640 }}>
        <h3 style={{ margin: "0 0 8px" }}>License &amp; copyright</h3>
        <p style={{ margin: 0, color: "var(--muted)", lineHeight: 1.6 }}>
          <strong>{BRAND.product}</strong> — Copyright © 2026 <strong>{BRAND.author}</strong> —{" "}
          <strong>{BRAND.company}</strong>.
          <br />
          Website:{" "}
          <a href={BRAND.website} target="_blank" rel="noreferrer">
            {BRAND.website}
          </a>
          <br />
          Licensed under the <strong>MIT License</strong>. See <span className="mono">LICENSE</span> and{" "}
          <span className="mono">NOTICE</span> in the project root.
        </p>
      </div>
    </>
  );
}