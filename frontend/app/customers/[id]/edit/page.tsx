import Link from "next/link";
import { notFound } from "next/navigation";
import { getCustomer, type Customer, API_BASE_URL } from "@/lib/api";
import CustomerForm from "@/app/customers/CustomerForm";

export default async function EditCustomerPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let customer: Customer | null;
  try {
    customer = await getCustomer(id);
  } catch {
    return (
      <main className="mx-auto w-full max-w-5xl px-6 py-10">
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
          Could not reach the CRM API at {API_BASE_URL}. Is the backend running?
        </div>
      </main>
    );
  }

  if (!customer) {
    notFound();
  }

  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10">
      <Link
        href={`/customers/${customer.id}`}
        className="text-sm text-gray-500 hover:text-gray-800"
      >
        ← {customer.name}
      </Link>
      <h1 className="mt-4 text-2xl font-semibold tracking-tight text-gray-900">
        Edit customer
      </h1>
      <div className="mt-6">
        <CustomerForm customer={customer} />
      </div>
    </main>
  );
}
