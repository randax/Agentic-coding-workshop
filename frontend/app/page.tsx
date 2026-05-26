import { getCustomers, type Customer, API_BASE_URL } from "@/lib/api";

const statusStyles: Record<Customer["status"], string> = {
  active: "bg-green-100 text-green-800 ring-green-600/20",
  suspended: "bg-amber-100 text-amber-800 ring-amber-600/20",
};

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-GB", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export default async function CustomersPage() {
  let customers: Customer[] = [];
  let error: string | null = null;

  try {
    customers = await getCustomers();
  } catch {
    error = `Could not reach the CRM API at ${API_BASE_URL}. Is the backend running?`;
  }

  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10">
      <header className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight text-gray-900">
          Customers
        </h1>
        <p className="mt-1 text-sm text-gray-500">
          {error
            ? "—"
            : `${customers.length} customer${customers.length === 1 ? "" : "s"}`}
        </p>
      </header>

      {error ? (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
          {error}
        </div>
      ) : customers.length === 0 ? (
        <div className="rounded-lg border border-dashed border-gray-300 p-10 text-center text-sm text-gray-500">
          No customers yet.
        </div>
      ) : (
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
                  <td className="px-4 py-3 font-medium text-gray-900">
                    {c.name}
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-gray-600">
                    {c.accountNumber}
                  </td>
                  <td className="px-4 py-3 text-gray-600">{c.email}</td>
                  <td className="px-4 py-3 text-gray-600">
                    {formatDate(c.customerSince)}
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${statusStyles[c.status]}`}
                    >
                      {c.status}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </main>
  );
}
