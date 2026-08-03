/**
 * Restrict notification deep-links to same-origin paths (open-redirect guard).
 */
export function safeNotificationUrl(raw: unknown, origin: string): string {
  if (typeof raw !== "string" || raw.trim() === "") {
    return "/";
  }
  try {
    const resolved = new URL(raw, origin);
    if (resolved.origin !== origin) {
      return "/";
    }
    return `${resolved.pathname}${resolved.search}${resolved.hash}` || "/";
  } catch {
    return "/";
  }
}
