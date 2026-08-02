import { describe, expect, it } from "vitest";
import { isPrivateApiPath, isPublicApiPath } from "./apiCachePolicy";

describe("apiCachePolicy", () => {
  it("keeps auth session routes out of the public API cache", () => {
    expect(isPrivateApiPath("/api/v1/auth/me")).toBe(true);
    expect(isPrivateApiPath("/api/v1/auth/google/start")).toBe(true);
    expect(isPrivateApiPath("/api/v1/auth/google/callback")).toBe(true);
    expect(isPublicApiPath("/api/v1/auth/me")).toBe(false);
  });

  it("still caches public read journeys", () => {
    expect(isPublicApiPath("/api/v1/search")).toBe(true);
    expect(isPublicApiPath("/api/v1/clubs/flamengo")).toBe(true);
    expect(isPublicApiPath("/api/v1/matches/foo")).toBe(true);
  });
});
