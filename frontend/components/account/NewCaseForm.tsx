"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  createCase,
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

export default function NewCaseForm({ customerId }: { customerId: number }) {
  const router = useRouter();
  const [subject, setSubject] = useState("");
  const [description, setDescription] = useState("");
  const [category, setCategory] = useState<CaseCategory>("general");
  const [priority, setPriority] = useState<CasePriority>("medium");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!subject.trim()) return;
    setBusy(true);
    setError(null);
    try {
      await createCase(customerId, { subject, description, category, priority });
      setSubject("");
      setDescription("");
      setCategory("general");
      setPriority("medium");
      router.refresh();
    } catch {
      setError("Could not open the case. Is the backend running?");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="space-y-4 rounded-lg border border-gray-200 bg-gray-50 p-4"
    >
      <h3 className="text-sm font-semibold text-gray-900">Open a new case</h3>

      {error && (
        <div
          role="alert"
          className="rounded-md border border-red-200 bg-red-50 p-2 text-sm text-red-800"
        >
          {error}
        </div>
      )}

      <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
        Subject
        <input
          aria-label="Subject"
          value={subject}
          required
          onChange={(e) => setSubject(e.target.value)}
          className="rounded-md border border-gray-300 px-3 py-2 text-sm normal-case text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
      </label>

      <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
        Description
        <textarea
          aria-label="Description"
          value={description}
          rows={3}
          onChange={(e) => setDescription(e.target.value)}
          className="rounded-md border border-gray-300 px-3 py-2 text-sm normal-case text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
      </label>

      <div className="flex flex-wrap gap-4">
        <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
          Category
          <select
            aria-label="Category"
            value={category}
            onChange={(e) => setCategory(e.target.value as CaseCategory)}
            className="rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
            {CATEGORIES.map((c) => (
              <option key={c.value} value={c.value}>
                {c.label}
              </option>
            ))}
          </select>
        </label>

        <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
          Priority
          <select
            aria-label="Priority"
            value={priority}
            onChange={(e) => setPriority(e.target.value as CasePriority)}
            className="rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
            {PRIORITIES.map((p) => (
              <option key={p.value} value={p.value}>
                {p.label}
              </option>
            ))}
          </select>
        </label>
      </div>

      <button
        type="submit"
        disabled={busy}
        className="rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-gray-800 disabled:opacity-50"
      >
        Open case
      </button>
    </form>
  );
}
