import { useCallback, useEffect, useRef, useState } from "react";
import { reportFreshFetch, reportServedFromCache } from "../pwa/offlineSignal";
import type { ApiRequestError, ApiResult } from "./client";

export interface ApiResourceState<T> {
  status: "loading" | "success" | "error";
  data: T | null;
  error: ApiRequestError | Error | null;
  cachedAt: string | null;
  /** True when the last successful response came from the service worker
   * cache (offline / degraded network) rather than a live network fetch. */
  isStale: boolean;
  retry: () => void;
}

/**
 * Runs `fetcher` whenever `deps` change and exposes loading/success/error
 * state plus SW-cache metadata (REQ: offline banner with cachedAt + retry).
 */
export function useApiResource<T>(
  fetcher: () => Promise<ApiResult<T>>,
  deps: readonly unknown[],
  /** When provided, seeds the resource as already-successful (e.g. from
   * PAT-004 SSR #initial-data) and skips the very first network fetch,
   * while still fetching normally on later dependency changes. */
  initialValue?: T,
): ApiResourceState<T> {
  const [state, setState] = useState<Omit<ApiResourceState<T>, "retry">>(() =>
    initialValue !== undefined
      ? { status: "success", data: initialValue, error: null, cachedAt: null, isStale: false }
      : { status: "loading", data: null, error: null, cachedAt: null, isStale: false },
  );
  const [attempt, setAttempt] = useState(0);
  const skipNextFetch = useRef(initialValue !== undefined);

  const load = useCallback(() => {
    setAttempt((n) => n + 1);
  }, []);

  // biome-ignore lint/correctness/useExhaustiveDependencies: caller supplies its own dependency array for this generic hook; `attempt` intentionally forces a re-run on retry().
  useEffect(() => {
    // SSR #initial-data seeds the UI, but we still fire a background fetch so
    // the service worker can warm GET /api/v1/* for bookmark/shared-link
    // entry paths (otherwise offline reload of SSR URLs has no API cache).
    if (skipNextFetch.current) {
      skipNextFetch.current = false;
      void fetcher().catch(() => {
        // Warming is best-effort; UI already has SSR data.
      });
      return;
    }

    let cancelled = false;
    setState((prev) => ({ ...prev, status: "loading" }));

    fetcher()
      .then((result) => {
        if (cancelled) return;
        if (result.cachedAt !== null) {
          reportServedFromCache(result.cachedAt);
        } else {
          reportFreshFetch();
        }
        setState({
          status: "success",
          data: result.data,
          error: null,
          cachedAt: result.cachedAt,
          isStale: result.cachedAt !== null,
        });
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        setState((prev) => ({
          status: "error",
          data: prev.data,
          error: error instanceof Error ? error : new Error(String(error)),
          cachedAt: prev.cachedAt,
          isStale: prev.isStale,
        }));
      });

    return () => {
      cancelled = true;
    };
  }, [...deps, attempt]);

  return { ...state, retry: load };
}
