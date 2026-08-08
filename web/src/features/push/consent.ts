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
