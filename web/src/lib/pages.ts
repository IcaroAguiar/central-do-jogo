import type { components } from "../api/generated/schema";

/** OpenAPI SSRPage discriminator for PAT-004 #initial-data payloads. */
export type SSRPage = components["schemas"]["SSRPage"];

/**
 * Canonical SSR page IDs. Values must stay identical to:
 * - OpenAPI `components.schemas.SSRPage`
 * - `internal/api` page constants
 */
export const SSR_PAGE = {
  home: "home",
  club: "club",
  match: "match",
} as const satisfies { [K in SSRPage]: K };
