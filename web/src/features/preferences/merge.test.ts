import { describe, expect, it } from "vitest";
import { mergePreferences, uniquePreserveOrder } from "./merge";

describe("mergePreferences", () => {
  it("unions favorites without silent primary overwrite when both sides agree", () => {
    const result = mergePreferences(
      { primaryClub: "flamengo", favoriteClubs: ["vasco"] },
      { primaryClub: "flamengo", favoriteClubs: ["botafogo", "vasco"] },
    );
    expect(result.primaryClub).toBe("flamengo");
    expect(result.favoriteClubs).toEqual(["vasco", "botafogo"]);
    expect(result.primaryConflict).toBeNull();
  });

  it("adopts remote primary when local is unset", () => {
    const result = mergePreferences(
      { primaryClub: null, favoriteClubs: [] },
      { primaryClub: "santos", favoriteClubs: ["santos"] },
    );
    expect(result.primaryClub).toBe("santos");
    expect(result.primaryConflict).toBeNull();
  });

  it("keeps local primary and surfaces conflict when both sides differ", () => {
    const result = mergePreferences(
      { primaryClub: "flamengo", favoriteClubs: ["flamengo"] },
      { primaryClub: "vasco", favoriteClubs: ["vasco"] },
    );
    expect(result.primaryClub).toBe("flamengo");
    expect(result.favoriteClubs).toEqual(["flamengo", "vasco"]);
    expect(result.primaryConflict).toEqual({ local: "flamengo", remote: "vasco" });
  });
});

describe("uniquePreserveOrder", () => {
  it("deduplicates while keeping first occurrence", () => {
    expect(uniquePreserveOrder(["a", "b", "a", " ", "c"])).toEqual(["a", "b", "c"]);
  });
});
