import { expect, test } from "@playwright/test";

/**
 * Phase 4 readiness probes that do not require live Google OAuth or VAPID
 * secrets. Full login / Push subscribe journeys remain TASK-035 residuals
 * (see docs/validation/phase-4-checklist.md).
 */
test.describe("phase 4 API readiness", () => {
  test("auth/me reports configuration without requiring a session", async ({ request }) => {
    const res = await request.get("/api/v1/auth/me");
    expect(res.ok()).toBeTruthy();
    const body = (await res.json()) as {
      authenticated: boolean;
      authEnabled: boolean;
    };
    expect(typeof body.authEnabled).toBe("boolean");
    expect(typeof body.authenticated).toBe("boolean");
  });

  test("push vapid endpoint responds according to VAPID configuration", async ({ request }) => {
    const res = await request.get("/api/v1/push/vapid-public-key");
    // Default CI/local .env leave VAPID empty → 503 push_disabled.
    // When operators enable VAPID, the same path returns 200 + publicKey.
    if (res.status() === 503) {
      const body = (await res.json()) as { error?: { code?: string } };
      expect(body.error?.code).toBe("push_disabled");
      return;
    }
    expect(res.ok()).toBeTruthy();
    const body = (await res.json()) as { enabled: boolean; publicKey: string };
    expect(body.enabled).toBe(true);
    expect(typeof body.publicKey).toBe("string");
    expect(body.publicKey.length).toBeGreaterThan(0);
  });
});
