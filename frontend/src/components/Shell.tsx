import { NavLink } from "react-router-dom";
import { ReactNode } from "react";
import { BRAND } from "../lib/brand";

export default function Shell({
  user,
  onLogout,
  children,
}: {
  user: string;
  onLogout: () => void;
  children: ReactNode;
}) {
  return (
    <div className="shell">
      <aside className="side">
        <div className="brand">
          <strong>The Gate</strong>
          <span>Independent edge control</span>
        </div>
        <nav className="nav">
          <NavLink to="/" end>
            Applications
          </NavLink>
          <NavLink to="/decisions">Decisions</NavLink>
          <NavLink to="/audit">Audit</NavLink>
          <NavLink to="/settings">Settings</NavLink>
        </nav>
        <div style={{ marginTop: "auto", fontSize: "0.85rem", opacity: 0.75 }}>
          Signed in as <b>{user}</b>
          <div style={{ marginTop: 10 }}>
            <button className="btn" onClick={onLogout} style={{ width: "100%" }}>
              Sign out
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