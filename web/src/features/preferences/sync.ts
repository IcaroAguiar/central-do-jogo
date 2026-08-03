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
let remoteReady = false;
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

/** True only after an owner-bound prefs GET succeeded (safe to PUT). */
export function getRemoteReady(): boolean {
  return remoteReady;
}

export function getPrimaryConflict(): PrimaryConflict | null {
  return primaryConflict;
}

function setSyncState(next: {
  authenticated?: boolean;
  remoteReady?: boolean;
  syncing?: boolean;
  conflict?: PrimaryConflict | null;
}): void {
  if (next.authenticated !== undefined) authenticated = next.authenticated;
  if (next.remoteReady !== undefined) remoteReady = next.remoteReady;
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
        remoteReady: outcome.remoteReady === true,
        conflict: outcome.conflict,
        syncing: false,
      });
    } catch {
      // Unexpected failures leave prior auth flag alone; only mark sync done.
      setSyncState({ syncing: false, remoteReady: false });
    }
  })();
}

export type SyncOutcome =
  | { authenticated: false; conflict: null; remoteReady: false }
  | {
      authenticated: true;
      conflict: PrimaryConflict | null;
      remoteReady: boolean;
      pushFailed?: boolean;
    };

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
    return { authenticated: false, conflict: null, remoteReady: false };
  }
  if (!me.authenticated || !me.authEnabled) {
    return { authenticated: false, conflict: null, remoteReady: false };
  }

  const ownerKey = me.email || me.displayName || "authenticated";
  const priorOwner = getPrefsOwner();
  const foreignLocal = priorOwner !== null && priorOwner !== ownerKey;

  let remote: Awaited<ReturnType<typeof fetchPreferences>>;
  try {
    remote = await fetchPreferences();
  } catch {
    if (foreignLocal) {
      // Do not enable PUT while foreign leftovers remain and remote is unknown.
      applyLocalPreferences(null, []);
      setPrefsOwner(ownerKey);
      return { authenticated: true, conflict: null, remoteReady: false, pushFailed: true };
    }
    // Same owner / first sync: session is real, but skip remote writes until GET works.
    return { authenticated: true, conflict: null, remoteReady: false, pushFailed: true };
  }

  const remoteSnapshot = {
    primaryClub: remote.primaryClubSlug ?? null,
    favoriteClubs: remote.favoriteClubSlugs ?? [],
  };

  if (foreignLocal) {
    // Previous account's local leftovers must not merge/PUT into this user.
    applyLocalPreferences(remoteSnapshot.primaryClub, remoteSnapshot.favoriteClubs);
    setPrefsOwner(ownerKey);
    return { authenticated: true, conflict: null, remoteReady: true };
  }

  const merge = mergePreferences(
    { primaryClub: getPrimaryClub(), favoriteClubs: getFavoriteClubs() },
    remoteSnapshot,
  );
  applyLocalPreferences(merge.primaryClub, merge.favoriteClubs);
  setPrefsOwner(ownerKey);

  if (merge.primaryConflict) {
    return { authenticated: true, conflict: merge.primaryConflict, remoteReady: true };
  }

  try {
    await pushLocalPreferences();
  } catch {
    return { authenticated: true, conflict: null, remoteReady: true, pushFailed: true };
  }

  return { authenticated: true, conflict: null, remoteReady: true };
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
  setSyncState({
    authenticated: value,
    remoteReady: value ? remoteReady : false,
    conflict: value ? primaryConflict : null,
  });
}

/** Test-only reset for vitest isolation. */
export function __resetSyncForTests(): void {
  authenticated = false;
  remoteReady = false;
  syncing = false;
  primaryConflict = null;
  syncStarted = false;
  if (hasStorage()) {
    window.localStorage.removeItem(PREFS_OWNER_KEY);
  }
}
