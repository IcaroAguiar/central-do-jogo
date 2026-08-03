/**
 * Shared Cache API allow/deny rules for the PWA service worker.
 * Kept free of ServiceWorker globals so unit tests can import it.
 */

/** Path fragments that must never be served from or written to the API cache. */
export const NEVER_CACHE_FRAGMENTS = [
  "/api/v1/auth",
  "/api/v1/preferences",
  "/api/v1/push",
  "/export",
  "/admin",
  "/privacy",
] as const;

export function isPrivateApiPath(pathname: string): boolean {
  return NEVER_CACHE_FRAGMENTS.some((fragment) => pathname.includes(fragment));
}

export function isPublicApiPath(pathname: string): boolean {
  return pathname.startsWith("/api/v1/") && !isPrivateApiPath(pathname);
}
