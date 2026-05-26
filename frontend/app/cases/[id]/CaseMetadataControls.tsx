"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  updateCaseMetadata,
  type Agent,
  type CaseCategory,
  type CasePriority,
} from "@/lib/api";

const CATEGORIES: { value: CaseCategory; label: string }[] = [
  { value: "billing", label: "Billing" },
  { value: "connectivity", label: "Connectivity" },
  { value: "hardware", label: "Hardware" },
  { value: "tv", label: "TV" },
  { value: "general", label: "General" },
];

const PRIORITIES: { value: CasePriority; label: string }[] = [
  { value: "low", label: "Low" },
  { value: "medium", label: "Medium" },
  { value: "high", label: "High" },
  { value: "urgent", label: "Urgent" },
];

export default function CaseMetadataControls({
  caseId,
  priority,
  category,
  assignedAgentId,
  agents,
}: {
  caseId: number;
  priority: CasePriority;
  category: CaseCategory;
  assignedAgentId?: number | null;
  agents: Agent[];
}) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function apply(patch: {
    priority?: CasePriority;
    category?: CaseCategory;
    assignedAgentId?: number;
  }) {
    setBusy(true);
    setError(null);
    try {
      await updateCaseMetadata(caseId, patch);
      router.refresh();
    } catch {
      setError("Could not update the case. Is the backend running?");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-3 rounded-lg border border-gray-200 bg-gray-50 p-4">
      <h3 className="text-sm font-semibold text-gray-900">Details</h3>

      {error && (
        <div
          role="alert"
          className="rounded-md border border-red-200 bg-red-50 p-2 text-sm text-red-800"
        >
          {error}
        </div>
      )}

      <div className="flex flex-wrap gap-4">
        <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
          Priority
          <select
            aria-label="Priority"
            defaultValue={priority}
            disabled={busy}
            onChange={(e) => apply({ priority: e.target.value as CasePriority })}
            className="rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
            {PRIORITIES.map((p) => (
              <option key={p.value} value={p.value}>
                {p.label}
              </option>
            ))}
          </select>
        </label>

        <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
          Category
          <select
            aria-label="Category"
            defaultValue={category}
            disabled={busy}
            onChange={(e) => apply({ category: e.target.value as CaseCategory })}
            className="rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
            {CATEGORIES.map((cat) => (
              <option key={cat.value} value={cat.value}>
                {cat.label}
              </option>
            ))}
          </select>
        </label>

        <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
          Assigned to
          <select
            aria-label="Assigned to"
            defaultValue={assignedAgentId ?? ""}
            disabled={busy}
            onChange={(e) =>
              apply({ assignedAgentId: Number(e.target.value) })
            }
            className="rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
            {assignedAgentId == null && (
              <option value="" disabled>
                Unassigned
              </option>
            )}
            {agents.map((a) => (
              <option key={a.id} value={a.id}>
                {a.name}
              </option>
            ))}
          </select>
        </label>
      </div>
    </div>
  );
}
