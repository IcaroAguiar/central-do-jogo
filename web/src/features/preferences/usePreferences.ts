import { useCallback, useSyncExternalStore } from "react";
import {
  getFavoriteClubs,
  getPrimaryClub,
  setPrimaryClub as setPrimaryClubValue,
  subscribe,
  toggleFavoriteClub as toggleFavoriteClubValue,
} from "./preferences";

export interface ClubPreferences {
  primaryClub: string | null;
  favoriteClubs: string[];
  isPrimary: (slug: string) => boolean;
  isFavorite: (slug: string) => boolean;
  setPrimaryClub: (slug: string | null) => void;
  toggleFavoriteClub: (slug: string) => void;
}

/** Reactive view over the localStorage-backed club preferences (cdj:primaryClub,
 * cdj:favoriteClubs), so UI toggles update immediately across the page. */
export function usePreferences(): ClubPreferences {
  const primaryClub = useSyncExternalStore(subscribe, getPrimaryClub, () => null);
  const favoriteClubs = useSyncExternalStore(subscribe, getFavoriteClubs, () => []);

  const isPrimary = useCallback((slug: string) => primaryClub === slug, [primaryClub]);
  const isFavorite = useCallback((slug: string) => favoriteClubs.includes(slug), [favoriteClubs]);

  return {
    primaryClub,
    favoriteClubs,
    isPrimary,
    isFavorite,
    setPrimaryClub: setPrimaryClubValue,
    toggleFavoriteClub: toggleFavoriteClubValue,
  };
}
