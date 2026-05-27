import Link from "next/link";
import type {
  FieldMeta,
  ModuleMeta,
  ModuleRecord,
  SubpanelMeta,
} from "@/lib/api";
import { formatCell } from "./format";

/** A subpanel paired with the related records the page fetched for it. */
export interface SubpanelData {
  meta: SubpanelMeta;
  records: ModuleRecord[];
}

/**
 * Renders a single record entirely from metadata: detailView panels (label +
 * formatted field values) and related-record subpanels. Knows nothing about
 * any specific module.
 */
export default function RecordView({
  meta,
  record,
  subpanels,
}: {
  meta: ModuleMeta;
  record: ModuleRecord;
  subpanels: SubpanelData[];
}) {
  const fieldsByName = new Map(meta.fields.map((f) => [f.name, f]));
  const panels = meta.detailView?.panels ?? [];

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight text-gray-900">
          {String(record.name ?? meta.labelSingular)}
        </h1>
        <Link
          href={`/m/${meta.module}/${record.id}/edit`}
          className="rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-gray-800"
        >
          Edit
        </Link>
      </div>

      {panels.map((panel) => (
        <section
          key={panel.label}
          className="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm"
        >
          <h2 className="border-b border-gray-100 bg-gray-50 px-4 py-2 text-xs font-medium uppercase tracking-wide text-gray-500">
            {panel.label}
          </h2>
          <dl className="divide-y divide-gray-100">
            {panel.fields.map((name) => {
              const field = fieldsByName.get(name);
              if (!field) return null;
              return (
                <div key={name} className="flex px-4 py-3 text-sm">
                  <dt className="w-48 shrink-0 text-gray-500">{field.label}</dt>
                  <dd className="text-gray-900">
                    {formatCell(record[name], field)}
                  </dd>
                </div>
              );
            })}
          </dl>
        </section>
      ))}

      {subpanels.map(({ meta: sp, records }) => (
        <section
          key={sp.label}
          className="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm"
        >
          <h2 className="border-b border-gray-100 bg-gray-50 px-4 py-2 text-xs font-medium uppercase tracking-wide text-gray-500">
            {sp.label}
          </h2>
          {records.length === 0 ? (
            <p className="px-4 py-6 text-center text-sm text-gray-500">
              No related {sp.label.toLowerCase()}.
            </p>
          ) : (
            <SubpanelTable columns={sp.columns} records={records} />
          )}
        </section>
      ))}
    </div>
  );
}

function SubpanelTable({
  columns,
  records,
}: {
  columns: FieldMeta[];
  records: ModuleRecord[];
}) {
  return (
    <table className="min-w-full divide-y divide-gray-200 text-sm">
      <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
        <tr>
          {columns.map((f) => (
            <th key={f.name} className="px-4 py-2">
              {f.label}
            </th>
          ))}
        </tr>
      </thead>
      <tbody className="divide-y divide-gray-100">
        {records.map((r, i) => (
          <tr key={(r.id as number | string) ?? i}>
            {columns.map((f) => (
              <td key={f.name} className="px-4 py-2 text-gray-700">
                {formatCell(r[f.name], f)}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
