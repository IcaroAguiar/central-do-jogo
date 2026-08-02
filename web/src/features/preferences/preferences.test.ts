import { afterEach, describe, expect, it } from "vitest";
import {
  __resetPreferencesForTests,
  FAVORITE_CLUBS_KEY,
  getFavoriteClubs,
  getPrimaryClub,
  PRIMARY_CLUB_KEY,
  setPrimaryClub,
  toggleFavoriteClub,
} from "./preferences";

describe("preferences", () => {
  afterEach(() => {
    __resetPreferencesForTests();
  });

  it("round-trips the primary club through localStorage", () => {
    expect(getPrimaryClub()).toBeNull();

    setPrimaryClub("flamengo");
    expect(getPrimaryClub()).toBe("flamengo");
    expect(window.localStorage.getItem(PRIMARY_CLUB_KEY)).toBe("flamengo");

    setPrimaryClub(null);
    expect(getPrimaryClub()).toBeNull();
    expect(window.localStorage.getItem(PRIMARY_CLUB_KEY)).toBeNull();
  });

  it("round-trips favorite clubs through localStorage", () => {
    expect(getFavoriteClubs()).toEqual([]);

    toggleFavoriteClub("vasco");
    expect(getFavoriteClubs()).toEqual(["vasco"]);
    expect(JSON.parse(window.localStorage.getItem(FAVORITE_CLUBS_KEY) ?? "[]")).toEqual(["vasco"]);

    toggleFavoriteClub("flamengo");
    expect(getFavoriteClubs().slice().sort()).toEqual(["flamengo", "vasco"]);

    toggleFavoriteClub("vasco");
    expect(getFavoriteClubs()).toEqual(["flamengo"]);
  });
});
