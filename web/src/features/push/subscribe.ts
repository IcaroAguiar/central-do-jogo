import { createPushSubscription, fetchPushVapidPublicKey } from "../../api/client";
import { urlBase64ToUint8Array } from "./consent";

/** Register the browser PushSubscription with the account-backed API. */
export async function subscribeCurrentBrowser(): Promise<void> {
  if (!("serviceWorker" in navigator) || !("PushManager" in window)) {
    throw new Error("Web Push is not supported in this browser");
  }
  const { publicKey } = await fetchPushVapidPublicKey();
  const registration = await navigator.serviceWorker.ready;
  const existing = await registration.pushManager.getSubscription();
  const subscription =
    existing ??
    (await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(publicKey) as BufferSource,
    }));
  const json = subscription.toJSON();
  if (!json.endpoint || !json.keys?.p256dh || !json.keys?.auth) {
    throw new Error("incomplete push subscription from the browser");
  }
  await createPushSubscription({
    endpoint: json.endpoint,
    keys: { p256dh: json.keys.p256dh, auth: json.keys.auth },
  });
}
