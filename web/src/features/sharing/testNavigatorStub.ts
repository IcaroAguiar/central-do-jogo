/**
 * jsdom (and polyfills installed by @testing-library/user-event) already
 * define `navigator.share`/`navigator.clipboard` in some test environments,
 * which makes `vi.stubGlobal("navigator", { ...navigator, clipboard: undefined })`
 * unreliable (the replacement object is not always what code under test
 * observes). Redefining the two properties directly on the real `navigator`
 * object, and restoring their original descriptors afterwards, is robust
 * regardless of what installed them first.
 */
let originalShare: PropertyDescriptor | undefined;
let originalClipboard: PropertyDescriptor | undefined;
let captured = false;

export function stubNavigatorShareAndClipboard(overrides: {
  share?: unknown;
  clipboard?: unknown;
}): void {
  if (!captured) {
    originalShare = Object.getOwnPropertyDescriptor(navigator, "share");
    originalClipboard = Object.getOwnPropertyDescriptor(navigator, "clipboard");
    captured = true;
  }
  Object.defineProperty(navigator, "share", { value: overrides.share, configurable: true });
  Object.defineProperty(navigator, "clipboard", { value: overrides.clipboard, configurable: true });
}

export function restoreNavigatorShareAndClipboard(): void {
  if (!captured) return;
  if (originalShare) {
    Object.defineProperty(navigator, "share", originalShare);
  } else {
    Reflect.deleteProperty(navigator, "share");
  }
  if (originalClipboard) {
    Object.defineProperty(navigator, "clipboard", originalClipboard);
  } else {
    Reflect.deleteProperty(navigator, "clipboard");
  }
}
