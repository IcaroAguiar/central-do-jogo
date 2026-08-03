import {
  fetchAuthMe,
  fetchPreferences,
  type PreferencesUpdate,
  putPreferences,
} from "../../api/client";
import { mergePreferences, type PrimaryConflict } from "./merge";
import { applyLocalPreferences, getFavoriteClubs, getPrimaryClub } from "./preferences";

export type { PrimaryConflict };

export const PREFS_OWNER_KEY = "cdj:prefsOwner";

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

function hasStorage(): boolean {
  try {
    return typeof window !== "undefined" && !!window.localStorage;
  } catch {
    return false;
  }
}

function getPrefsOwner(): string | null {
  if (!hasStorage()) return null;
  return window.localStorage.getItem(PREFS_OWNER_KEY);
}

function setPrefsOwner(owner: string | null): void {
  if (!hasStorage()) return;
  if (owner === null) {
    window.localStorage.removeItem(PREFS_OWNER_KEY);
  } else {
    window.localStorage.setItem(PREFS_OWNER_KEY, owner);
  }
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
      // Unexpected failures leave prior auth flag alone; only mark sync done.
      setSyncState({ syncing: false });
    }
  })();
}

export type SyncOutcome =
  | { authenticated: false; conflict: null }
  | { authenticated: true; conflict: PrimaryConflict | null; pushFailed?: boolean };

/**
 * Pull remote prefs and merge with local without silent overwrite.
 * Foreign localStorage (different prefs owner) never auto-pushes into the
 * new account — remote wins for that principal.
 */
export async function syncPreferencesWithAccount(): Promise<SyncOutcome> {
  let me: Awaited<ReturnType<typeof fetchAuthMe>>;
  try {
    me = await fetchAuthMe();
  } catch {
    return { authenticated: false, conflict: null };
  }
  if (!me.authenticated || !me.authEnabled) {
    return { authenticated: false, conflict: null };
  }

  const ownerKey = me.email || me.displayName || "authenticated";
  let remote: Awaited<ReturnType<typeof fetchPreferences>>;
  try {
    remote = await fetchPreferences();
  } catch {
    // Session may still be valid even if prefs GET fails; do not claim anonymous.
    return { authenticated: true, conflict: null, pushFailed: true };
  }

  const remoteSnapshot = {
    primaryClub: remote.primaryClubSlug ?? null,
    favoriteClubs: remote.favoriteClubSlugs ?? [],
  };
  const priorOwner = getPrefsOwner();
  const foreignLocal = priorOwner !== null && priorOwner !== ownerKey;

  if (foreignLocal) {
    // Previous account's local leftovers must not merge/PUT into this user.
    applyLocalPreferences(remoteSnapshot.primaryClub, remoteSnapshot.favoriteClubs);
    setPrefsOwner(ownerKey);
    return { authenticated: true, conflict: null };
  }

  const merge = mergePreferences(
    { primaryClub: getPrimaryClub(), favoriteClubs: getFavoriteClubs() },
    remoteSnapshot,
  );
  applyLocalPreferences(merge.primaryClub, merge.favoriteClubs);
  setPrefsOwner(ownerKey);

  if (merge.primaryConflict) {
    return { authenticated: true, conflict: merge.primaryConflict };
  }

  try {
    await pushLocalPreferences();
  } catch {
    return { authenticated: true, conflict: null, pushFailed: true };
  }

  return { authenticated: true, conflict: null };
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
  setSyncState({ authenticated: value, conflict: value ? primaryConflict : null });
}

/** Test-only reset for vitest isolation. */
export function __resetSyncForTests(): void {
  authenticated = false;
  syncing = false;
  primaryConflict = null;
  syncStarted = false;
  if (hasStorage()) {
    window.localStorage.removeItem(PREFS_OWNER_KEY);
  }
}
