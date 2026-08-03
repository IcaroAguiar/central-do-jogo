import {
  fetchPreferences,
  type PreferencesResponse,
  type PreferencesUpdate,
  putPreferences,
} from "../../api/client";
import { mergePreferences, type PrimaryConflict } from "./merge";
import { applyLocalPreferences, getFavoriteClubs, getPrimaryClub } from "./preferences";

export type { PrimaryConflict };

type Listener = () => void;
const listeners = new Set<Listener>();

let authenticated = false;
let syncing = false;
let primaryConflict: PrimaryConflict | null = null;
let syncStarted = false;

function notify(): void {
  for (const listener of listeners) {
    listener();
  }
}

export function subscribeSync(listener: Listener): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export function getSyncing(): boolean {
  return syncing;
}

export function getAuthenticated(): boolean {
  return authenticated;
}

export function getPrimaryConflict(): PrimaryConflict | null {
  return primaryConflict;
}

function setSyncState(next: {
  authenticated?: boolean;
  syncing?: boolean;
  conflict?: PrimaryConflict | null;
}): void {
  if (next.authenticated !== undefined) authenticated = next.authenticated;
  if (next.syncing !== undefined) syncing = next.syncing;
  if (next.conflict !== undefined) primaryConflict = next.conflict;
  notify();
}

/** Idempotent account sync for the SPA lifetime (REQ-006). */
export function ensurePreferencesSynced(): void {
  if (syncStarted || typeof window === "undefined") return;
  syncStarted = true;
  setSyncState({ syncing: true });
  void (async () => {
    try {
      const outcome = await syncPreferencesWithAccount();
      setSyncState({
        authenticated: outcome.authenticated,
        conflict: outcome.conflict,
        syncing: false,
      });
    } catch {
      setSyncState({ authenticated: false, conflict: null, syncing: false });
    }
  })();
}

export interface SyncOutcome {
  authenticated: boolean;
  conflict: PrimaryConflict | null;
}

/** Pull remote prefs, merge with local without silent overwrite, and push when safe. */
export async function syncPreferencesWithAccount(): Promise<SyncOutcome> {
  let remote: PreferencesResponse;
  try {
    remote = await fetchPreferences();
  } catch {
    return { authenticated: false, conflict: null };
  }

  const merge = mergePreferences(
    { primaryClub: getPrimaryClub(), favoriteClubs: getFavoriteClubs() },
    {
      primaryClub: remote.primaryClubSlug ?? null,
      favoriteClubs: remote.favoriteClubSlugs ?? [],
    },
  );
  applyLocalPreferences(merge.primaryClub, merge.favoriteClubs);

  if (!merge.primaryConflict) {
    await pushLocalPreferences();
  }

  return {
    authenticated: true,
    conflict: merge.primaryConflict,
  };
}

export async function pushLocalPreferences(): Promise<void> {
  const body: PreferencesUpdate = {
    primaryClubSlug: getPrimaryClub(),
    favoriteClubSlugs: getFavoriteClubs(),
  };
  await putPreferences(body);
}

export function clearPrimaryConflict(): void {
  setSyncState({ conflict: null });
}

export function markAuthenticated(value: boolean): void {
  setSyncState({ authenticated: value });
}

/** Test-only reset for vitest isolation. */
export function __resetSyncForTests(): void {
  authenticated = false;
  syncing = false;
  primaryConflict = null;
  syncStarted = false;
}
