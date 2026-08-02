import { registerSW } from "virtual:pwa-register";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./app/App";
import "./styles.css";

const rootElement = document.getElementById("root");
if (!rootElement) {
  throw new Error("Root element #root was not found");
}

createRoot(rootElement).render(
  <StrictMode>
    <App />
  </StrictMode>,
);

// autoUpdate silently activates new service worker versions on the next
// navigation (TASK-025); no user-facing "update available" prompt is
// needed for this MVP's single-bundle app shell.
registerSW({ immediate: true });
