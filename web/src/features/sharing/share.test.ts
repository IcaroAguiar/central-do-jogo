import { afterEach, describe, expect, it, vi } from "vitest";
import { shareOrCopy } from "./share";
import {
  restoreNavigatorShareAndClipboard,
  stubNavigatorShareAndClipboard,
} from "./testNavigatorStub";

const payload = {
  title: "Flamengo x Vasco",
  text: "Onde assistir",
  url: "https://example.com/jogos/x",
};

describe("shareOrCopy", () => {
  afterEach(() => {
    restoreNavigatorShareAndClipboard();
  });

  it("uses the Web Share API when available", async () => {
    const share = vi.fn().mockResolvedValue(undefined);
    stubNavigatorShareAndClipboard({ share, clipboard: undefined });

    const outcome = await shareOrCopy(payload);

    expect(share).toHaveBeenCalledWith(payload);
    expect(outcome).toEqual({ status: "shared" });
  });

  it("reports cancelled when the user dismisses the native share sheet", async () => {
    const share = vi.fn().mockRejectedValue(new DOMException("aborted", "AbortError"));
    stubNavigatorShareAndClipboard({ share, clipboard: undefined });

    const outcome = await shareOrCopy(payload);

    expect(outcome).toEqual({ status: "cancelled" });
  });

  it("falls back to clipboard copy when Web Share is unavailable", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    stubNavigatorShareAndClipboard({ share: undefined, clipboard: { writeText } });

    const outcome = await shareOrCopy(payload);

    expect(writeText).toHaveBeenCalledWith(payload.url);
    expect(outcome).toEqual({ status: "copied" });
  });

  it("reports unavailable when neither Web Share nor clipboard exist", async () => {
    stubNavigatorShareAndClipboard({ share: undefined, clipboard: undefined });

    const outcome = await shareOrCopy(payload);

    expect(outcome).toEqual({ status: "unavailable" });
  });
});
