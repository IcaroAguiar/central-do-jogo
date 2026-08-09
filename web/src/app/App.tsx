import { useEffect } from "react";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { AdminPage } from "../features/admin/AdminPage";
import { ClubPage } from "../features/clubs/ClubPage";
import { MatchPage } from "../features/matches/MatchPage";
import { SettingsPage } from "../features/settings/SettingsPage";
import { HomePage } from "../pages/HomePage";
import { NotFoundPage } from "../pages/NotFoundPage";
import { Layout } from "./Layout";

/**
 * PAT-004 progressive enhancement strategy (also documented in
 * web/server-templates/base.tmpl):
 *
 * Go SSR ("/", "/clubes/{slug}", "/jogos/{slug}") always renders complete,
 * indexable semantic HTML inside a #ssr-content wrapper, plus an empty
 * #root and a stable-named `/app.js` module script. The Vite SPA shell
 * (index.html, used for client-side navigation and local dev) renders
 * straight into #root with no #ssr-content.
 *
 * On mount, this component removes #ssr-content (if present) so the SSR
 * markup never duplicates the interactive UI React renders into #root, then
 * client-side routing (react-router) takes over for every subsequent
 * navigation — including back to a URL Go would otherwise SSR, since the
 * browser never re-requests it.
 */
export function App() {
  useEffect(() => {
    document.getElementById("ssr-content")?.remove();
  }, []);

  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route index element={<HomePage />} />
          <Route path="clubes/:slug" element={<ClubPage />} />
          <Route path="jogos/:slug" element={<MatchPage />} />
          <Route path="conta" element={<SettingsPage />} />
          <Route path="admin" element={<AdminPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
