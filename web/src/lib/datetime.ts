/**
 * All product timestamps are displayed in America/Sao_Paulo ("Horário de
 * Brasília") regardless of the visitor's device timezone (CON-008). Brazil
 * has not observed DST since 2019, so this is a stable, deterministic
 * offset-based zone identifier.
 */
export const BRASILIA_TIME_ZONE = "America/Sao_Paulo";
export const BRASILIA_LABEL = "Horário de Brasília";

const dateTimeFormatter = new Intl.DateTimeFormat("pt-BR", {
  timeZone: BRASILIA_TIME_ZONE,
  day: "2-digit",
  month: "2-digit",
  year: "numeric",
  hour: "2-digit",
  minute: "2-digit",
});

const weekdayFormatter = new Intl.DateTimeFormat("pt-BR", {
  timeZone: BRASILIA_TIME_ZONE,
  weekday: "long",
});

const timeFormatter = new Intl.DateTimeFormat("pt-BR", {
  timeZone: BRASILIA_TIME_ZONE,
  hour: "2-digit",
  minute: "2-digit",
});

/** Formats an ISO instant as "dd/mm/aaaa HH:mm" in Brasília time. */
export function formatKickoff(iso: string | null | undefined): string {
  if (!iso) {
    return "Data a confirmar";
  }
  return dateTimeFormatter.format(new Date(iso));
}

/** Formats an ISO instant with the explicit "Horário de Brasília" suffix. */
export function formatKickoffWithLabel(iso: string | null | undefined): string {
  if (!iso) {
    return "Data a confirmar";
  }
  return `${formatKickoff(iso)} (${BRASILIA_LABEL})`;
}

/** Formats an ISO instant as a lowercase Brasília weekday name (e.g. "domingo"). */
export function formatWeekday(iso: string): string {
  return weekdayFormatter.format(new Date(iso));
}

/** Formats an ISO instant as "HH:mm" in Brasília time, for compact agenda rows. */
export function formatTime(iso: string): string {
  return timeFormatter.format(new Date(iso));
}

/** Formats an arbitrary ISO instant (e.g. a cache/attempt timestamp) as a
 * relative-ish Brasília date+time string, used for "última tentativa" and
 * offline "dados salvos em" copy. */
export function formatTimestamp(iso: string | null | undefined): string | null {
  if (!iso) {
    return null;
  }
  return `${formatKickoff(iso)} (${BRASILIA_LABEL})`;
}
