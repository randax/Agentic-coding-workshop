import Link from "next/link";
import type { Customer } from "@/lib/api";
import { formatDate } from "@/lib/format";
import StatusBadge from "@/components/StatusBadge";

/** Presentational table of customers. Renders an empty state when the list is empty. */
export default function CustomerTable({ customers }: { customers: Customer[] }) {
  if (customers.length === 0) {
    return (
      <div className="rounded-lg border border-dashed border-gray-300 p-10 text-center text-sm text-gray-500">
        No customers match your search.
      </div>
    );
  }

  return (
    <div className="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm">
      <table className="min-w-full divide-y divide-gray-200 text-sm">
        <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
          <tr>
            <th className="px-4 py-3">Name</th>
            <th className="px-4 py-3">Account #</th>
            <th className="px-4 py-3">Email</th>
            <th className="px-4 py-3">Customer since</th>
            <th className="px-4 py-3">Status</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {customers.map((c) => (
            <tr key={c.id} className="hover:bg-gray-50">
              <td className="px-4 py-3 font-medium">
                <Link
                  href={`/customers/${c.id}`}
                  className="text-gray-900 hover:text-blue-700 hover:underline"
                >
                  {c.name}
                </Link>
              </td>
              <td className="px-4 py-3 font-mono text-xs text-gray-600">
                {c.accountNumber}
              </td>
              <td className="px-4 py-3 text-gray-600">{c.email}</td>
              <td className="px-4 py-3 text-gray-600">
                {formatDate(c.customerSince)}
              </td>
              <td className="px-4 py-3">
                <StatusBadge status={c.status} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
