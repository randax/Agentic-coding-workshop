import { cookies } from "next/headers";
import { globalSearch, type SearchGroup, API_BASE_URL } from "@/lib/api";
import SearchResults from "./SearchResults";

/**
 * Global search results page. The query lives in the URL (`?q=`); this server
 * component reads it, forwards the session cookie so results are scoped to what
 * the user may see, and renders them grouped by module.
 */
export default async function SearchPage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string }>;
}) {
  const { q } = await searchParams;
  const query = (q ?? "").trim();

  let groups: SearchGroup[] = [];
  let error: string | null = null;
  if (query) {
    const cookie = (await cookies()).toString();
    try {
      groups = await globalSearch(query, cookie);
    } catch {
      error = `Could not search via the API at ${API_BASE_URL}. Is the backend running?`;
    }
  }

  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10">
      <header className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight text-gray-900">
          Search
        </h1>
        {query && (
          <p className="mt-1 text-sm text-gray-500">
            Results for “{query}”.
          </p>
        )}
      </header>

      {error ? (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
          {error}
        </div>
      ) : !query ? (
        <p className="text-sm text-gray-500">
          Type a query in the search box to find records across modules.
        </p>
      ) : (
        <SearchResults query={query} groups={groups} />
      )}
    </main>
  );
}
