import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ReportForm } from "./ReportForm";

vi.mock("../../api/client", () => ({
  createReport: vi.fn(),
}));

describe("ReportForm", () => {
  it("renders contextual report controls", () => {
    render(<ReportForm contextType="match" contextSlug="flamengo-x-vasco" />);
    expect(screen.getByRole("heading", { name: /Relatar um erro/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/Descreva o problema/i)).toBeInTheDocument();
  });
});
