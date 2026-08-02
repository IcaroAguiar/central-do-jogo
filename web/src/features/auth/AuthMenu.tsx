import { useEffect, useState } from "react";
import { type AuthMeResponse, fetchAuthMe, logoutAuth } from "../../api/client";

/** Compact login/logout control for the app header (TASK-027). */
export function AuthMenu() {
  const [me, setMe] = useState<AuthMeResponse | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      try {
        const body = await fetchAuthMe();
        if (!cancelled) {
          setMe(body);
        }
      } catch {
        // Public content must keep working when auth is unavailable (RISK-008).
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  if (!me?.authEnabled) {
    return null;
  }

  if (!me.authenticated) {
    return (
      <a className="auth-link" href="/api/v1/auth/google/start">
        Entrar
      </a>
    );
  }

  async function logout() {
    setBusy(true);
    try {
      await logoutAuth();
      setMe({ authenticated: false, authEnabled: true });
    } finally {
      setBusy(false);
    }
  }

  const label = me.displayName || me.email || "Conta";
  return (
    <div className="auth-menu">
      <span className="auth-user" title={me.email}>
        {label}
        {me.role === "maintainer" ? " · mantenedor" : ""}
      </span>
      <button type="button" className="auth-link" onClick={() => void logout()} disabled={busy}>
        Sair
      </button>
    </div>
  );
}
