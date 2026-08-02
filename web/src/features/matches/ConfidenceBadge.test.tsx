import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { ConfidenceLevel } from "../../api/client";
import { ConfidenceBadge } from "./ConfidenceBadge";

describe("ConfidenceBadge", () => {
  it.each<[ConfidenceLevel, string]>([
    ["high", "Confiança alta"],
    ["medium", "Confiança média"],
    ["low", "Confiança baixa"],
  ])("renders the %s confidence as %s", (level, label) => {
    render(<ConfidenceBadge level={level} />);
    expect(screen.getByText(label)).toBeInTheDocument();
  });
});
