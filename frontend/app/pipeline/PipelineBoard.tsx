import Link from "next/link";
import type { PipelineStage } from "@/lib/api";
import { formatMonthlyPrice } from "@/lib/format";

function stageLabel(stage: string): string {
  const s = stage.replace(/_/g, " ");
  return s.charAt(0).toUpperCase() + s.slice(1);
}

function amount(value: unknown): string {
  // Reuse the NOK formatter but drop the "/ mo" suffix — deals aren't monthly.
  return formatMonthlyPrice(Number(value)).replace(" / mo", "");
}

/**
 * Renders the sales pipeline as a column per stage, each showing its rolled-up
 * total and the opportunities in that stage (linked to their record).
 */
export default function PipelineBoard({ stages }: { stages: PipelineStage[] }) {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {stages.map((s) => (
        <section
          key={s.stage}
          data-testid={`stage-${s.stage}`}
          className="rounded-lg border border-gray-200 bg-white shadow-sm"
        >
          <header className="flex items-baseline justify-between border-b border-gray-100 px-4 py-2">
            <h2 className="text-sm font-semibold text-gray-900">
              {stageLabel(s.stage)}
            </h2>
            <span className="text-xs text-gray-500">
              {s.count} · {amount(s.totalAmount)}
            </span>
          </header>
          <ul className="divide-y divide-gray-100">
            {s.items.length === 0 ? (
              <li className="px-4 py-6 text-center text-xs text-gray-400">
                Empty
              </li>
            ) : (
              s.items.map((o) => (
                <li key={o.id as number} className="px-4 py-2 text-sm">
                  <Link
                    href={`/m/opportunities/${o.id}`}
                    className="font-medium text-gray-900 hover:text-blue-700"
                  >
                    {String(o.name)}
                  </Link>
                  <span className="ml-2 text-gray-500">{amount(o.amount)}</span>
                </li>
              ))
            )}
          </ul>
        </section>
      ))}
    </div>
  );
}
