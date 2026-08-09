import type { components } from "./generated/schema";

export type AuthMeResponse = components["schemas"]["AuthMeResponse"];
export type PreferencesResponse = components["schemas"]["PreferencesResponse"];
export type PreferencesUpdate = components["schemas"]["PreferencesUpdate"];
export type PushVapidPublicKeyResponse = components["schemas"]["PushVapidPublicKeyResponse"];
export type PushSubscriptionCreate = components["schemas"]["PushSubscriptionCreate"];
export type PushSubscriptionListResponse = components["schemas"]["PushSubscriptionListResponse"];
export type PushSubscriptionSummary = components["schemas"]["PushSubscriptionSummary"];
export type PrivacyExportResponse = components["schemas"]["PrivacyExportResponse"];
export type PrivacyAnalyticsEventCreate = components["schemas"]["PrivacyAnalyticsEventCreate"];
export type ClubDetail = components["schemas"]["ClubDetail"];
export type ClubMatchesResponse = components["schemas"]["ClubMatchesResponse"];
export type ClubMatchSummary = components["schemas"]["ClubMatchSummary"];
export type MatchDetail = components["schemas"]["MatchDetail"];
export type SearchResponse = components["schemas"]["SearchResponse"];
export type SearchClubResult = components["schemas"]["SearchClubResult"];
export type SearchMatchResult = components["schemas"]["SearchMatchResult"];
export type AgendaRange = components["schemas"]["AgendaRange"];
export type AvailabilityState = components["schemas"]["AvailabilityState"];
export type ConfidenceLevel = components["schemas"]["ConfidenceLevel"];
export type AccessType = components["schemas"]["AccessType"];
export type KickoffState = components["schemas"]["KickoffState"];
export type BroadcastView = components["schemas"]["BroadcastView"];
export type LineupView = components["schemas"]["LineupView"];
export type NewsLinkView = components["schemas"]["NewsLinkView"];

/**
 * X-Cache header set by our service worker (web/src/pwa/sw.ts) on cached
 * GET /api/v1/* responses so the UI can render "showing saved data from
 * <cachedAt>" instead of silently serving stale content.
 */
export const CACHED_AT_HEADER = "x-cdj-cached-at";

/** ApiRequestError wraps the standard {"error":{"code","message"}} envelope
 * every /api/v1 route returns (see internal/platform/http/apierror.go). */
export class ApiRequestError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "ApiRequestError";
    this.status = status;
    this.code = code;
  }
}

export interface ApiResult<T> {
  data: T;
  /** ISO timestamp when this response was cached by the service worker, or
   * null when it came straight from the network. */
  cachedAt: string | null;
}

interface ErrorEnvelope {
  error: { code: string; message: string };
}

async function request<T>(path: string): Promise<ApiResult<T>> {
  const response = await fetch(path, { headers: { Accept: "application/json" } });

  if (!response.ok) {
    let code = "unknown_error";
    let message = response.statusText || "request failed";
    try {
      const body = (await response.json()) as ErrorEnvelope;
      code = body.error.code;
      message = body.error.message;
    } catch {
      // Body was not JSON (e.g. network-level failure page); keep defaults.
    }
    throw new ApiRequestError(response.status, code, message);
  }

  const data = (await response.json()) as T;
  return { data, cachedAt: response.headers.get(CACHED_AT_HEADER) };
}

export function fetchSearch(query: string): Promise<ApiResult<SearchResponse>> {
  const params = new URLSearchParams({ q: query });
  return request<SearchResponse>(`/api/v1/search?${params.toString()}`);
}

export function fetchClub(slug: string): Promise<ApiResult<ClubDetail>> {
  return request<ClubDetail>(`/api/v1/clubs/${encodeURIComponent(slug)}`);
}

