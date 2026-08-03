import { afterEach, describe, expect, it, vi } from "vitest";
import { MAX_FAVORITE_CLUBS, uniquePreserveOrder } from "./merge";
import {
  __resetPreferencesForTests,
  getFavoriteClubs,
  getPrimaryClub,
  setPrimaryClub,
} from "./preferences";
import {
  __resetSyncForTests,
  getAuthenticated,
  markAuthenticated,
  PREFS_OWNER_KEY,
  syncPreferencesWithAccount,
} from "./sync";

vi.mock("../../api/client", () => ({
  fetchAuthMe: vi.fn(),
  fetchPreferences: vi.fn(),
  putPreferences: vi.fn(),
}));

import { fetchAuthMe, fetchPreferences, putPreferences } from "../../api/client";

const fetchAuthMeMock = vi.mocked(fetchAuthMe);
const fetchPreferencesMock = vi.mocked(fetchPreferences);
const putPreferencesMock = vi.mocked(putPreferences);

describe("syncPreferencesWithAccount", () => {
  afterEach(() => {
    __resetPreferencesForTests();
    __resetSyncForTests();
    vi.clearAllMocks();
  });

  it("does not push foreign local leftovers into a different account", async () => {
    window.localStorage.setItem(PREFS_OWNER_KEY, "a@example.com");
    setPrimaryClub("flamengo");
    fetchAuthMeMock.mockResolvedValue({
      authenticated: true,
      authEnabled: true,
      email: "b@example.com",
      role: "user",
    });
    fetchPreferencesMock.mockResolvedValue({
      primaryClubSlug: null,
      favoriteClubSlugs: ["vasco"],
    });

    const outcome = await syncPreferencesWithAccount();

    expect(outcome.authenticated).toBe(true);
    expect(putPreferencesMock).not.toHaveBeenCalled();
    expect(getPrimaryClub()).toBeNull();
    expect(getFavoriteClubs()).toEqual(["vasco"]);
    expect(window.localStorage.getItem(PREFS_OWNER_KEY)).toBe("b@example.com");
  });

  it("keeps authenticated true when push fails after a successful GET", async () => {
    fetchAuthMeMock.mockResolvedValue({
      authenticated: true,
      authEnabled: true,
      email: "fan@example.com",
      role: "user",
    });
    fetchPreferencesMock.mockResolvedValue({
      primaryClubSlug: null,
      favoriteClubSlugs: [],
    });
    putPreferencesMock.mockRejectedValue(new Error("put failed"));

    const outcome = await syncPreferencesWithAccount();

    expect(outcome).toMatchObject({ authenticated: true, pushFailed: true });
    expect(putPreferencesMock).toHaveBeenCalled();
  });

  it("markAuthenticated(false) clears the remote-persist gate", () => {
    markAuthenticated(true);
    expect(getAuthenticated()).toBe(true);
    markAuthenticated(false);
    expect(getAuthenticated()).toBe(false);
  });
});

describe("uniquePreserveOrder cap", () => {
  it("stops at MAX_FAVORITE_CLUBS", () => {
    const slugs = Array.from({ length: MAX_FAVORITE_CLUBS + 5 }, (_, i) => `club-${i}`);
    expect(uniquePreserveOrder(slugs)).toHaveLength(MAX_FAVORITE_CLUBS);
  });
});
