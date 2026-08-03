/**
 * Explicit local/remote preference merge (REQ-006).
 * Favorites are unioned; primary club never silently overwrites either side.
 */

export interface ClubPreferenceSnapshot {
  primaryClub: string | null;
  favoriteClubs: string[];
}

export interface PrimaryConflict {
  local: string;
  remote: string;
}

export interface MergeResult {
  primaryClub: string | null;
  favoriteClubs: string[];
  primaryConflict: PrimaryConflict | null;
}

/** Deduplicate while preserving first-seen order (local favorites first). */
export function uniquePreserveOrder(slugs: string[]): string[] {
  const seen = new Set<string>();
  const out: string[] = [];
  for (const slug of slugs) {
    const trimmed = slug.trim();
    if (!trimmed || seen.has(trimmed)) continue;
    seen.add(trimmed);
    out.push(trimmed);
  }
  return out;
}

export function mergePreferences(
  local: ClubPreferenceSnapshot,
  remote: ClubPreferenceSnapshot,
): MergeResult {
  const favoriteClubs = uniquePreserveOrder([...local.favoriteClubs, ...remote.favoriteClubs]);

  if (local.primaryClub === remote.primaryClub) {
    return { primaryClub: local.primaryClub, favoriteClubs, primaryConflict: null };
  }
  if (local.primaryClub === null) {
    return { primaryClub: remote.primaryClub, favoriteClubs, primaryConflict: null };
  }
  if (remote.primaryClub === null) {
    return { primaryClub: local.primaryClub, favoriteClubs, primaryConflict: null };
  }
  return {
    primaryClub: local.primaryClub,
    favoriteClubs,
    primaryConflict: { local: local.primaryClub, remote: remote.primaryClub },
  };
}
