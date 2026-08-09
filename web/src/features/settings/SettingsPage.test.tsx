import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { SettingsPage } from "./SettingsPage";

vi.mock("../../api/client", () => ({
  fetchAuthMe: vi.fn(),
  fetchPrivacyExport: vi.fn(),
  deletePrivacyAccount: vi.fn(),
}));

import { fetchAuthMe } from "../../api/client";

describe("SettingsPage", () => {
  beforeEach(() => {
    vi.mocked(fetchAuthMe).mockReset();
  });

  it("prompts login when unauthenticated", async () => {
    vi.mocked(fetchAuthMe).mockResolvedValue({ authenticated: false, authEnabled: true });
    render(
      <MemoryRouter>
        <SettingsPage />
      </MemoryRouter>,
    );
    expect(await screen.findByText(/Entre para exportar/i)).toBeInTheDocument();
  });

  it("shows export and delete when authenticated", async () => {
    vi.mocked(fetchAuthMe).mockResolvedValue({
      authenticated: true,
      authEnabled: true,
      email: "a@example.com",
      displayName: "Ana",
      role: "user",
    });
    render(
      <MemoryRouter>
        <SettingsPage />
      </MemoryRouter>,
    );
    expect(await screen.findByRole("button", { name: /Baixar meus dados/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Excluir conta/i })).toBeInTheDocument();
  });
});
