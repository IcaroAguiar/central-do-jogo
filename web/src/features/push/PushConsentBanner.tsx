import { useEffect, useState } from "react";
import { fetchPushVapidPublicKey } from "../../api/client";
import { getAuthenticated, subscribeSync } from "../preferences/sync";
import { usePreferences } from "../preferences/usePreferences";
import { dismissPushConsent, shouldOfferPushConsent } from "./consent";
import { subscribeCurrentBrowser } from "./subscribe";

/** Contextual banner: ask for notification permission only after following a club (REQ-011). */
export function PushConsentBanner() {
  const prefs = usePreferences();
  const [authenticated, setAuthenticated] = useState(getAuthenticated);
  const [pushConfigured, setPushConfigured] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [visible, setVisible] = useState(false);

  useEffect(() => subscribeSync(() => setAuthenticated(getAuthenticated())), []);

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      try {
        await fetchPushVapidPublicKey();
        if (!cancelled) setPushConfigured(true);
      } catch {
        if (!cancelled) setPushConfigured(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    setVisible(
      shouldOfferPushConsent({
        primaryClub: prefs.primaryClub,
        favoriteClubs: prefs.favoriteClubs,
        pushConfigured,
        authenticated,
      }),
    );
  }, [prefs.primaryClub, prefs.favoriteClubs, pushConfigured, authenticated]);

  if (!visible) {
    return null;
  }

  async function enable() {
    setBusy(true);
    setError(null);
    try {
      const permission = await Notification.requestPermission();
      if (permission !== "granted") {
        dismissPushConsent();
        setVisible(false);
        return;
      }
      await subscribeCurrentBrowser();
      dismissPushConsent();
      setVisible(false);
    } catch {
      setError("Não foi possível ativar os alertas agora. Tente de novo em instantes.");
    } finally {
      setBusy(false);
    }
  }

  function later() {
    dismissPushConsent();
    setVisible(false);
  }

  return (
    <div className="push-consent" role="status">
      <p>
        Quer alertas de transmissão e escalação oficial deste clube? Pedimos permissão só depois que
        você segue um time.
      </p>
      {error ? <p className="push-consent__error">{error}</p> : null}
      <div className="push-consent__actions">
        <button type="button" onClick={() => void enable()} disabled={busy}>
          Ativar alertas
        </button>
        <button type="button" className="push-consent__later" onClick={later} disabled={busy}>
          Agora não
        </button>
      </div>
    </div>
  );
}
