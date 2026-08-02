import { defineConfig, devices } from "@playwright/test";

/**
 * Public-journey smoke suite (TEST-008 partial, without Push notifications;
 * TEST-010 offline). Runs against an already-running full stack (Go server
 * + Postgres + seeded data + built web assets) since these journeys need
 * real SSR HTML, real API responses, and a real installed service worker —
 * see e2e/public/README.md for how to bring that stack up locally.
 */
export default defineConfig({
  testDir: "./tests",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: [["list"]],
  use: {
    baseURL: process.env.E2E_BASE_URL ?? "http://127.0.0.1:8080",
    trace: "on-first-retry",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
