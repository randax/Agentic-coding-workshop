import Link from "next/link";
import type { DashboardData } from "@/lib/api";
import { formatDate } from "@/lib/format";

/** Humanizes an enum-ish value: "in_progress" → "in progress". */
const humanize = (s: string) => s.replace(/_/g, " ");

/**
 * The dashboard landing page's body: four dashlets rendered from already-scoped
 * data. Presentational only — the data (and its visibility scoping) comes from
 * the server. Knows nothing about how the data was fetched.
 */
export default function Dashboard({ data }: { data: DashboardData }) {
  return (
    <div className="grid gap-6 md:grid-cols-2">
      <Dashlet title="My Open Cases" empty="No open cases." isEmpty={data.myOpenCases.length === 0}>
        <ul className="divide-y divide-gray-100">
          {data.myOpenCases.map((c) => (
            <li key={c.id} className="flex items-center justify-between gap-3 py-2 text-sm">
              <Link href={`/cases/${c.id}`} className="font-medium text-gray-900 hover:text-blue-700">
                {c.subject}
              </Link>
              <span className="shrink-0 text-xs uppercase tracking-wide text-gray-500">
                {humanize(c.status)}
              </span>
            </li>
          ))}
        </ul>
      </Dashlet>

      <Dashlet title="My Tasks" empty="No open tasks." isEmpty={data.myTasks.length === 0}>
        <ul className="divide-y divide-gray-100">
          {data.myTasks.map((t) => (
            <li key={t.id} className="flex items-center justify-between gap-3 py-2 text-sm">
              <span className="text-gray-900">{t.subject}</span>
              <span className="shrink-0 text-xs text-gray-500">{formatDate(t.occurredAt)}</span>
            </li>
          ))}
        </ul>
      </Dashlet>

      <Dashlet title="Recent Leads" empty="No recent leads." isEmpty={data.recentLeads.length === 0}>
        <ul className="divide-y divide-gray-100">
          {data.recentLeads.map((l) => (
            <li key={l.id} className="flex items-center justify-between gap-3 py-2 text-sm">
              <Link href={`/m/leads/${l.id}`} className="font-medium text-gray-900 hover:text-blue-700">
                {l.name}
              </Link>
              <span className="shrink-0 text-xs text-gray-500">{l.company}</span>
            </li>
          ))}
        </ul>
      </Dashlet>

      <Dashlet
        title="Pipeline by Stage"
        empty="No opportunities."
        isEmpty={data.pipelineByStage.every((s) => s.count === 0)}
      >
        <ul className="divide-y divide-gray-100">
          {data.pipelineByStage.map((s) => (
            <li key={s.stage} className="flex items-center justify-between gap-3 py-2 text-sm">
              <span className="capitalize text-gray-900">{humanize(s.stage)}</span>
              <span className="flex shrink-0 items-center gap-4 text-gray-500">
                <span>{s.count}</span>
                <span className="tabular-nums">kr {Math.round(s.totalAmount).toLocaleString("en-GB")}</span>
              </span>
            </li>
          ))}
        </ul>
      </Dashlet>
    </div>
  );
}

function Dashlet({
  title,
  empty,
  isEmpty,
  children,
}: {
  title: string;
  empty: string;
  isEmpty: boolean;
  children: React.ReactNode;
}) {
  return (
    <section aria-label={title} className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
      <h2 className="mb-2 text-sm font-semibold text-gray-900">{title}</h2>
      {isEmpty ? (
        <p className="py-4 text-center text-sm text-gray-500">{empty}</p>
      ) : (
        children
      )}
    </section>
  );
}
