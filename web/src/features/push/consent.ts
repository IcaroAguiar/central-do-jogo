/**
 * Contextual Web Push consent (REQ-011): permission is only requested after
 * the visitor follows a club (favorite or primary).
 */

export const PUSH_CONSENT_DISMISSED_KEY = "cdj:pushConsentDismissed";

export function hasFollowedClub(primaryClub: string | null, favoriteClubs: string[]): boolean {
  return primaryClub !== null || favoriteClubs.length > 0;
}

export function isPushConsentDismissed(): boolean {
  try {
    return (
      typeof window !== "undefined" &&
      window.localStorage.getItem(PUSH_CONSENT_DISMISSED_KEY) === "1"
    );
  } catch {
    return false;
  }
}

export function dismissPushConsent(): void {
  try {
    if (typeof window !== "undefined") {
      window.localStorage.setItem(PUSH_CONSENT_DISMISSED_KEY, "1");
    }
  } catch {
    // ignore quota / private mode
  }
}

export function notificationPermission(): NotificationPermission | "unsupported" {
  if (typeof window === "undefined" || typeof Notification === "undefined") {
    return "unsupported";
  }
  return Notification.permission;
}

export function shouldOfferPushConsent(input: {
  primaryClub: string | null;
  favoriteClubs: string[];
  pushConfigured: boolean;
  authenticated: boolean;
}): boolean {
  if (!input.pushConfigured || !input.authenticated) return false;
  if (!hasFollowedClub(input.primaryClub, input.favoriteClubs)) return false;
  if (isPushConsentDismissed()) return false;
  const permission = notificationPermission();
  return permission === "default";
}

/** Convert a URL-safe base64 VAPID key to a Uint8Array for PushManager. */
export function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, "+").replace(/_/g, "/");
  const raw = atob(base64);
  const output = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i += 1) {
    output[i] = raw.charCodeAt(i);
  }
  return output;
}
