import type { AgendaRange } from "../../api/client";

const TABS: { value: AgendaRange; label: string }[] = [
  { value: "week", label: "Semana" },
  { value: "month", label: "Mês" },
  { value: "season", label: "Temporada" },
];

export interface AgendaTabsProps {
  active: AgendaRange;
  onChange: (range: AgendaRange) => void;
}

/** Week/month/season agenda range selector (REQ-004), implemented as an
 * accessible tablist. */
export function AgendaTabs({ active, onChange }: AgendaTabsProps) {
  return (
    <div role="tablist" aria-label="Período da agenda" className="agenda-tabs">
      {TABS.map((tab) => (
        <button
          key={tab.value}
          type="button"
          role="tab"
          id={`agenda-tab-${tab.value}`}
          aria-selected={active === tab.value}
          aria-controls="agenda-panel"
          className="agenda-tabs__tab"
          onClick={() => onChange(tab.value)}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}
