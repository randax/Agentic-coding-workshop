import type { FieldMeta } from "@/lib/api";

/** Formats a single record value according to its field type, for display. */
export function formatCell(value: unknown, field: FieldMeta): string {
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
