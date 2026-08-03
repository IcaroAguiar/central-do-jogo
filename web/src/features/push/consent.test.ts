import { describe, expect, it } from "vitest";
import { hasFollowedClub, shouldOfferPushConsent, urlBase64ToUint8Array } from "./consent";

describe("push consent", () => {
  it("requires a followed club before offering permission", () => {
    expect(hasFollowedClub(null, [])).toBe(false);
    expect(hasFollowedClub("flamengo", [])).toBe(true);
    expect(hasFollowedClub(null, ["vasco"])).toBe(true);
  });

  it("offers consent only when configured, authenticated, followed, and default permission", () => {
    expect(
      shouldOfferPushConsent({
        primaryClub: "flamengo",
        favoriteClubs: [],
        pushConfigured: true,
        authenticated: true,
      }),
    ).toBe(typeof Notification !== "undefined" && Notification.permission === "default");
  });

  it("decodes url-safe base64 VAPID keys", () => {
    const bytes = urlBase64ToUint8Array("AQID");
    expect(Array.from(bytes)).toEqual([1, 2, 3]);
  });
});
