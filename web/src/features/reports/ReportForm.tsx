import { type FormEvent, useState } from "react";
import { createReport } from "../../api/client";

type Props = {
  contextType: "match" | "club";
  contextSlug: string;
};

/** Contextual anonymous error report form (REQ-014 / TASK-032). */
export function ReportForm({ contextType, contextSlug }: Props) {
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await createReport({ contextType, contextSlug, message: message.trim() });
      setDone(true);
      setMessage("");
    } catch {
      setError("Não foi possível enviar o relato. Tente novamente em instantes.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="report-form" aria-labelledby="report-heading">
      <h2 id="report-heading">Relatar um erro</h2>
      <p className="note">
        Relatos são anônimos e não alteram os dados automaticamente. Um mantenedor revisa a fila.
      </p>
      {done ? <p role="status">Obrigado — seu relato entrou na fila.</p> : null}
      {error ? (
        <p className="settings-error" role="alert">
          {error}
        </p>
      ) : null}
      <form onSubmit={(e) => void onSubmit(e)}>
        <label htmlFor="report-message">Descreva o problema</label>
        <textarea
          id="report-message"
          name="message"
          required
          maxLength={1000}
          rows={3}
          value={message}
          onChange={(e) => setMessage(e.target.value)}
        />
        <button type="submit" disabled={busy || message.trim().length === 0}>
          {busy ? "Enviando…" : "Enviar relato"}
        </button>
      </form>
    </section>
  );
}