export function fetchClubMatches(
  slug: string,
  range: AgendaRange,
  season?: number,
): Promise<ApiResult<ClubMatchesResponse>> {
  const params = new URLSearchParams({ range });
  if (season) {
    params.set("season", String(season));
  }
  return request<ClubMatchesResponse>(
    `/api/v1/clubs/${encodeURIComponent(slug)}/matches?${params.toString()}`,
  );
}

export function fetchMatch(slug: string): Promise<ApiResult<MatchDetail>> {
  return request<MatchDetail>(`/api/v1/matches/${encodeURIComponent(slug)}`);
}

export async function fetchAuthMe(): Promise<AuthMeResponse> {
  const response = await fetch("/api/v1/auth/me", {
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  });
  if (!response.ok) {
    throw new ApiRequestError(response.status, "auth_me_failed", "failed to load auth status");
  }
  return (await response.json()) as AuthMeResponse;
}

export async function logoutAuth(): Promise<void> {
  const response = await fetch("/api/v1/auth/logout", {
    method: "POST",
    credentials: "same-origin",
  });
  if (!response.ok && response.status !== 204) {
    throw new ApiRequestError(response.status, "logout_failed", "failed to logout");
  }
}

export async function fetchPreferences(): Promise<PreferencesResponse> {
  const response = await fetch("/api/v1/preferences", {
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  });
  if (!response.ok) {
    let code = "preferences_failed";
    let message = "failed to load preferences";
    try {
      const body = (await response.json()) as ErrorEnvelope;
      code = body.error.code;
      message = body.error.message;
    } catch {
      // keep defaults
    }
    throw new ApiRequestError(response.status, code, message);
  }
  return (await response.json()) as PreferencesResponse;
}

export async function putPreferences(body: PreferencesUpdate): Promise<PreferencesResponse> {
  const response = await fetch("/api/v1/preferences", {
    method: "PUT",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    let code = "preferences_put_failed";
    let message = "failed to save preferences";
    try {
      const envelope = (await response.json()) as ErrorEnvelope;
      code = envelope.error.code;
      message = envelope.error.message;
    } catch {
      // keep defaults
    }
    throw new ApiRequestError(response.status, code, message);
  }
  return (await response.json()) as PreferencesResponse;
}

async function authedJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  if (!headers.has("Accept")) {
    headers.set("Accept", "application/json");
  }
  const response = await fetch(path, {
    ...init,
    credentials: "same-origin",
    headers,
  });
  if (!response.ok) {
    let code = "request_failed";
    let message = response.statusText || "request failed";
    try {
      const body = (await response.json()) as ErrorEnvelope;
      code = body.error.code;
      message = body.error.message;
    } catch {
      // keep defaults
    }
    throw new ApiRequestError(response.status, code, message);
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

export function fetchPushVapidPublicKey(): Promise<PushVapidPublicKeyResponse> {
  return authedJSON<PushVapidPublicKeyResponse>("/api/v1/push/vapid-public-key");
}

export function fetchPushSubscriptions(): Promise<PushSubscriptionListResponse> {
  return authedJSON<PushSubscriptionListResponse>("/api/v1/push/subscriptions");
}

export function createPushSubscription(
  body: PushSubscriptionCreate,
): Promise<PushSubscriptionSummary> {
  return authedJSON<PushSubscriptionSummary>("/api/v1/push/subscriptions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
}

export function fetchPrivacyExport(): Promise<PrivacyExportResponse> {
  return authedJSON<PrivacyExportResponse>("/api/v1/privacy/export");
}

export function deletePrivacyAccount(): Promise<void> {
  return authedJSON<void>("/api/v1/privacy/account", { method: "DELETE" });
}

export function postPrivacyAnalyticsEvent(body: PrivacyAnalyticsEventCreate): Promise<void> {
  return authedJSON<void>("/api/v1/privacy/events", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
}

export function deletePushSubscription(endpoint: string): Promise<void> {
  return authedJSON<void>("/api/v1/push/subscriptions", {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ endpoint }),
  });
}
