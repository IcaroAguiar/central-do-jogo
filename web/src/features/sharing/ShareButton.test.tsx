import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it } from "vitest";
import { ShareButton } from "./ShareButton";
import {
  restoreNavigatorShareAndClipboard,
  stubNavigatorShareAndClipboard,
} from "./testNavigatorStub";

describe("ShareButton", () => {
  afterEach(() => {
    restoreNavigatorShareAndClipboard();
  });

  it("shows an accessible manual-copy fallback when share and clipboard are both unavailable", async () => {
    const user = userEvent.setup();
    stubNavigatorShareAndClipboard({ share: undefined, clipboard: undefined });

    render(
      <ShareButton title="Flamengo" text="Agenda" url="https://example.com/clubes/flamengo" />,
    );
    await user.click(screen.getByRole("button", { name: "Compartilhar" }));

    await waitFor(() =>
      expect(screen.getByRole("textbox")).toHaveValue("https://example.com/clubes/flamengo"),
    );
  });
});
