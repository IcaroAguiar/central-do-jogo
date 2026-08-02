import { Link } from "react-router-dom";

export interface NotFoundPageProps {
  message?: string;
}

export function NotFoundPage({ message }: NotFoundPageProps) {
  return (
    <div className="not-found">
      <h1>Página não encontrada</h1>
      <p>{message ?? "Não encontramos o que você procurava."}</p>
      <Link to="/">Voltar para a página inicial</Link>
    </div>
  );
}
