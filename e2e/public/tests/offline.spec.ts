import { expect, test } from "@playwright/test";

/**
 * TEST-010: after the service worker has installed and a club page has been
 * viewed online, the same page should still render with an explicit offline
 * banner instead of a blank/broken screen.
 *
 * Note: the club page must be reached via client-side navigation (not a
 * fresh `page.goto`) to populate the service worker's API cache. A direct
 * SSR load embeds the club data in `#initial-data` (PAT-004) and skips the
 * `/api/v1/clubs/{slug}` round-trip entirely (see
 * web/src/lib/initialData.ts), so it never gets written to the SW's API
 * cache and offline reloads would have nothing to fall back to.
 *
 * Also note: Chromium's `navigator.onLine` does not reliably flip to
 * `false` under Playwright's `context.setOffline()` (it only fails the
 * underlying network requests), so the app derives its offline banner from
 * observed cache-fallback responses instead — see
 * web/src/pwa/offlineSignal.ts.
 */
test.describe("offline resilience", () => {
  test("shows cached club data and an offline banner without a network connection", async ({
    page,
    context,
  }) => {
    await page.goto("/");
    await page.waitForFunction(
      () => !("serviceWorker" in navigator) || navigator.serviceWorker.controller !== null,
    );

    const search = page.getByRole("combobox", { name: "Buscar clube ou partida" });
    await search.fill("fl");
    await page.getByRole("option").first().click();
    await expect(page).toHaveURL(/\/clubes\//);
    await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
    await page.waitForLoadState("networkidle");

    await context.setOffline(true);
    try {
      await page.reload();
      await expect(page.getByText("Você está offline.")).toBeVisible();
      await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
    } finally {
      await context.setOffline(false);
    }
  });
});
