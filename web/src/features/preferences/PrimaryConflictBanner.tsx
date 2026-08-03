import { usePreferences } from "./usePreferences";

/** Explicit primary-club conflict resolver so sync never silently overwrites (REQ-006). */
export function PrimaryConflictBanner() {
  const { primaryConflict, resolvePrimaryConflict } = usePreferences();
  if (!primaryConflict) {
    return null;
  }

  return (
    <div className="prefs-conflict" role="status">
      <p>
        Seu clube principal local ({primaryConflict.local}) difere do salvo na conta (
        {primaryConflict.remote}). Escolha qual manter — não sobrescrevemos em silêncio.
      </p>
      <div className="prefs-conflict-actions">
        <button type="button" onClick={() => resolvePrimaryConflict(primaryConflict.local)}>
          Manter {primaryConflict.local}
        </button>
        <button type="button" onClick={() => resolvePrimaryConflict(primaryConflict.remote)}>
          Usar {primaryConflict.remote}
        </button>
      </div>
    </div>
  );
}
