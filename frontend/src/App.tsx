import { Navigate, Route, Routes } from "react-router-dom";
import { useEffect, useState } from "react";
import { api } from "./lib/api";
import Login from "./pages/Login";
import Shell from "./components/Shell";
import Applications from "./pages/Applications";
import AppLogs from "./pages/AppLogs";
import Decisions from "./pages/Decisions";
import AllowDeny from "./pages/AllowDeny";
import SecurityRules from "./pages/SecurityRules";
import Certificates from "./pages/Certificates";
import Traffic from "./pages/Traffic";
import Audit from "./pages/Audit";
import Settings from "./pages/Settings";
import { useI18n } from "./i18n/I18nProvider";

export default function App() {
  const { t } = useI18n();
  const [user, setUser] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .me()
      .then((m) => setUser(m.username))
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <div className="login-page">{t("common.loading")} The Gate...</div>;
  }

  if (!user) {
    return (
      <Login
        onSuccess={(u) => setUser(u)}
      />
    );
  }

  return (
    <Shell
      user={user}
      onLogout={async () => {
        await api.logout();
        setUser(null);
      }}
    >
      <Routes>
        <Route path="/" element={<Applications />} />
        <Route path="/apps/:id/logs" element={<AppLogs />} />
        <Route path="/allow-deny" element={<AllowDeny />} />
        <Route path="/rules" element={<SecurityRules />} />
        <Route path="/certificates" element={<Certificates />} />
        <Route path="/traffic" element={<Traffic />} />
        <Route path="/decisions" element={<Decisions />} />
        <Route path="/audit" element={<Audit />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Shell>
  );
}