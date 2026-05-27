"use client";

import { useState } from "react";
import {
  runReport,
  saveReport,
  type FieldMeta,
  type ReportAggregation,
  type ReportDefinition,
  type ReportOperator,
  type ReportResult,
} from "@/lib/api";
import ReportChart from "./ReportChart";

/** A module the builder can report over, with the fields it can group/filter on. */
export interface ReportModule {
  name: string;
  label: string;
  fields: FieldMeta[];
}

const AGGREGATIONS: ReportAggregation[] = ["count", "sum", "avg"];
const OPERATORS: ReportOperator[] = ["eq", "contains", "gt", "lt"];

/** gt/lt compare numbers, so their value is coerced from the text input. */
const isNumericOp = (op: ReportOperator) => op === "gt" || op === "lt";

/** A filter row in the builder (its value stays a string until the report runs). */
interface FilterRow {
  field: string;
  operator: ReportOperator;
  value: string;
}

const inputClass =
  "rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500";

/**
 * Interactive report builder: pick a module, choose grouping and aggregation,
 * add filters (which may target custom fields), then run the report — rendering
 * the result as a chart — and optionally save it to re-run later.
 */
export default function ReportBuilder({ modules }: { modules: ReportModule[] }) {
  const [moduleName, setModuleName] = useState(modules[0]?.name ?? "");
  const fields = (modules.find((m) => m.name === moduleName) ?? modules[0])?.fields ?? [];

  const [groupBy, setGroupBy] = useState(fields[0]?.name ?? "");
  const [aggregation, setAggregation] = useState<ReportAggregation>("count");
  const [aggField, setAggField] = useState(fields[0]?.name ?? "");
  const [filters, setFilters] = useState<FilterRow[]>([]);
  const [name, setName] = useState("");

  const [result, setResult] = useState<ReportResult | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState<string | null>(null);

  // Switching modules resets the field-dependent selections.
  function changeModule(next: string) {
    setModuleName(next);
    const first = modules.find((m) => m.name === next)?.fields[0]?.name ?? "";
    setGroupBy(first);
    setAggField(first);
    setFilters([]);
    setResult(null);
  }

  function updateFilter(i: number, patch: Partial<FilterRow>) {
    setFilters((rows) => rows.map((r, j) => (j === i ? { ...r, ...patch } : r)));
  }

  function definition(): ReportDefinition {
    const def: ReportDefinition = { module: moduleName, groupBy, aggregation };
    if (aggregation !== "count") {
      def.aggField = aggField;
    }
    if (filters.length > 0) {
      def.filters = filters.map((f) => ({
        field: f.field,
        operator: f.operator,
        value: isNumericOp(f.operator) ? Number(f.value) : f.value,
      }));
    }
    return def;
  }

  async function handleRun() {
    setBusy(true);
    setError(null);
    try {
      setResult(await runReport(definition()));
    } catch {
      setError("Could not run the report.");
    } finally {
      setBusy(false);
    }
  }

  async function handleSave() {
    setBusy(true);
    setError(null);
    setSaved(null);
    try {
      await saveReport(name, definition());
      setSaved(`Saved “${name}”.`);
    } catch {
      setError("Could not save the report. A name is required.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex flex-col gap-6">
      {error && (
        <div role="alert" className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800">
          {error}
        </div>
      )}
      {saved && (
        <div className="rounded-lg border border-green-200 bg-green-50 p-3 text-sm text-green-800">
          {saved}
        </div>
      )}

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <label className="flex flex-col gap-1 text-sm font-medium text-gray-700">
          Module
          <select aria-label="Module" value={moduleName} onChange={(e) => changeModule(e.target.value)} className={inputClass}>
            {modules.map((m) => (
              <option key={m.name} value={m.name}>{m.label}</option>
            ))}
          </select>
        </label>

        <label className="flex flex-col gap-1 text-sm font-medium text-gray-700">
          Group by
          <select aria-label="Group by" value={groupBy} onChange={(e) => setGroupBy(e.target.value)} className={inputClass}>
            {fields.map((f) => (
              <option key={f.name} value={f.name}>{f.label}</option>
            ))}
          </select>
        </label>

        <label className="flex flex-col gap-1 text-sm font-medium text-gray-700">
          Aggregation
          <select aria-label="Aggregation" value={aggregation} onChange={(e) => setAggregation(e.target.value as ReportAggregation)} className={inputClass}>
            {AGGREGATIONS.map((a) => (
              <option key={a} value={a}>{a}</option>
            ))}
          </select>
        </label>

        {aggregation !== "count" && (
          <label className="flex flex-col gap-1 text-sm font-medium text-gray-700">
            Field to aggregate
            <select aria-label="Field to aggregate" value={aggField} onChange={(e) => setAggField(e.target.value)} className={inputClass}>
              {fields.map((f) => (
                <option key={f.name} value={f.name}>{f.label}</option>
              ))}
            </select>
          </label>
        )}
      </div>

      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium text-gray-700">Filters</span>
          <button
            type="button"
            onClick={() => setFilters((rows) => [...rows, { field: fields[0]?.name ?? "", operator: "eq", value: "" }])}
            className="text-sm font-medium text-blue-700 hover:text-blue-900"
          >
            Add filter
          </button>
        </div>
        {filters.map((f, i) => (
          <div key={i} className="flex flex-wrap items-center gap-2">
            <select aria-label="Filter field" value={f.field} onChange={(e) => updateFilter(i, { field: e.target.value })} className={inputClass}>
              {fields.map((fld) => (
                <option key={fld.name} value={fld.name}>{fld.label}</option>
              ))}
            </select>
            <select aria-label="Filter operator" value={f.operator} onChange={(e) => updateFilter(i, { operator: e.target.value as ReportOperator })} className={inputClass}>
              {OPERATORS.map((op) => (
                <option key={op} value={op}>{op}</option>
              ))}
            </select>
            <input aria-label="Filter value" value={f.value} onChange={(e) => updateFilter(i, { value: e.target.value })} className={inputClass} />
            <button
              type="button"
              onClick={() => setFilters((rows) => rows.filter((_, j) => j !== i))}
              className="text-sm text-gray-500 hover:text-red-700"
            >
              Remove
            </button>
          </div>
        ))}
      </div>

      <div className="flex flex-wrap items-end gap-3">
        <button
          type="button"
          onClick={handleRun}
          disabled={busy}
          className="rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-gray-800 disabled:opacity-50"
        >
          Run report
        </button>
        <label className="flex flex-col gap-1 text-sm font-medium text-gray-700">
          Report name
          <input aria-label="Report name" value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. Leads by status" className={inputClass} />
        </label>
        <button
          type="button"
          onClick={handleSave}
          disabled={busy}
          className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 disabled:opacity-50"
        >
          Save report
        </button>
      </div>

      {result && (
        <section aria-label="Report result" className="rounded-lg border border-gray-200 bg-white p-4">
          <ReportChart result={result} />
        </section>
      )}
    </div>
  );
}
