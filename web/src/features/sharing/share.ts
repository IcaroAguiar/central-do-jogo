export interface SharePayload {
  title: string;
  text: string;
  url: string;
}

export type ShareOutcome =
  | { status: "shared" }
  | { status: "copied" }
  | { status: "cancelled" }
  | { status: "unavailable" };

/**
 * Tries the native Web Share API first (best mobile UX), falls back to
 * clipboard copy, and reports "unavailable" when neither API exists so the
 * caller can render an accessible manual-copy fallback (an always-visible
 * readonly text field, see ShareButton.tsx).
 */
export async function shareOrCopy(payload: SharePayload): Promise<ShareOutcome> {
  if (typeof navigator !== "undefined" && typeof navigator.share === "function") {
    try {
      await navigator.share(payload);
      return { status: "shared" };
    } catch (error) {
      if (error instanceof DOMException && error.name === "AbortError") {
        return { status: "cancelled" };
      }
      // Fall through to clipboard when share fails for a non-user reason.
    }
  }

  if (typeof navigator !== "undefined" && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(payload.url);
      return { status: "copied" };
    } catch {
      return { status: "unavailable" };
    }
  }

  return { status: "unavailable" };
}
