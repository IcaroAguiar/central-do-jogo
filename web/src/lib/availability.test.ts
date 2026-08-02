import { describe, expect, it } from "vitest";
import {
  availabilityLabel,
  confidenceLabel,
  isIncomplete,
  needsIncompleteDisclaimer,
} from "./availability";

describe("availabilityLabel", () => {
  it("maps every REQ-010 gap state to an explicit pt-BR label", () => {
    expect(availabilityLabel("available")).toBe("Disponível");
    expect(availabilityLabel("awaiting_publication")).toContain("Aguardando");
    expect(availabilityLabel("not_found")).toContain("Não encontrado");
    expect(availabilityLabel("divergent")).toContain("divergentes");
    expect(availabilityLabel("no_coverage")).toContain("cobertura");
  });
});

describe("isIncomplete", () => {
  it("is false only for the available state", () => {
    expect(isIncomplete("available")).toBe(false);
    expect(isIncomplete("awaiting_publication")).toBe(true);
    expect(isIncomplete("not_found")).toBe(true);
    expect(isIncomplete("divergent")).toBe(true);
    expect(isIncomplete("no_coverage")).toBe(true);
  });
});

describe("confidenceLabel", () => {
  it("maps every confidence level to a pt-BR label", () => {
    expect(confidenceLabel("high")).toContain("alta");
    expect(confidenceLabel("medium")).toContain("média");
    expect(confidenceLabel("low")).toContain("baixa");
  });
});

describe("needsIncompleteDisclaimer", () => {
  it("requires no disclaimer for an available broadcast with only high confidence", () => {
    expect(needsIncompleteDisclaimer("available", ["high"])).toBe(false);
  });

  it("requires a disclaimer when any broadcast confidence is low or medium", () => {
    expect(needsIncompleteDisclaimer("available", ["low"])).toBe(true);
    expect(needsIncompleteDisclaimer("available", ["high", "medium"])).toBe(true);
  });

  it("requires a disclaimer whenever the broadcast state itself is incomplete", () => {
    expect(needsIncompleteDisclaimer("awaiting_publication", [])).toBe(true);
  });
});
