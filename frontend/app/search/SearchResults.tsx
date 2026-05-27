import Link from "next/link";
import type { SearchGroup, SearchHit, SearchModule } from "@/lib/api";

/** Human label per module, shown as the group heading. */
const MODULE_LABEL: Record<SearchModule, string> = {
  accounts: "Accounts",
  contacts: "Contacts",
  leads: "Leads",
  opportunities: "Opportunities",
  cases: "Cases",
};

/** The route a hit links to. Most modules use the generic record view; cases
 * have their own dedicated page. */
function hitHref(hit: SearchHit): string {
  if (hit.module === "cases") return `/cases/${hit.id}`;
  return `/m/${hit.module}/${hit.id}`;
}

/**
 * Presentational global-search results: hits grouped and labeled by module,
 * each linking to its record. The data (and its visibility scoping) comes from
 * the server; this component knows nothing about how it was fetched.
 */
export default function SearchResults({
  query,
  groups,
}: {
  query: string;
  groups: SearchGroup[];
}) {
  if (groups.length === 0) {
    return (
      <p className="text-sm text-gray-500">
        No results for “<span className="font-medium text-gray-700">{query}</span>”.
      </p>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      {groups.map((group) => (
        <section
          key={group.module}
          aria-label={MODULE_LABEL[group.module]}
          className="rounded-lg border border-gray-200 bg-white"
        >
          <h2 className="border-b border-gray-100 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-gray-500">
            {MODULE_LABEL[group.module]}
          </h2>
          <ul className="divide-y divide-gray-100">
            {group.hits.map((hit) => (
              <li key={hit.id} className="px-4 py-2 text-sm">
                <Link
                  href={hitHref(hit)}
                  className="font-medium text-gray-900 hover:text-blue-700"
                >
                  {hit.title}
                </Link>
              </li>
            ))}
          </ul>
        </section>
      ))}
    </div>
  );
}
