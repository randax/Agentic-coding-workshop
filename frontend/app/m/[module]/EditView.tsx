"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  createModuleRecord,
  updateModuleRecord,
  type FieldMeta,
  type ModuleMeta,
  type ModuleRecord,
} from "@/lib/api";

/**
 * Generic create/edit form. Renders one input per `editView` field, typed by
 * the field's metadata (string→text, enum→select, currency→number,
 * bool→checkbox). Given a `record` it edits (pre-filled, PUT); without one it
 * creates a new record (blank, POST). Either way it navigates to the record
 * afterwards. Knows nothing about any specific module.
 */
export default function EditView({
  meta,
  record,
}: {
  meta: ModuleMeta;
  record?: ModuleRecord;
}) {
  const router = useRouter();
  const isCreate = record === undefined;
  const fieldsByName = new Map(meta.fields.map((f) => [f.name, f]));
  const editFields = (meta.editView?.fields ?? [])
    .map((name) => fieldsByName.get(name))
    .filter((f): f is FieldMeta => f !== undefined);

  const [values, setValues] = useState<ModuleRecord>(() => {
    const initial: ModuleRecord = {};
    for (const f of editFields) initial[f.name] = record?.[f.name] ?? "";
    return initial;
  });
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function set(name: string, value: unknown) {
    setValues((v) => ({ ...v, [name]: value }));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      const saved = isCreate
        ? await createModuleRecord(meta.module, values)
        : await updateModuleRecord(meta.module, record.id as number, values);
      router.push(`/m/${meta.module}/${saved.id}`);
      router.refresh();
    } catch {
      setError("Could not save. Check the fields and that the backend is running.");
      setBusy(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="max-w-lg space-y-5">
      {error && (
        <div
          role="alert"
          className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800"
        >
          {error}
        </div>
      )}

      {editFields.map((f) => (
        <label
          key={f.name}
          className="flex flex-col gap-1 text-sm font-medium text-gray-700"
        >
          {f.label}
          <FieldInput field={f} value={values[f.name]} onChange={(v) => set(f.name, v)} />
        </label>
      ))}

      <button
        type="submit"
        disabled={busy}
        className="rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-gray-800 disabled:opacity-50"
      >
        {isCreate ? "Create" : "Save changes"}
      </button>
    </form>
  );
}

function FieldInput({
  field,
  value,
  onChange,
}: {
  field: FieldMeta;
  value: unknown;
  onChange: (value: unknown) => void;
}) {
  const cls =
    "rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500";

  if (field.type === "enum") {
    return (
      <select
        aria-label={field.label}
        value={String(value ?? "")}
        onChange={(e) => onChange(e.target.value)}
        className={cls}
      >
        {(field.options ?? []).map((opt) => (
          <option key={opt} value={opt}>
            {opt.charAt(0).toUpperCase() + opt.slice(1)}
          </option>
        ))}
      </select>
    );
  }

  if (field.type === "bool") {
    return (
      <input
        type="checkbox"
        aria-label={field.label}
        checked={Boolean(value)}
        onChange={(e) => onChange(e.target.checked)}
        className="h-4 w-4 self-start"
      />
    );
  }

  return (
    <input
      type={field.type === "currency" ? "number" : "text"}
      aria-label={field.label}
      value={value == null ? "" : String(value)}
      onChange={(e) =>
        onChange(field.type === "currency" ? Number(e.target.value) : e.target.value)
      }
      className={cls}
    />
  );
}
