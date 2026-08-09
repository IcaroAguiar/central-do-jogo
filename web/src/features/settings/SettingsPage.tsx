import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import {
  type AuthMeResponse,
  deletePrivacyAccount,
  fetchAuthMe,
  fetchPrivacyExport,
} from "../../api/client";
import { markAuthenticated } from "../preferences/sync";

/** Account privacy settings: export JSON and self-serve delete (REQ-019). */
export function SettingsPage() {
  const navigate = useNavigate();
  const [me, setMe] = useState<AuthMeResponse | null>(null);
  const [busy, setBusy] = useState<"export" | "delete" | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      try {
        const body = await fetchAuthMe();
        if (!cancelled) {
          setMe(body);
        }
      } catch {
        if (!cancelled) {
          setError("Não foi possível carregar a conta.");
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  if (me && (!me.authEnabled || !me.authenticated)) {
    return (
      <section className="settings-page">
        <h1>Conta</h1>
        <p className="lede">Entre para exportar ou excluir seus dados.</p>
        <p>
          <a className="auth-link" href="/api/v1/auth/google/start">
            Entrar
          </a>
        </p>
      </section>
    );
  }

  async function onExport() {
    setBusy("export");
    setError(null);
    try {
      const data = await fetchPrivacyExport();
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `central-do-jogo-export-${data.exportedAt.slice(0, 10)}.json`;
      a.click();
      URL.revokeObjectURL(url);
    } catch {
      setError("Falha ao exportar dados.");
    } finally {
      setBusy(null);
    }
  }

  async function onDelete() {
    const ok = window.confirm(
      "Excluir permanentemente sua conta e dados associados? Esta ação não pode ser desfeita.",
    );
    if (!ok) {
      return;
    }
    setBusy("delete");
    setError(null);
    try {
      await deletePrivacyAccount();
      markAuthenticated(false);
      navigate("/", { replace: true });
    } catch {
      setError("Falha ao excluir a conta.");
      setBusy(null);
    }
  }

  const label = me?.displayName || me?.email || "Conta";

  return (
    <section className="settings-page">
      <h1>Conta</h1>
      <p className="lede">
        Gerencie a exportação e a exclusão dos dados da sua conta ({label}).
      </p>
      {error ? (
        <p className="settings-error" role="alert">
          {error}
        </p>
      ) : null}
      <div className="settings-actions">
        <button type="button" onClick={() => void onExport()} disabled={busy !== null || !me}>
          {busy === "export" ? "Exportando…" : "Baixar meus dados (JSON)"}
        </button>
        <button
          type="button"
          className="settings-danger"
          onClick={() => void onDelete()}
          disabled={busy !== null || !me}
        >
          {busy === "delete" ? "Excluindo…" : "Excluir conta"}
        </button>
      </div>
      <p className="note">
        A exportação não inclui tokens de sessão, cookies nem endpoints de Push.{" "}
        <Link to="/">Voltar ao início</Link>
      </p>
    </section>
  );
}
