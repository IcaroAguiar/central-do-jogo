import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../../api/client", () => {
  class ApiRequestError extends Error {
    status: number;
    code: string;
    constructor(status: number, code: string, message: string) {
      super(message);
      this.status = status;
      this.code = code;
    }
  }
  return { ApiRequestError, fetchSearch: vi.fn() };
});

import { fetchSearch } from "../../api/client";
import { SearchBox } from "./SearchBox";

const mockedFetchSearch = vi.mocked(fetchSearch);

describe("SearchBox", () => {
  beforeEach(() => {
    mockedFetchSearch.mockReset();
  });

  it("supports arrow-key navigation and Enter to select a result", async () => {
    mockedFetchSearch.mockResolvedValue({
      data: {
        query: "fla",
        clubs: [{ slug: "flamengo", name: "Flamengo", shortName: "FLA" }],
        matches: [],
      },
      cachedAt: null,
    });

    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <SearchBox />
      </MemoryRouter>,
    );

    await user.type(screen.getByRole("combobox"), "fla");

    const option = await screen.findByRole("option", { name: /Flamengo/ });
    expect(option).toHaveAttribute("aria-selected", "false");

    await user.keyboard("{ArrowDown}");
    expect(option).toHaveAttribute("aria-selected", "true");

    await user.keyboard("{Enter}");

    await waitFor(() => expect(screen.queryByRole("listbox")).not.toBeInTheDocument());
  });

  it("closes the listbox on Escape without clearing the typed query", async () => {
    mockedFetchSearch.mockResolvedValue({
      data: {
        query: "fla",
        clubs: [{ slug: "flamengo", name: "Flamengo", shortName: "FLA" }],
        matches: [],
      },
      cachedAt: null,
    });

    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <SearchBox />
      </MemoryRouter>,
    );

    const input = screen.getByRole("combobox");
    await user.type(input, "fla");
    await screen.findByRole("option", { name: /Flamengo/ });

    await user.keyboard("{Escape}");

    await waitFor(() => expect(screen.queryByRole("listbox")).not.toBeInTheDocument());
    expect(input).toHaveValue("fla");
  });
});
