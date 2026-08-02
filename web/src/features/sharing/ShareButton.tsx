import { useId, useState } from "react";
import { shareOrCopy } from "./share";

export interface ShareButtonProps {
  title: string;
  text: string;
  url: string;
}

type Status = "idle" | "shared" | "copied" | "unavailable";

const STATUS_MESSAGES: Record<Exclude<Status, "idle">, string> = {
  shared: "Compartilhado.",
  copied: "Link copiado para a área de transferência.",
  unavailable: "Não foi possível compartilhar automaticamente. Copie o link abaixo.",
};

/** Share button with an accessible fallback: when neither Web Share nor
 * clipboard write succeeds, a readonly text field with the link appears so
 * the link can always be copied manually (keyboard/screen-reader friendly). */
export function ShareButton({ title, text, url }: ShareButtonProps) {
  const [status, setStatus] = useState<Status>("idle");
  const fallbackId = useId();

  async function handleShare() {
    const outcome = await shareOrCopy({ title, text, url });
    if (outcome.status === "cancelled") {
      setStatus("idle");
      return;
    }
    setStatus(
      outcome.status === "shared"
        ? "shared"
        : outcome.status === "copied"
          ? "copied"
          : "unavailable",
    );
  }

  const showFallbackInput = status === "unavailable";

  return (
    <div className="share">
      <button
        type="button"
        onClick={handleShare}
        aria-describedby={status !== "idle" ? fallbackId : undefined}
      >
        Compartilhar
      </button>
      <p id={fallbackId} role="status" aria-live="polite" className="share__status">
        {status !== "idle" ? STATUS_MESSAGES[status] : ""}
      </p>
      {showFallbackInput ? (
        <label className="share__fallback">
          <span>Link para copiar manualmente</span>
          <input
            type="text"
            readOnly
            value={url}
            onFocus={(event) => event.currentTarget.select()}
          />
        </label>
      ) : null}
    </div>
  );
}
