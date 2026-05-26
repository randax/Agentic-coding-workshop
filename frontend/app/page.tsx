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
      <header className="mb-6">
        <h1 className="text-2xl font-semibold tracking-tight text-gray-900">
          Customers
        </h1>
        <p className="mt-1 text-sm text-gray-500">
          {error
            ? "—"
            : `${customers.length} customer${customers.length === 1 ? "" : "s"}`}
        </p>
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
