import { test } from "@playwright/test";

/**
 * Placeholder so the admin e2e tree is runnable before TASK-035 adds real
 * maintainer journeys. Remove when the first admin spec lands.
 */
test.describe("admin suite skeleton", () => {
  test.skip(true, "TASK-035: admin Playwright coverage not implemented yet");
  test("placeholder", async () => {
    // intentionally empty
  });
});
