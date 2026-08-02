import { formatTimestamp } from "../lib/datetime";
import { useOfflineSignal } from "./useOfflineSignal";
import { useOnlineStatus } from "./useOnlineStatus";

/** Global offline indicator (TASK-025): visible whenever the browser
 * reports it is offline, OR the app has directly observed a request being
 * served from the service worker's cache fallback (see
 * web/src/pwa/offlineSignal.ts for why `navigator.onLine` alone is not
 * trustworthy). Shows how stale the visible data is and a retry action that
 * simply reloads to re-attempt the network. */
export function OfflineBanner() {
  const isOnline = useOnlineStatus();
  const degraded = useOfflineSignal();
  if (isOnline && !degraded.isDegraded) {
    return null;
  }

  const cachedAtLabel = formatTimestamp(degraded.cachedAt);

  return (
    <div role="status" aria-live="assertive" className="offline-banner">
      <p>
        Você está offline.{" "}
        {cachedAtLabel
          ? `Mostrando dados salvos em ${cachedAtLabel}.`
          : "Alguns dados podem estar desatualizados."}
      </p>
      <button type="button" onClick={() => window.location.reload()}>
        Tentar novamente
      </button>
    </div>
  );
}
