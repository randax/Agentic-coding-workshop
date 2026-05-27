"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

/**
 * The global search box in the app nav. Submitting navigates to the /search
 * results page with the query in the URL, so that page (a server component) is
 * the single source of truth for what was searched.
 */
export default function GlobalSearch() {
  const router = useRouter();
  const [value, setValue] = useState("");

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const q = value.trim();
    if (q) router.push(`/search?q=${encodeURIComponent(q)}`);
  }

  return (
    <form onSubmit={onSubmit} role="search">
      <input
        type="search"
        aria-label="Search all records"
        placeholder="Search…"
        value={value}
        onChange={(e) => setValue(e.target.value)}
        className="w-44 rounded-md border border-gray-300 px-3 py-1.5 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
      />
    </form>
  );
}
