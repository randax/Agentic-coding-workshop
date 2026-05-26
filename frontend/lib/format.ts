/** Formats an ISO-8601 date string as a short, human-readable date (e.g. "9 Dec 2020"). */
export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-GB", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}
