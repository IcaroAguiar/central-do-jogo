/**
 * Shared "are we actually degraded" signal for the offline banner (TASK-025).
 *
 * `navigator.onLine` alone is not a reliable offline indicator: some
 * browsers/environments (notably Chromium under Playwright's
 * `context.setOffline()`, and some captive-portal/VPN setups on real
 * devices) never flip it to `false` even though every network request is
 * failing. We instead track observable evidence from the API client: every
 * time `useApiResource` gets data back from the service worker's cache
 * fallback (web/src/pwa/sw.ts, `networkFirstApi`) because the live network
 * request failed, that is real proof the app is degraded, regardless of
 * what `navigator.onLine` claims. The signal clears itself as soon as any
 * resource completes a fresh (non-cached) network fetch again.
 */

export interface OfflineSignalState {
  isDegraded: boolean;
  cachedAt: string | null;
}

let state: OfflineSignalState = { isDegraded: false, cachedAt: null };
const listeners = new Set<() => void>();

function notify(): void {
  for (const listener of listeners) {
    listener();
  }
}

export function subscribe(listener: () => void): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export function getSnapshot(): OfflineSignalState {
  return state;
}

/** Called by useApiResource when a response came from the SW cache fallback
 * (i.e. the live network request failed). */
export function reportServedFromCache(cachedAt: string): void {
  state = { isDegraded: true, cachedAt };
  notify();
}

/** Called by useApiResource whenever a resource completes a genuinely fresh
 * network fetch, proving connectivity is restored. */
export function reportFreshFetch(): void {
  if (state.isDegraded) {
    state = { isDegraded: false, cachedAt: null };
    notify();
  }
}
