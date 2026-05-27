import type { FieldMeta, ModuleMeta, ModuleRecord } from "@/lib/api";

/** Formats a single cell value according to its field type. */
function formatCell(value: unknown, field: FieldMeta): string {
  if (value == null) return "—";
  switch (field.type) {
    case "currency": {
      const n = Number(value);
      return `kr ${Number.isInteger(n) ? n : n.toFixed(2)}`;
    }
    case "bool":
      return value ? "Yes" : "No";
    case "enum": {
      const s = String(value);
      return s.charAt(0).toUpperCase() + s.slice(1);
    }
    default:
      return String(value);
  }
}

/**
 * Renders a module's list as a table entirely from its metadata: one column per
 * `listView.columns` entry (header = the field's label), each cell formatted by
 * the field's type. Knows nothing about any specific module.
 */
export default function ListView({
  meta,
  records,
}: {
  meta: ModuleMeta;
  records: ModuleRecord[];
}) {
  const fieldsByName = new Map(meta.fields.map((f) => [f.name, f]));
  const columns = meta.listView.columns
    .map((name) => fieldsByName.get(name))
    .filter((f): f is FieldMeta => f !== undefined);

  return (
    <table className="min-w-full divide-y divide-gray-200 text-sm">
      <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
        <tr>
          {columns.map((f) => (
            <th key={f.name} className="px-4 py-3">
              {f.label}
            </th>
          ))}
        </tr>
      </thead>
      <tbody className="divide-y divide-gray-100">
        {records.map((record, i) => (
          <tr key={(record.id as number | string) ?? i} className="hover:bg-gray-50">
            {columns.map((f) => (
              <td key={f.name} className="px-4 py-3 text-gray-700">
                {formatCell(record[f.name], f)}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
