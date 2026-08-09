import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import {
  type AdminAtRiskMatch,
  type AdminAuditEvent,
  type AdminReportItem,
  type AdminSourceHealthItem,
  type AuthMeResponse,
  fetchAdminAtRiskMatches,
  fetchAdminAudit,
  fetchAdminReports,
  fetchAdminSourceHealth,
  fetchAuthMe,
  postAdminMatchAction,
  reviewAdminReport,
} from "../../api/client";

async function loadPanelData() {
  const [health, risk, audit, openReports] = await Promise.all([
    fetchAdminSourceHealth(),
    fetchAdminAtRiskMatches(),
    fetchAdminAudit(),
    fetchAdminReports(),
  ]);
  return {
    sources: health.sources ?? [],
    matches: risk.matches ?? [],
    events: audit.events ?? [],
    reports: openReports.reports ?? [],
  };
}

/** Maintainer panel: source health, at-risk matches, audit (TASK-031). */
export function AdminPage() {
  const [me, setMe] = useState<AuthMeResponse | null>(null);
  const [sources, setSources] = useState<AdminSourceHealthItem[]>([]);
  const [matches, setMatches] = useState<AdminAtRiskMatch[]>([]);
  const [events, setEvents] = useState<AdminAuditEvent[]>([]);
  const [reports, setReports] = useState<AdminReportItem[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [busySlug, setBusySlug] = useState<string | null>(null);

  async function reloadPanel() {
    const data = await loadPanelData();
    setSources(data.sources);
    setMatches(data.matches);
    setEvents(data.events);
    setReports(data.reports);
  }

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      try {
        const body = await fetchAuthMe();
        if (cancelled) {
          return;
        }
        setMe(body);
        if (body.authenticated && body.role === "maintainer") {
          const data = await loadPanelData();
          if (!cancelled) {
            setSources(data.sources);
            setMatches(data.matches);
            setEvents(data.events);
            setReports(data.reports);
          }
        }
      } catch {
        if (!cancelled) {
          setError("Não foi possível carregar o painel.");
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  async function onAction(slug: string, action: "confirm" | "correct" | "mark_divergent") {
    const reason = window.prompt("Motivo da ação (obrigatório):");
    if (!reason?.trim()) {
      return;
    }
    setBusySlug(slug);
    setError(null);
    try {
      await postAdminMatchAction(slug, {
        action,
        surface: "broadcast",
        reason: reason.trim(),
        value: action === "correct" ? "available" : undefined,
      });
      await reloadPanel();
    } catch {
      setError("Falha ao aplicar ação.");
    } finally {
      setBusySlug(null);
    }
  }

  if (me && (!me.authEnabled || !me.authenticated || me.role !== "maintainer")) {
    return (
      <section className="admin-page">
        <h1>Painel</h1>
        <p className="lede">Acesso restrito a mantenedores.</p>
        <p>
          <Link to="/">Voltar ao início</Link>
        </p>
      </section>
    );
  }

  return (
    <section className="admin-page">
      <h1>Painel do mantenedor</h1>
      <p className="lede">Saúde de fontes, partidas em risco e auditoria.</p>
      {error ? (
        <p className="settings-error" role="alert">
          {error}
        </p>
      ) : null}

      <h2>Fontes</h2>
      {sources.length === 0 ? (
        <p className="note">Nenhuma saúde de fonte registrada ainda.</p>
      ) : (
        <ul className="admin-list">
          {sources.map((s) => (
            <li key={s.sourceId}>
              <strong>{s.sourceId}</strong> — falhas: {s.consecutiveFailures}
              {s.lastError ? ` — ${s.lastError}` : ""}
            </li>
          ))}
        </ul>
      )}

      <h2>Partidas em risco</h2>
      {matches.length === 0 ? (
        <p className="note">Nenhuma partida em risco no momento.</p>
      ) : (
        <ul className="admin-list">
          {matches.map((m) => (
            <li key={m.slug}>
              <div>
                <Link to={`/jogos/${m.slug}`}>
                  {m.homeClub} x {m.awayClub}
                </Link>
                <span className="note">
                  {" "}
                  · TV {m.broadcastState} · escalação {m.lineupState}
                </span>
              </div>
              <div className="admin-actions">
                <button
                  type="button"
                  disabled={busySlug === m.slug}
                  onClick={() => void onAction(m.slug, "confirm")}
                >
                  Confirmar TV
                </button>
                <button
                  type="button"
                  disabled={busySlug === m.slug}
                  onClick={() => void onAction(m.slug, "correct")}
                >
                  Corrigir
                </button>
                <button
                  type="button"
                  disabled={busySlug === m.slug}
                  onClick={() => void onAction(m.slug, "mark_divergent")}
                >
                  Marcar divergente
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}

      <h2>Relatos abertos</h2>
      {reports.length === 0 ? (
        <p className="note">Nenhum relato aberto.</p>
      ) : (
        <ul className="admin-list">
          {reports.map((r) => (
            <li key={r.id}>
              <div>
                {r.createdAt}: {r.contextType}/{r.contextSlug || "—"} — {r.message}
              </div>
              <div className="admin-actions">
                <button
                  type="button"
                  onClick={() =>
                    void reviewAdminReport(r.id, { status: "reviewed" }).then(reloadPanel)
                  }
                >
                  Revisado
                </button>
                <button
                  type="button"
                  onClick={() =>
                    void reviewAdminReport(r.id, { status: "dismissed" }).then(reloadPanel)
                  }
                >
                  Dispensar
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}

      <h2>Auditoria recente</h2>
      {events.length === 0 ? (
        <p className="note">Sem eventos ainda.</p>
      ) : (
        <ul className="admin-list">
          {events.map((e) => (
            <li key={e.id}>
              {e.createdAt}: {e.actor} — {e.action} em {e.entityType}/{e.entityId} ({e.reason})
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
