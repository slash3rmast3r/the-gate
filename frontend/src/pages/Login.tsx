import { FormEvent, useEffect, useState } from "react";
import { api } from "../lib/api";
import { BRAND } from "../lib/brand";

export default function Login({ onSuccess }: { onSuccess: (u: string) => void }) {
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
      setError(err.message || "Login failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="login-page">
      <form className="login-card form" onSubmit={submit}>
        <div className="mark">The Gate</div>
        <div className="sub">Protect and route your applications — independent of vendor lock-in.</div>
        {error && <div className="error">{error}</div>}
        {opts.local && (
          <>
            <label>
              Username
              <input value={username} onChange={(e) => setUsername(e.target.value)} autoComplete="username" />
            </label>
            <label>
              Password
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete="current-password"
              />
            </label>
            <button className="btn primary" disabled={busy} type="submit">
              {busy ? "Signing in…" : "Enter"}
            </button>
          </>
        )}
        {opts.oidc && (
          <a className="btn accent" href="/api/auth/oidc/login" style={{ textAlign: "center", display: "block" }}>
            Sign in with Authentik / OIDC
          </a>
        )}
        {!opts.local && !opts.oidc && <div className="error">No login method configured.</div>}
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