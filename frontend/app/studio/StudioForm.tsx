"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { addCustomField, type FieldType } from "@/lib/api";

const TYPES: FieldType[] = ["string", "enum", "currency", "bool", "date"];

/** Admin form to add a custom field to a module. On success it confirms and
 * refreshes (so the new field shows up in that module's views). */
export default function StudioForm({ modules }: { modules: string[] }) {
  const router = useRouter();
  const [module, setModule] = useState(modules[0] ?? "");
  const [name, setName] = useState("");
  const [label, setLabel] = useState("");
  const [type, setType] = useState<FieldType>("string");
  const [options, setOptions] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [added, setAdded] = useState<string | null>(null);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    setAdded(null);
    try {
      await addCustomField({
        module,
        name,
        label,
        type,
        options:
          type === "enum"
            ? options.split(",").map((o) => o.trim()).filter(Boolean)
            : undefined,
      });
      setAdded(`Added “${label}” to ${module}.`);
      setName("");
      setLabel("");
      setOptions("");
      router.refresh();
    } catch {
      setError("Could not add the field. Admin access is required.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="max-w-lg space-y-5">
      {error && (
        <div role="alert" className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800">
          {error}
        </div>
      )}
      {added && (
        <div className="rounded-lg border border-green-200 bg-green-50 p-3 text-sm text-green-800">
          {added}
        </div>
      )}

      <label className="flex flex-col gap-1 text-sm font-medium text-gray-700">
        Module
        <select
          aria-label="Module"
          value={module}
          onChange={(e) => setModule(e.target.value)}
          className="rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        >
          {modules.map((m) => (
            <option key={m} value={m}>
              {m}
            </option>
          ))}
        </select>
      </label>

      <label className="flex flex-col gap-1 text-sm font-medium text-gray-700">
        Field name
        <input
          aria-label="Field name"
          value={name}
          required
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. churnRisk"
          className="rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
      </label>

      <label className="flex flex-col gap-1 text-sm font-medium text-gray-700">
        Label
        <input
          aria-label="Label"
          value={label}
          required
          onChange={(e) => setLabel(e.target.value)}
          className="rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
      </label>

      <label className="flex flex-col gap-1 text-sm font-medium text-gray-700">
        Type
        <select
          aria-label="Type"
          value={type}
          onChange={(e) => setType(e.target.value as FieldType)}
          className="rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        >
          {TYPES.map((t) => (
            <option key={t} value={t}>
              {t}
            </option>
          ))}
        </select>
      </label>

      {type === "enum" && (
        <label className="flex flex-col gap-1 text-sm font-medium text-gray-700">
          Options (comma-separated)
          <input
            aria-label="Options"
            value={options}
            onChange={(e) => setOptions(e.target.value)}
            placeholder="low, medium, high"
            className="rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          />
        </label>
      )}

      <button
        type="submit"
        disabled={busy}
        className="rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-gray-800 disabled:opacity-50"
      >
        Add field
      </button>
    </form>
  );
}
