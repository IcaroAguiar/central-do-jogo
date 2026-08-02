import { useSyncExternalStore } from "react";
import { getSnapshot, type OfflineSignalState, subscribe } from "./offlineSignal";

const SERVER_SNAPSHOT: OfflineSignalState = { isDegraded: false, cachedAt: null };

function getServerSnapshot(): OfflineSignalState {
  return SERVER_SNAPSHOT;
}

/** Reactive view over the shared offline-degradation signal (see
 * offlineSignal.ts). `getSnapshot` returns the same module-level object
 * reference until `reportServedFromCache`/`reportFreshFetch` update it, so
 * this is safe to use with useSyncExternalStore without causing render
 * loops. */
export function useOfflineSignal(): OfflineSignalState {
  return useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
}
