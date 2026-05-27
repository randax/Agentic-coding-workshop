import type { CaseStatus } from "@/lib/api";

/** Human-readable labels for the case lifecycle states. */
export const STATUS_LABELS: Record<CaseStatus, string> = {
  open: "Open",
  in_progress: "In progress",
  resolved: "Resolved",
  closed: "Closed",
};

/** Badge styles (Tailwind) per case status. */
export const STATUS_STYLES: Record<CaseStatus, string> = {
  open: "bg-blue-100 text-blue-800 ring-blue-600/20",
  in_progress: "bg-amber-100 text-amber-800 ring-amber-600/20",
  resolved: "bg-green-100 text-green-800 ring-green-600/20",
  closed: "bg-gray-100 text-gray-600 ring-gray-500/20",
};
