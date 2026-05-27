import type { FieldMeta } from "@/lib/api";
import { formatDate } from "@/lib/format";

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
      // Render enum values readably: "closed_won" → "Closed won".
      const s = String(value).replace(/_/g, " ");
      return s.charAt(0).toUpperCase() + s.slice(1);
    }
    case "date": {
      const s = String(value);
      // Go's zero time serializes as year 0001; show it as blank.
      if (!s || s.startsWith("0001-01-01")) return "—";
      return formatDate(s);
    }
    default:
      return String(value);
  }
}
