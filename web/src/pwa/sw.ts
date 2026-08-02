/// <reference lib="webworker" />

/**
 * Custom PWA service worker (TASK-025), built via vite-plugin-pwa's
 * `injectManifest` strategy (see web/vite.config.ts). We write plain Cache
 * API logic instead of pulling in the Workbox runtime, since the caching
 * rules here are simple and this keeps the shipped worker small and easy to
 * audit end to end:
 *
 *   - Precache the built app shell (self.__WB_MANIFEST, filled in at build
 *     time by vite-plugin-pwa) so the app can open offline.
 *   - Cache successful GET /api/v1/* responses (public read journeys only)
 *     with a timestamp header, network-first, falling back to cache when
 *     offline so the UI can show "dados salvos em <timestamp>".
 *   - Never intercept /api/v1/auth/*, any export endpoint, /api/*admin*, or
 *     /api/*privacy* — those must always hit the network directly.
 */
import { isPrivateApiPath, isPublicApiPath } from "./apiCachePolicy";

declare let self: ServiceWorkerGlobalScope & {
  __WB_MANIFEST: Array<{ url: string; revision: string | null }>;
};

const PRECACHE_NAME = "cdj-precache-v1";
const API_CACHE_NAME = "cdj-api-v1";
export const CACHED_AT_HEADER = "x-cdj-cached-at";

// De-duplicated because vite-plugin-pwa's manifest can list the same static
// asset twice (once from `includeAssets`, once from the dist glob); `Cache
// .addAll()` throws on duplicate URLs.
const PRECACHE_URLS = Array.from(new Set(self.__WB_MANIFEST.map((entry) => entry.url)));


self.addEventListener("install", (event) => {
  event.waitUntil(
    caches
      .open(PRECACHE_NAME)
      .then((cache) => cache.addAll(PRECACHE_URLS))
      .then(() => self.skipWaiting()),
  );
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) =>
        Promise.all(
          keys
            .filter((key) => key !== PRECACHE_NAME && key !== API_CACHE_NAME)
            .map((key) => caches.delete(key)),
        ),
      )
      .then(() => self.clients.claim()),
  );
});

async function withCachedAtHeader(response: Response): Promise<Response> {
  const headers = new Headers(response.headers);
  headers.set(CACHED_AT_HEADER, new Date().toISOString());
  const body = await response.clone().arrayBuffer();
  return new Response(body, { status: response.status, statusText: response.statusText, headers });
}

async function networkFirstApi(request: Request): Promise<Response> {
  const cache = await caches.open(API_CACHE_NAME);
  try {
    const response = await fetch(request);
    if (response.ok) {
      await cache.put(request, await withCachedAtHeader(response.clone()));
    }
    return response;
  } catch (error) {
    const cached = await cache.match(request);
    if (cached) {
      return cached;
    }
    throw error;
  }
}

// The Go static file server 301-redirects GET /index.html to GET / (standard
// net/http canonicalization), so the cached Response for it carries
// `redirected: true`. Chromium refuses to fulfill a "navigate" fetch event
// with a redirected Response (it fails the navigation with net::ERR_FAILED
// instead of rendering it), so we rebuild a plain, non-redirected Response
// with the same body/status/headers before using it as an offline fallback.
async function stripRedirectedFlag(response: Response): Promise<Response> {
  if (!response.redirected) {
    return response;
  }
  const body = await response.blob();
  return new Response(body, {
    status: response.status,
    statusText: response.statusText,
    headers: response.headers,
  });
}

async function networkFirstNavigation(request: Request): Promise<Response> {
  try {
    return await fetch(request);
  } catch {
    const precache = await caches.open(PRECACHE_NAME);
    const shell = (await precache.match("/index.html")) ?? (await precache.match("/"));
    if (shell) {
      return stripRedirectedFlag(shell);
    }
    throw new Error("offline and no cached app shell available");
  }
}

async function cacheFirstAsset(request: Request): Promise<Response> {
  const cached = await caches.match(request);
  if (cached) {
    return cached;
  }
  return fetch(request);
}

self.addEventListener("fetch", (event) => {
  const request = event.request;
  if (request.method !== "GET") {
    return;
  }

  const url = new URL(request.url);
  if (isPrivateApiPath(url.pathname)) {
    return;
  }

  if (isPublicApiPath(url.pathname)) {
    event.respondWith(networkFirstApi(request));
    return;
  }

  if (request.mode === "navigate") {
    event.respondWith(networkFirstNavigation(request));
    return;
  }

  if (url.origin === self.location.origin) {
    event.respondWith(cacheFirstAsset(request));
  }
});
