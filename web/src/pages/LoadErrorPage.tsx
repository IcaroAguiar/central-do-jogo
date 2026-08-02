export interface LoadErrorPageProps {
  message?: string;
  onRetry?: () => void;
}

/** Shown for network / non-404 API failures (distinct from NotFoundPage). */
export function LoadErrorPage({ message, onRetry }: LoadErrorPageProps) {
  return (
    <div className="load-error" role="alert">
      <h1>Não foi possível carregar</h1>
      <p>{message ?? "Verifique sua conexão e tente novamente."}</p>
      {onRetry ? (
        <button type="button" onClick={onRetry}>
          Tentar de novo
        </button>
      ) : null}
    </div>
  );
}
