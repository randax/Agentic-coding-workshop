"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import type { FieldMeta, ModuleMeta, ModuleRecord } from "@/lib/api";
import { formatCell } from "./format";

type Sort = { col: string; dir: "asc" | "desc" };

/**
 * Renders a module's list as a filterable, sortable table entirely from its
 * metadata: one column per `listView.columns` entry (header = field label),
 * cells formatted by field type, and each row linking to its record view.
 * Knows nothing about any specific module.
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

  const [query, setQuery] = useState("");
  const [sort, setSort] = useState<Sort | null>(null);

  const rows = useMemo(() => {
    let out = records;
    const q = query.trim().toLowerCase();
    if (q) {
      out = out.filter((r) =>
        columns.some((f) => formatCell(r[f.name], f).toLowerCase().includes(q)),
      );
    }
    if (sort) {
      const f = fieldsByName.get(sort.col);
      out = [...out].sort((a, b) => {
        const av = a[sort.col];
        const bv = b[sort.col];
        let cmp: number;
        if (typeof av === "number" && typeof bv === "number") {
          cmp = av - bv;
        } else {
          cmp = formatCell(av, f!).localeCompare(formatCell(bv, f!));
        }
        return sort.dir === "asc" ? cmp : -cmp;
      });
    }
    return out;
    // fieldsByName/columns are derived from meta and stable per render.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [records, query, sort]);

  function toggleSort(col: string) {
    setSort((s) =>
      s && s.col === col
        ? { col, dir: s.dir === "asc" ? "desc" : "asc" }
        : { col, dir: "asc" },
    );
  }

  return (
    <div>
      <div className="border-b border-gray-100 bg-white p-3">
        <input
          aria-label="Filter"
          placeholder="Filter…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="w-64 rounded-md border border-gray-300 px-3 py-1.5 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
      </div>
      <table className="min-w-full divide-y divide-gray-200 text-sm">
        <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
          <tr>
            {columns.map((f) => (
              <th
                key={f.name}
                aria-sort={
                  sort?.col === f.name
                    ? sort.dir === "asc"
                      ? "ascending"
                      : "descending"
                    : "none"
                }
                className="px-4 py-3"
              >
                <button
                  type="button"
                  onClick={() => toggleSort(f.name)}
                  className="uppercase tracking-wide hover:text-gray-900"
                >
                  {f.label}
                </button>
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {rows.map((record, i) => (
            <tr key={(record.id as number | string) ?? i} className="hover:bg-gray-50">
              {columns.map((f, ci) => (
                <td key={f.name} className="px-4 py-3 text-gray-700">
                  {ci === 0 ? (
                    <Link
                      href={`/m/${meta.module}/${record.id}`}
                      className="font-medium text-gray-900 hover:text-blue-700"
                    >
                      {formatCell(record[f.name], f)}
                    </Link>
                  ) : (
                    formatCell(record[f.name], f)
                  )}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
