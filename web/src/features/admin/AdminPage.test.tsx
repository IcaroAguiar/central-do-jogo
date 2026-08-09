import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { AdminPage } from "./AdminPage";

vi.mock("../../api/client", () => ({
  fetchAuthMe: vi.fn(),
  fetchAdminSourceHealth: vi.fn(),
  fetchAdminAtRiskMatches: vi.fn(),
  fetchAdminAudit: vi.fn(),
  fetchAdminReports: vi.fn(),
  postAdminMatchAction: vi.fn(),
  reviewAdminReport: vi.fn(),
}));

import { fetchAuthMe } from "../../api/client";

describe("AdminPage", () => {
  beforeEach(() => {
    vi.mocked(fetchAuthMe).mockReset();
  });

  it("blocks non-maintainers", async () => {
    vi.mocked(fetchAuthMe).mockResolvedValue({
      authenticated: true,
      authEnabled: true,
      role: "user",
    });
    render(
      <MemoryRouter>
        <AdminPage />
      </MemoryRouter>,
    );
    expect(await screen.findByText(/Acesso restrito a mantenedores/i)).toBeInTheDocument();
  });
});
