import { useEffect, useState } from "react";

type AuthMe = {
  authenticated: boolean;
  authEnabled: boolean;
  email?: string;
  displayName?: string;
  role?: "user" | "maintainer";
};

/** Compact login/logout control for the app header (TASK-027). */
export function AuthMenu() {
  const [me, setMe] = useState<AuthMe | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      try {
        const res = await fetch("/api/v1/auth/me", { credentials: "same-origin" });
        if (!res.ok) {
          return;
        }
        const body = (await res.json()) as AuthMe;
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
      await fetch("/api/v1/auth/logout", {
        method: "POST",
        credentials: "same-origin",
        headers: { Origin: window.location.origin },
      });
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
