"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { updateCaseStatus, type CaseStatus } from "@/lib/api";
import { STATUS_LABELS } from "@/app/cases/status";

// Mirror of the backend lifecycle graph, used only to offer the agent the legal
// next states. The backend remains the authority and rejects illegal moves.
const NEXT_STATES: Record<CaseStatus, CaseStatus[]> = {
  open: ["in_progress"],
  in_progress: ["resolved"],
  resolved: ["closed", "in_progress"],
  closed: [],
};

export default function CaseStatusControl({
  caseId,
  status,
}: {
  caseId: number;
  status: CaseStatus;
}) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const nextStates = NEXT_STATES[status];

  async function advance(to: CaseStatus) {
    setBusy(true);
    setError(null);
    try {
      await updateCaseStatus(caseId, to);
      router.refresh();
    } catch {
      setError("Could not change the status. Is the backend running?");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-3 rounded-lg border border-gray-200 bg-gray-50 p-4">
      <h3 className="text-sm font-semibold text-gray-900">Status</h3>

      {error && (
        <div
          role="alert"
          className="rounded-md border border-red-200 bg-red-50 p-2 text-sm text-red-800"
        >
          {error}
        </div>
      )}

      {nextStates.length === 0 ? (
        <p className="text-sm text-gray-500">
          This case is closed — no further status changes.
        </p>
      ) : (
        <div className="flex flex-wrap gap-2">
          {nextStates.map((to) => (
            <button
              key={to}
              type="button"
              disabled={busy}
              onClick={() => advance(to)}
              className="rounded-md bg-gray-900 px-3 py-1.5 text-sm font-medium text-white shadow-sm hover:bg-gray-800 disabled:opacity-50"
            >
              Mark as {STATUS_LABELS[to]}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
