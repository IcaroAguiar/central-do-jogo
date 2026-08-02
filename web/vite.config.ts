import react from "@vitejs/plugin-react";
import { VitePWA } from "vite-plugin-pwa";
import { defineConfig } from "vitest/config";

// PAT-004 (progressive enhancement): the Go SSR pages ("/", "/clubes/{slug}",
// "/jogos/{slug}") cannot know Vite's content-hashed build output at template
// parse time. We trade per-file cache-busting hashes for stable entry
// filenames (app.js / app.css) so the Go templates can link to them directly.
// Non-entry chunks (route-level code splitting, if introduced later) still
// get hashed names and are resolved at runtime by the module graph, so this
// only affects the single top-level entry.
export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      strategies: "injectManifest",
      srcDir: "src/pwa",
      filename: "sw.ts",
      injectManifest: {
        // The app shell is small and single-bundle; globbing dist is enough
        // to precache the shell for offline app-open (REQ per PWA scope).
        globPatterns: ["**/*.{js,css,html,ico,svg,woff2}"],
      },
      registerType: "autoUpdate",
      includeAssets: ["favicon.svg"],
      manifest: false,
      devOptions: {
        enabled: false,
      },
    }),
  ],
  build: {
    rollupOptions: {
      output: {
        entryFileNames: "app.js",
        assetFileNames: (assetInfo) => {
          const name = assetInfo.names?.[0] ?? "";
          if (name.endsWith(".css")) {
            return "app.css";
          }
          return "assets/[name]-[hash][extname]";
        },
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/healthz": "http://127.0.0.1:8080",
      "/api": "http://127.0.0.1:8080",
    },
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: "./src/test/setup.ts",
    exclude: ["**/node_modules/**", "**/dist/**", "../e2e/**"],
  },
});
