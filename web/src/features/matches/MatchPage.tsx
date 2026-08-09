import { useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import type { MatchDetail } from "../../api/client";
import { ApiRequestError, fetchMatch } from "../../api/client";
import { useApiResource } from "../../api/useApiResource";
import { availabilityLabel, needsIncompleteDisclaimer } from "../../lib/availability";
import { formatKickoffWithLabel, formatTimestamp } from "../../lib/datetime";
import { readInitialData } from "../../lib/initialData";
import { SSR_PAGE } from "../../lib/pages";
import { LoadErrorPage } from "../../pages/LoadErrorPage";
import { NotFoundPage } from "../../pages/NotFoundPage";
import { ReportForm } from "../reports/ReportForm";
import { ShareButton } from "../sharing/ShareButton";
import { ConfidenceBadge } from "./ConfidenceBadge";

interface MatchInitialData {
  notFound?: boolean;
  match?: MatchDetail;
}

export function MatchPage() {
  const { slug = "" } = useParams<{ slug: string }>();
  const [initial] = useState(() =>
    readInitialData<typeof SSR_PAGE.match, MatchInitialData>(SSR_PAGE.match),
  );

  const match = useApiResource(
    () => fetchMatch(slug),
    [slug],
    initial?.match?.slug === slug ? initial.match : undefined,
  );

  const shareUrl = useMemo(
    () =>
      typeof window !== "undefined" ? `${window.location.origin}/jogos/${slug}` : `/jogos/${slug}`,
    [slug],
  );

  if (initial?.notFound) {
    return <NotFoundPage message="Não encontramos uma partida com este endereço." />;
  }

  if (match.error && !match.data) {
    if (match.error instanceof ApiRequestError && match.error.status === 404) {
      return <NotFoundPage message="Não encontramos uma partida com este endereço." />;
    }
    return (
      <LoadErrorPage
        message="Não foi possível carregar esta partida. Verifique a conexão e tente novamente."
        onRetry={match.retry}
      />
    );
  }

  if (match.status === "loading" && !match.data) {
    return (
      <p aria-live="polite" className="loading">
        Carregando partida…
      </p>
    );
  }

  if (!match.data) {
    return <NotFoundPage message="Não encontramos uma partida com este endereço." />;
  }

  const detail = match.data;
  const showBroadcastDisclaimer = needsIncompleteDisclaimer(
    detail.broadcastState,
    detail.broadcasts.map((b) => b.confidence),
  );

  return (
    <article>
      <header className="match-header">
        <p className="eyebrow">
          {detail.competition.name} · {detail.round}
        </p>
        <h1>
          {detail.homeClub.name} x {detail.awayClub.name}
        </h1>
        <p className="match-header__meta">
          {formatKickoffWithLabel(detail.kickoffAt)}
          {detail.venue ? ` — ${detail.venue}` : ""}
        </p>
        <ShareButton
          title={`${detail.homeClub.name} x ${detail.awayClub.name}`}
          text={`Onde assistir ${detail.homeClub.name} x ${detail.awayClub.name} — Central do Jogo`}
          url={shareUrl}
        />
      </header>

      <section
        aria-labelledby="broadcasts-heading"
        className="match-section match-section--primary"
      >
        <h2 id="broadcasts-heading">Onde assistir</h2>
        {showBroadcastDisclaimer ? (
          <p className="disclaimer" role="note">
            Esta informação pode mudar até o início da partida.
          </p>
        ) : null}
        {detail.broadcasts.length > 0 ? (
          <ul className="broadcast-list">
            {detail.broadcasts.map((broadcast, index) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: broadcasts have no stable id in the API contract.
              <li key={`${broadcast.channel}-${index}`}>
                <span className="broadcast-list__channel">
                  {broadcast.channel}
                  {broadcast.platform ? ` (${broadcast.platform})` : ""}
                </span>
                <span>
                  {" "}
                  —{" "}
                  {broadcast.access === "subscription"
                    ? "Assinatura"
                    : broadcast.access === "free"
                      ? "Gratuito"
                      : "Acesso a confirmar"}
                </span>
                <ConfidenceBadge level={broadcast.confidence} />
                <span className="broadcast-list__source">Fonte: {broadcast.source}</span>
              </li>
            ))}
          </ul>
        ) : (
          <GapNotice state={detail.broadcastState} lastAttemptAt={detail.broadcastLastAttemptAt} />
        )}
      </section>

      <section aria-labelledby="lineups-heading" className="match-section">
        <h2 id="lineups-heading">Escalações</h2>
        {detail.lineups.length > 0 ? (
          detail.lineups.map((lineup) => (
            <article key={lineup.side} className="lineup">
              <h3>
                {lineup.side === "home" ? detail.homeClub.name : detail.awayClub.name}
                {lineup.formation ? ` — ${lineup.formation}` : ""}
              </h3>
              {lineup.coach ? <p>Técnico: {lineup.coach}</p> : null}
              <p className="lineup__meta">
                {lineup.official ? "Escalação oficial" : "Escalação provável"}
                {lineup.publishedAt
                  ? ` — publicada em ${formatKickoffWithLabel(lineup.publishedAt)}`
                  : ""}
              </p>
              <ul>
                {lineup.players.map((player) => (
                  <li key={`${lineup.side}-${player.shirtNumber}-${player.name}`}>
                    {player.shirtNumber ? `${player.shirtNumber} — ` : ""}
                    {player.name}
                    {!player.isStarter ? " (reserva)" : ""}
                  </li>
                ))}
              </ul>
            </article>
          ))
        ) : (
          <GapNotice state={detail.lineupState} lastAttemptAt={detail.lineupLastAttemptAt} />
        )}
      </section>

      <section aria-labelledby="news-heading" className="match-section">
        <h2 id="news-heading">Notícias relacionadas</h2>
        {detail.news.length > 0 ? (
          <ul className="news-list">
            {detail.news.map((news) => (
              <li key={news.url}>
                <a href={news.url} target="_blank" rel="noreferrer">
                  {news.title}
                </a>
                <span> — {news.source}</span>
              </li>
            ))}
          </ul>
        ) : (
          <GapNotice state={detail.newsState} lastAttemptAt={detail.newsLastAttemptAt} />
        )}
      </section>

      <ReportForm contextType="match" contextSlug={slug} />
    </article>
  );
}

interface GapNoticeProps {
  state: MatchDetail["broadcastState"];
  lastAttemptAt: string | null;
}

/** REQ-010: renders the explicit gap state plus the last refresh attempt
 * timestamp, even when nothing was found, instead of silently showing
 * nothing. */
function GapNotice({ state, lastAttemptAt }: GapNoticeProps) {
  const lastAttempt = formatTimestamp(lastAttemptAt);
  return (
    <p className="gap-notice">
      {availabilityLabel(state)}
      {lastAttempt
        ? ` — última tentativa em ${lastAttempt}`
        : " — nenhuma tentativa registrada ainda"}
      .
    </p>
  );
}
