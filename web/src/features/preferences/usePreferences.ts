import { useCallback, useEffect, useSyncExternalStore } from "react";
import type { PrimaryConflict } from "./merge";
import {
  applyLocalPreferences,
  getFavoriteClubs,
  getPrimaryClub,
  setPrimaryClub as setPrimaryClubValue,
  subscribe,
  toggleFavoriteClub as toggleFavoriteClubValue,
} from "./preferences";
import {
  clearPrimaryConflict,
  ensurePreferencesSynced,
  getAuthenticated,
  getPrimaryConflict,
  getSyncing,
  pushLocalPreferences,
  subscribeSync,
} from "./sync";

export interface ClubPreferences {
  primaryClub: string | null;
  favoriteClubs: string[];
  syncing: boolean;
  primaryConflict: PrimaryConflict | null;
  isPrimary: (slug: string) => boolean;
  isFavorite: (slug: string) => boolean;
  setPrimaryClub: (slug: string | null) => void;
  toggleFavoriteClub: (slug: string) => void;
  resolvePrimaryConflict: (slug: string) => void;
}

/** Reactive view over localStorage-backed club preferences with optional
 * account sync after login (REQ-006). */
export function usePreferences(): ClubPreferences {
  const primaryClub = useSyncExternalStore(subscribe, getPrimaryClub, () => null);
  const favoriteClubs = useSyncExternalStore(subscribe, getFavoriteClubs, () => []);
  const syncing = useSyncExternalStore(subscribeSync, getSyncing, () => false);
  const authenticated = useSyncExternalStore(subscribeSync, getAuthenticated, () => false);
  const primaryConflict = useSyncExternalStore(subscribeSync, getPrimaryConflict, () => null);

  useEffect(() => {
    ensurePreferencesSynced();
  }, []);

  const persistRemote = useCallback(
    (nextConflict: PrimaryConflict | null = primaryConflict) => {
      if (!authenticated || nextConflict) return;
      void pushLocalPreferences().catch(() => {
        // Local prefs remain; remote retry happens on next sync.
      });
    },
    [authenticated, primaryConflict],
  );

  const setPrimaryClub = useCallback(
    (slug: string | null) => {
      setPrimaryClubValue(slug);
      if (primaryConflict && slug !== null) {
        if (slug === primaryConflict.local || slug === primaryConflict.remote) {
          clearPrimaryConflict();
          persistRemote(null);
          return;
        }
      }
      persistRemote();
    },
    [persistRemote, primaryConflict],
  );

  const toggleFavoriteClub = useCallback(
    (slug: string) => {
      toggleFavoriteClubValue(slug);
      persistRemote();
    },
    [persistRemote],
  );

  const resolvePrimaryConflict = useCallback(
    (slug: string) => {
      applyLocalPreferences(slug, getFavoriteClubs());
      clearPrimaryConflict();
      if (authenticated) {
        void pushLocalPreferences().catch(() => undefined);
      }
    },
    [authenticated],
  );

  const isPrimary = useCallback((slug: string) => primaryClub === slug, [primaryClub]);
  const isFavorite = useCallback((slug: string) => favoriteClubs.includes(slug), [favoriteClubs]);

  return {
    primaryClub,
    favoriteClubs,
    syncing,
    primaryConflict,
    isPrimary,
    isFavorite,
    setPrimaryClub,
    toggleFavoriteClub,
    resolvePrimaryConflict,
  };
}
