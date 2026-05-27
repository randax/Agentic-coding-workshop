"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { runRecordAction, type ActionMeta } from "@/lib/api";

/** Renders a button per metadata-declared record action (e.g. Convert a lead),
 * running it for this record and refreshing on success. */
export default function RecordActions({
  actions,
  recordId,
}: {
  actions: ActionMeta[];
  recordId: number;
}) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (actions.length === 0) return null;

  async function run(action: ActionMeta) {
    setBusy(true);
    setError(null);
    try {
      await runRecordAction(action, recordId);
      router.refresh();
    } catch {
      setError(`Could not ${action.label.toLowerCase()}.`);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex items-center gap-3">
      {error && <span className="text-sm text-red-700">{error}</span>}
      {actions.map((a) => (
        <button
          key={a.label}
          type="button"
          onClick={() => run(a)}
          disabled={busy}
          className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-800 shadow-sm hover:bg-gray-50 disabled:opacity-50"
        >
          {a.label}
        </button>
      ))}
    </div>
  );
}
