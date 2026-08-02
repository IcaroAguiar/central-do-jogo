import { useEffect, useRef, useState } from "react";
import { ApiRequestError, fetchSearch, type SearchResponse } from "../../api/client";

export const SEARCH_DEBOUNCE_MS = 300;
export const MIN_QUERY_LENGTH = 1;

export type SearchStatus = "idle" | "loading" | "success" | "error" | "rate_limited";

export interface SearchState {
  status: SearchStatus;
  results: SearchResponse | null;
  message: string;
}

/** Debounced free-text search (REQ-005) with explicit handling for the
 * SEC-001 rate limit response so the UI never silently retries a 429. */
export function useSearch(query: string): SearchState {
  const [state, setState] = useState<SearchState>({ status: "idle", results: null, message: "" });
  const requestId = useRef(0);

  useEffect(() => {
    const trimmed = query.trim();
    if (trimmed.length < MIN_QUERY_LENGTH) {
      setState({ status: "idle", results: null, message: "" });
      return;
    }

    const currentRequest = ++requestId.current;
    const timer = window.setTimeout(() => {
      setState((prev) => ({ ...prev, status: "loading" }));
      fetchSearch(trimmed)
        .then(({ data }) => {
          if (requestId.current !== currentRequest) return;
          const total = data.clubs.length + data.matches.length;
          setState({
            status: "success",
            results: data,
            message:
              total === 0 ? "Nenhum resultado encontrado." : `${total} resultado(s) encontrado(s).`,
          });
        })
        .catch((error: unknown) => {
          if (requestId.current !== currentRequest) return;
          if (error instanceof ApiRequestError && error.status === 429) {
            setState({
              status: "rate_limited",
              results: null,
              message: "Muitas buscas em sequência. Aguarde alguns segundos e tente novamente.",
            });
            return;
          }
          setState({
            status: "error",
            results: null,
            message: "Não foi possível buscar agora. Tente novamente.",
          });
        });
    }, SEARCH_DEBOUNCE_MS);

    return () => window.clearTimeout(timer);
  }, [query]);

  return state;
}
