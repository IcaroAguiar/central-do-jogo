import { expect, test } from "@playwright/test";

/**
 * End-to-end happy path across the three public read journeys plus sharing
 * (TEST-008 partial: no Push notifications in Phase 3): search → club
 * agenda → match detail → share.
 */
test.describe("public smoke: search → club → match → share", () => {
  test("finds a club via search, opens its agenda, opens a match, and can share it", async ({
    page,
    context,
  }) => {
    await page.goto("/");

    const search = page.getByRole("combobox", { name: "Buscar clube ou partida" });
    await search.fill("fl");

    const firstOption = page.getByRole("option").first();
    await expect(firstOption).toBeVisible();
    await firstOption.click();

    await expect(page).toHaveURL(/\/clubes\//);
    await expect(page.getByRole("heading", { level: 1 })).toBeVisible();

    const firstMatchLink = page.locator(".agenda-list a").first();
    const matchCount = await firstMatchLink.count();
    test.skip(matchCount === 0, "seeded club has no upcoming matches in this environment");

    await firstMatchLink.click();
    await expect(page).toHaveURL(/\/jogos\//);
    await expect(page.getByRole("heading", { name: "Onde assistir" })).toBeVisible();

    await context.grantPermissions(["clipboard-read", "clipboard-write"]);
    await page.getByRole("button", { name: "Compartilhar" }).click();

    await expect(page.getByRole("status").filter({ hasText: /copiado|Compartilhado/i })).toBeVisible();
  });
});
