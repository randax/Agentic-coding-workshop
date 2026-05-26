"use client";

import { useRouter, usePathname, useSearchParams } from "next/navigation";

/**
 * Search and status controls for the customer list. Each change is written to
 * the URL query string; the server page reads those params and refetches, so
 * the URL is the single source of truth for the active filter.
 */
export default function CustomerSearch() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  function setParam(key: string, value: string) {
    const params = new URLSearchParams(searchParams.toString());
    if (value) {
      params.set(key, value);
    } else {
      params.delete(key);
    }
    const qs = params.toString();
    router.replace(qs ? `${pathname}?${qs}` : pathname);
  }

  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
      <input
        type="search"
        aria-label="Search customers"
        placeholder="Search by name, email, or account number"
        defaultValue={searchParams.get("search") ?? ""}
        onChange={(e) => setParam("search", e.target.value)}
        className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 sm:max-w-xs"
      />
      <select
        aria-label="Status"
        defaultValue={searchParams.get("status") ?? ""}
        onChange={(e) => setParam("status", e.target.value)}
        className="rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
      >
        <option value="">All statuses</option>
        <option value="active">Active</option>
        <option value="suspended">Suspended</option>
      </select>
    </div>
  );
}
