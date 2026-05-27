import type { ReportResult } from "@/lib/api";

/** Turns a snake/lower-case group label into a readable one ("closed_won" → "Closed won"). */
function label(group: string): string {
  const s = group.replace(/_/g, " ");
  return s.charAt(0).toUpperCase() + s.slice(1);
}

/**
 * Renders a report result as a horizontal bar chart — one labeled bar per group,
 * its length proportional to the group's aggregated value, with the value shown.
 * Purely presentational: the data (and its access scoping) comes from the caller.
 */
export default function ReportChart({ result }: { result: ReportResult }) {
  if (result.rows.length === 0) {
    return <p className="text-sm text-gray-500">No data for this report.</p>;
  }

  const max = Math.max(...result.rows.map((r) => r.value), 0);

  return (
    <div className="flex flex-col gap-2">
      {result.rows.map((row) => (
        <div
          key={row.group}
          data-testid={`report-bar-${row.group}`}
          className="flex items-center gap-3 text-sm"
        >
          <span className="w-32 shrink-0 truncate text-gray-700" title={label(row.group)}>
            {label(row.group)}
          </span>
          <div className="h-5 flex-1 rounded bg-gray-100">
            <div
              className="h-5 rounded bg-blue-500"
              style={{ width: max > 0 ? `${(row.value / max) * 100}%` : "0%" }}
            />
          </div>
          <span className="w-24 shrink-0 text-right tabular-nums text-gray-900">
            {row.value.toLocaleString()}
          </span>
        </div>
      ))}
    </div>
  );
}
