import { Link, Outlet } from "react-router-dom";
import { AuthMenu } from "../features/auth/AuthMenu";
import { BRASILIA_LABEL } from "../lib/datetime";
import { OfflineBanner } from "../pwa/OfflineBanner";

/** Shared shell for every route: skip-link (a11y), primary nav, the
 * <main> landmark routed content mounts into, and a footer note pinning the
 * product's reference timezone (CON-008). */
export function Layout() {
  return (
    <div className="app-shell">
      <a className="skip-link" href="#main-content">
        Pular para o conteúdo
      </a>
      <header className="app-header">
        <nav aria-label="Navegação principal">
          <Link to="/" className="app-brand">
            Central do Jogo
          </Link>
          <AuthMenu />
        </nav>
      </header>
      <OfflineBanner />
      <main id="main-content" tabIndex={-1}>
        <Outlet />
      </main>
      <footer className="app-footer">
        <p>Todos os horários exibidos em {BRASILIA_LABEL} (America/Sao_Paulo).</p>
      </footer>
    </div>
  );
}
