import Link from "next/link";
import type { Case, CaseStatus, CasePriority } from "@/lib/api";

export const STATUS_LABELS: Record<CaseStatus, string> = {
  open: "Open",
  in_progress: "In progress",
  resolved: "Resolved",
  closed: "Closed",
};

export const STATUS_STYLES: Record<CaseStatus, string> = {
  open: "bg-blue-100 text-blue-800 ring-blue-600/20",
  in_progress: "bg-amber-100 text-amber-800 ring-amber-600/20",
  resolved: "bg-green-100 text-green-800 ring-green-600/20",
  closed: "bg-gray-100 text-gray-600 ring-gray-500/20",
};

const PRIORITY_STYLES: Record<CasePriority, string> = {
  low: "text-gray-500",
  medium: "text-gray-700",
  high: "text-orange-700",
  urgent: "text-red-700 font-semibold",
};

export default function CaseList({ cases }: { cases: Case[] }) {
  if (cases.length === 0) {
    return (
      <div className="rounded-lg border border-dashed border-gray-300 p-10 text-center text-sm text-gray-500">
        No support cases to display yet.
      </div>
    );
  }

  return (
    <div className="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm">
      <table className="min-w-full divide-y divide-gray-200 text-sm">
        <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
          <tr>
            <th className="px-4 py-3">Subject</th>
            <th className="px-4 py-3">Category</th>
            <th className="px-4 py-3">Priority</th>
            <th className="px-4 py-3">Status</th>
            <th className="px-4 py-3">Assigned to</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {cases.map((c) => (
            <tr key={c.id} className="hover:bg-gray-50">
              <td className="px-4 py-3 font-medium">
                <Link
                  href={`/cases/${c.id}`}
                  className="text-gray-900 hover:text-blue-700 hover:underline"
                >
                  {c.subject}
                </Link>
              </td>
              <td className="px-4 py-3 capitalize text-gray-600">
                {c.category}
              </td>
              <td className={`px-4 py-3 capitalize ${PRIORITY_STYLES[c.priority]}`}>
                {c.priority}
              </td>
              <td className="px-4 py-3">
                <span
                  className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${STATUS_STYLES[c.status]}`}
                >
                  {STATUS_LABELS[c.status]}
                </span>
              </td>
              <td className="px-4 py-3 text-gray-600">
                {c.assignedAgent ? (
                  c.assignedAgent.name
                ) : (
                  <span className="text-gray-400">Unassigned</span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
