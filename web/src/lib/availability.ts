import type { AvailabilityState, ConfidenceLevel } from "../api/client";

/**
 * pt-BR copy for REQ-010: every match-scoped data surface (broadcast,
 * lineup, news) always has an explicit state instead of silently omitting
 * data when nothing is known yet.
 */
const AVAILABILITY_LABELS: Record<AvailabilityState, string> = {
  available: "Disponível",
  awaiting_publication: "Aguardando divulgação oficial",
  not_found: "Não encontrado nas fontes monitoradas",
  divergent: "Fontes divergentes",
  no_coverage: "Sem cobertura para esta competição",
};

export function availabilityLabel(state: AvailabilityState): string {
  return AVAILABILITY_LABELS[state] ?? state;
}

/** True when the data surface is not simply "available", meaning the UI
 * should show an explicit gap message alongside any partial data. */
export function isIncomplete(state: AvailabilityState): boolean {
  return state !== "available";
}

const CONFIDENCE_LABELS: Record<ConfidenceLevel, string> = {
  high: "Confiança alta",
  medium: "Confiança média",
  low: "Confiança baixa",
};

export function confidenceLabel(level: ConfidenceLevel): string {
  return CONFIDENCE_LABELS[level] ?? level;
}

/** Broadcasts with medium/low confidence, or a non-"available" broadcast
 * state, mean the "onde assistir" answer could still change before kickoff
 * and the UI must disclose that instead of presenting it as final. */
export function needsIncompleteDisclaimer(
  broadcastState: AvailabilityState,
  confidences: readonly ConfidenceLevel[],
): boolean {
  if (isIncomplete(broadcastState)) {
    return true;
  }
  return confidences.some((level) => level === "low" || level === "medium");
}
