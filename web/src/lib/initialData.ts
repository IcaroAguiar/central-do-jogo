/**
 * Reads the #initial-data payload Go SSR (PAT-004, internal/platform/render)
 * embeds on first paint of "/", "/clubes/{slug}", and "/jogos/{slug}". This
 * lets the first render skip a duplicate network round-trip. It is only
 * meaningful for the exact page the browser navigated to directly, so the
 * payload is consumed (and discarded) at most once per full page load —
 * subsequent client-side navigations always fetch through the API client.
 */

import type { SSRPage } from "./pages";

type InitialDataEnvelope<TPage extends SSRPage, TData> = {
  page: TPage;
} & TData;

let consumed = false;

function readRaw(): Record<string, unknown> | null {
  if (consumed || typeof document === "undefined") {
    return null;
  }
  const el = document.getElementById("initial-data");
  consumed = true;
  if (!el?.textContent) {
    return null;
  }
  try {
    return JSON.parse(el.textContent) as Record<string, unknown>;
  } catch {
    return null;
  }
}

/** Returns the parsed initial-data payload only when it matches `page`,
 * otherwise null (including once it has already been consumed). */
export function readInitialData<TPage extends SSRPage, TData>(
  page: TPage,
): InitialDataEnvelope<TPage, TData> | null {
  const raw = readRaw();
  if (!raw || raw.page !== page) {
    return null;
  }
  return raw as InitialDataEnvelope<TPage, TData>;
}

/** Test-only helper to reset the module-level consumption guard. */
export function __resetInitialDataForTests(): void {
  consumed = false;
}
