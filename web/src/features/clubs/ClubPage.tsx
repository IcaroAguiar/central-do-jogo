import { useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import type { AgendaRange, ClubDetail, ClubMatchesResponse } from "../../api/client";
import { ApiRequestError, fetchClub, fetchClubMatches } from "../../api/client";
import { useApiResource } from "../../api/useApiResource";
import { formatKickoff } from "../../lib/datetime";
import { readInitialData } from "../../lib/initialData";
import { LoadErrorPage } from "../../pages/LoadErrorPage";
import { NotFoundPage } from "../../pages/NotFoundPage";
import { usePreferences } from "../preferences/usePreferences";
import { ShareButton } from "../sharing/ShareButton";
import { AgendaTabs } from "./AgendaTabs";

interface ClubInitialData {
  notFound?: boolean;
  club?: ClubDetail;
  matches?: ClubMatchesResponse;
}

export function ClubPage() {
  const { slug = "" } = useParams<{ slug: string }>();
  const [initial] = useState(() => readInitialData<"club", ClubInitialData>("club"));
  const initialMatchesForSlug = initial?.club?.slug === slug ? initial.matches : undefined;
  const [range, setRange] = useState<AgendaRange>(initialMatchesForSlug?.range ?? "week");
  const prefs = usePreferences();

  const club = useApiResource(
    () => fetchClub(slug),
    [slug],
    initial?.club?.slug === slug ? initial.club : undefined,
  );
  const matches = useApiResource(
    () => fetchClubMatches(slug, range),
    [slug, range],
    initialMatchesForSlug?.range === range ? initialMatchesForSlug : undefined,
  );

  const shareUrl = useMemo(
    () =>
      typeof window !== "undefined"
        ? `${window.location.origin}/clubes/${slug}`
        : `/clubes/${slug}`,
    [slug],
  );

  if (initial?.notFound) {
    return <NotFoundPage message="Não encontramos um clube com este endereço." />;
  }

  if (club.error && !club.data) {
    if (club.error instanceof ApiRequestError && club.error.status === 404) {
      return <NotFoundPage message="Não encontramos um clube com este endereço." />;
    }
    return (
      <LoadErrorPage
        message="Não foi possível carregar este clube. Verifique a conexão e tente novamente."
        onRetry={club.retry}
      />
    );
  }

  if (club.status === "loading" && !club.data) {
    return (
      <p aria-live="polite" className="loading">
        Carregando clube…
      </p>
    );
  }

  if (!club.data) {
    return <NotFoundPage message="Não encontramos um clube com este endereço." />;
  }

  const isPrimary = prefs.isPrimary(slug);
  const isFavorite = prefs.isFavorite(slug);

  return (
    <article>
      <header className="club-header">
        <p className="eyebrow">Clube</p>
        <h1>{club.data.name}</h1>
        {club.data.shortName ? <p className="club-header__short">{club.data.shortName}</p> : null}
        <div className="club-header__actions">
          <button
            type="button"
            aria-pressed={isPrimary}
            onClick={() => prefs.setPrimaryClub(isPrimary ? null : slug)}
          >
            {isPrimary ? "Clube principal ✓" : "Definir como principal"}
          </button>
          <button
            type="button"
            aria-pressed={isFavorite}
            onClick={() => prefs.toggleFavoriteClub(slug)}
          >
            {isFavorite ? "Favorito ✓" : "Adicionar aos favoritos"}
          </button>
          <ShareButton
            title={club.data.name}
            text={`Agenda de ${club.data.name} — Central do Jogo`}
            url={shareUrl}
          />
        </div>
      </header>

      <section aria-labelledby="agenda-heading">
        <h2 id="agenda-heading">Agenda</h2>
        <AgendaTabs active={range} onChange={setRange} />
        <div role="tabpanel" id="agenda-panel" aria-labelledby={`agenda-tab-${range}`}>
          {matches.status === "loading" && !matches.data ? (
            <p aria-live="polite">Carregando jogos…</p>
          ) : matches.data && matches.data.matches.length > 0 ? (
            <ul className="agenda-list">
              {matches.data.matches.map((match) => (
                <li key={match.slug}>
                  <a href={`/jogos/${match.slug}`}>
                    {match.homeClub.shortName} x {match.awayClub.shortName}
                  </a>
                  <span> — {formatKickoff(match.kickoffAt)}</span>
                </li>
              ))}
            </ul>
          ) : (
            <p>Nenhum jogo encontrado para este período.</p>
          )}
        </div>
      </section>
    </article>
  );
}
