import { useState } from "react";
import { Link } from "react-router-dom";
import { SearchBox } from "../features/search/SearchBox";
import { readInitialData } from "../lib/initialData";

interface HomeClubLink {
  slug: string;
  name: string;
}

interface HomeInitialData {
  clubs?: HomeClubLink[];
}

export function HomePage() {
  const [initial] = useState(() => readInitialData<"home", HomeInitialData>("home"));
  const clubs = initial?.clubs ?? [];

  return (
    <div className="home">
      <p className="eyebrow">Pré-jogo · Brasil</p>
      <h1>Central do Jogo</h1>
      <p className="lede">
        Onde assistir, escalações oficiais e notícias do futebol brasileiro, com proveniência e
        confiança explícitas.
      </p>
      <SearchBox />
      {clubs.length > 0 ? (
        <section aria-labelledby="clubs-heading" className="home-clubs">
          <h2 id="clubs-heading">Clubes suportados</h2>
          <ul>
            {clubs.map((club) => (
              <li key={club.slug}>
                <Link to={`/clubes/${club.slug}`}>{club.name}</Link>
              </li>
            ))}
          </ul>
        </section>
      ) : null}
    </div>
  );
}
