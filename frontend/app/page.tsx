import Link from "next/link";
import {
  getCustomers,
  type Customer,
  type CustomerStatus,
  API_BASE_URL,
} from "@/lib/api";
import CustomerSearch from "./CustomerSearch";
import CustomerTable from "./CustomerTable";

/** Narrows a raw status query param to a valid CustomerStatus, or undefined. */
function parseStatus(raw: string | undefined): CustomerStatus | undefined {
  return raw === "active" || raw === "suspended" ? raw : undefined;
}

export default async function CustomersPage({
  searchParams,
}: {
  searchParams: Promise<{ search?: string; status?: string }>;
}) {
  const { search, status } = await searchParams;

  let customers: Customer[] = [];
  let error: string | null = null;

  try {
    customers = await getCustomers({ search, status: parseStatus(status) });
  } catch {
    error = `Could not reach the CRM API at ${API_BASE_URL}. Is the backend running?`;
  }

  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10">
      <header className="mb-6 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-gray-900">
            Customers
          </h1>
          <p className="mt-1 text-sm text-gray-500">
            {error
              ? "—"
              : `${customers.length} customer${customers.length === 1 ? "" : "s"}`}
          </p>
        </div>
        <Link
          href="/customers/new"
          className="shrink-0 rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-gray-800"
        >
          New customer
        </Link>
      </header>

      <div className="mb-6">
        <CustomerSearch />
      </div>

      {error ? (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
          {error}
        </div>
      ) : (
        <CustomerTable customers={customers} />
      )}
    </main>
  );
}
