import { type KeyboardEvent, useId, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { formatKickoff } from "../../lib/datetime";
import { useSearch } from "./useSearch";

interface SearchOption {
  id: string;
  label: string;
  description?: string;
  href: string;
}

/** Free-text search combobox (REQ-005): debounced fetch, full keyboard
 * navigation (↑/↓/Enter/Esc), aria-live status, and explicit SEC-001
 * rate-limit messaging. */
export function SearchBox() {
  const [query, setQuery] = useState("");
  const [isOpen, setIsOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(-1);
  const navigate = useNavigate();
  const listboxId = useId();
  const state = useSearch(query);

  const options = useMemo<SearchOption[]>(() => {
    if (!state.results) return [];
    const clubOptions: SearchOption[] = state.results.clubs.map((club) => ({
      id: `club-${club.slug}`,
      label: club.name,
      description: "Clube",
      href: `/clubes/${club.slug}`,
    }));
    const matchOptions: SearchOption[] = state.results.matches.map((match) => ({
      id: `match-${match.slug}`,
      label: `${match.homeClub.shortName} x ${match.awayClub.shortName}`,
      description: formatKickoff(match.kickoffAt),
      href: `/jogos/${match.slug}`,
    }));
    return [...clubOptions, ...matchOptions];
  }, [state.results]);

  const showListbox = isOpen && query.trim().length > 0;

  function selectOption(option: SearchOption) {
    setIsOpen(false);
    setActiveIndex(-1);
    setQuery("");
    navigate(option.href);
  }

  function handleKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (!showListbox || options.length === 0) {
      return;
    }
    switch (event.key) {
      case "ArrowDown":
        event.preventDefault();
        setActiveIndex((index) => (index + 1) % options.length);
        break;
      case "ArrowUp":
        event.preventDefault();
        setActiveIndex((index) => (index <= 0 ? options.length - 1 : index - 1));
        break;
      case "Enter":
        if (activeIndex >= 0 && activeIndex < options.length) {
          event.preventDefault();
          selectOption(options[activeIndex]);
        }
        break;
      case "Escape":
        setIsOpen(false);
        setActiveIndex(-1);
        break;
      default:
        break;
    }
  }

  const activeOptionId =
    activeIndex >= 0 && options[activeIndex]
      ? `${listboxId}-${options[activeIndex].id}`
      : undefined;

  return (
    <div className="search-box">
      <label htmlFor={`${listboxId}-input`} className="search-box__label">
        Buscar clube ou partida
      </label>
      <input
        id={`${listboxId}-input`}
        type="text"
        role="combobox"
        aria-expanded={showListbox}
        aria-controls={listboxId}
        aria-activedescendant={activeOptionId}
        aria-autocomplete="list"
        autoComplete="off"
        value={query}
        placeholder="Ex.: Flamengo, Vasco x Fluminense…"
        onChange={(event) => {
          setQuery(event.target.value);
          setIsOpen(true);
          setActiveIndex(-1);
        }}
        onFocus={() => setIsOpen(true)}
        onKeyDown={handleKeyDown}
      />
      <p role="status" aria-live="polite" className="search-box__status">
        {state.status === "loading" ? "Buscando…" : state.message}
      </p>
      {showListbox && options.length > 0 ? (
        <div id={listboxId} role="listbox" className="search-box__listbox">
          {options.map((option, index) => (
            // biome-ignore lint/a11y/useFocusableInteractive: options are intentionally not focusable; the input keeps DOM focus and tracks the active option via aria-activedescendant (WAI-ARIA combobox pattern).
            <div
              key={option.id}
              id={`${listboxId}-${option.id}`}
              role="option"
              aria-selected={index === activeIndex}
              className={index === activeIndex ? "is-active" : undefined}
            >
              <button type="button" tabIndex={-1} onClick={() => selectOption(option)}>
                <span>{option.label}</span>
                {option.description ? (
                  <span className="search-box__option-meta">{option.description}</span>
                ) : null}
              </button>
            </div>
          ))}
        </div>
      ) : null}
    </div>
  );
}
