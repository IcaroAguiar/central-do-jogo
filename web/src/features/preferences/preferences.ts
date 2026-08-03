/**
 * Club preferences with localStorage as the visitor source of truth (REQ-006).
 * Account sync lives in sync.ts / usePreferences and never silently overwrites
 * local choices when remotes disagree.
 */
export const PRIMARY_CLUB_KEY = "cdj:primaryClub";
export const FAVORITE_CLUBS_KEY = "cdj:favoriteClubs";

type Listener = () => void;
const listeners = new Set<Listener>();

function notify(): void {
  for (const listener of listeners) {
    listener();
  }
}

function hasStorage(): boolean {
  try {
    return typeof window !== "undefined" && !!window.localStorage;
  } catch {
    return false;
  }
}

export function subscribe(listener: Listener): () => void {
  listeners.add(listener);
  const onStorage = (event: StorageEvent) => {
    if (event.key === null || event.key === PRIMARY_CLUB_KEY || event.key === FAVORITE_CLUBS_KEY) {
      listener();
    }
  };
  if (typeof window !== "undefined") {
    window.addEventListener("storage", onStorage);
  }
  return () => {
    listeners.delete(listener);
    if (typeof window !== "undefined") {
      window.removeEventListener("storage", onStorage);
    }
  };
}

export function getPrimaryClub(): string | null {
  if (!hasStorage()) return null;
  return window.localStorage.getItem(PRIMARY_CLUB_KEY);
}

export function setPrimaryClub(slug: string | null): void {
  if (!hasStorage()) return;
  if (slug === null) {
    window.localStorage.removeItem(PRIMARY_CLUB_KEY);
  } else {
    window.localStorage.setItem(PRIMARY_CLUB_KEY, slug);
  }
  notify();
}

const EMPTY_FAVORITES: string[] = [];
// getFavoriteClubs is used as a useSyncExternalStore getSnapshot (usePreferences.ts).
// React re-invokes getSnapshot on every render and compares the result via
// Object.is to the previous one; returning a freshly-allocated array on every
// call (even when localStorage is unchanged) makes every render look like a
// "new" snapshot, which triggers an infinite render loop. Caching by raw
// string keeps the reference stable across calls when nothing changed.
let favoriteClubsCacheRaw: string | null | undefined;
let favoriteClubsCache: string[] = EMPTY_FAVORITES;

export function getFavoriteClubs(): string[] {
  if (!hasStorage()) return EMPTY_FAVORITES;
  const raw = window.localStorage.getItem(FAVORITE_CLUBS_KEY);
  if (raw === favoriteClubsCacheRaw) {
    return favoriteClubsCache;
  }
  favoriteClubsCacheRaw = raw;
  favoriteClubsCache = parseFavoriteClubs(raw);
  return favoriteClubsCache;
}

function parseFavoriteClubs(raw: string | null): string[] {
  if (!raw) return EMPTY_FAVORITES;
  try {
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return EMPTY_FAVORITES;
    return parsed.filter((item): item is string => typeof item === "string");
  } catch {
    return EMPTY_FAVORITES;
  }
}

function setFavoriteClubs(slugs: string[]): void {
  if (!hasStorage()) return;
  window.localStorage.setItem(FAVORITE_CLUBS_KEY, JSON.stringify(slugs));
  notify();
}

export function isFavoriteClub(slug: string): boolean {
  return getFavoriteClubs().includes(slug);
}

export function toggleFavoriteClub(slug: string): boolean {
  const current = getFavoriteClubs();
  const isFavorite = current.includes(slug);
  const next = isFavorite ? current.filter((item) => item !== slug) : [...current, slug];
  setFavoriteClubs(next);
  return !isFavorite;
}

/** Apply a merged snapshot to localStorage in one notify cycle. */
export function applyLocalPreferences(primaryClub: string | null, favoriteClubs: string[]): void {
  if (!hasStorage()) return;
  if (primaryClub === null) {
    window.localStorage.removeItem(PRIMARY_CLUB_KEY);
  } else {
    window.localStorage.setItem(PRIMARY_CLUB_KEY, primaryClub);
  }
  window.localStorage.setItem(FAVORITE_CLUBS_KEY, JSON.stringify(favoriteClubs));
  notify();
}

/** Test-only helper to reset preference state between test cases. */
export function __resetPreferencesForTests(): void {
  if (!hasStorage()) return;
  window.localStorage.removeItem(PRIMARY_CLUB_KEY);
  window.localStorage.removeItem(FAVORITE_CLUBS_KEY);
  window.localStorage.removeItem("cdj:prefsOwner");
  favoriteClubsCacheRaw = undefined;
  favoriteClubsCache = EMPTY_FAVORITES;
}
