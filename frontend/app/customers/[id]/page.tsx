import Link from "next/link";
import { notFound } from "next/navigation";
import { getCustomer, API_BASE_URL } from "@/lib/api";
import CustomerDetail, { type TabKey } from "./CustomerDetail";

const VALID_TABS: TabKey[] = ["profile", "subscriptions", "cases"];

function resolveTab(tab: string | undefined): TabKey {
  return VALID_TABS.includes(tab as TabKey) ? (tab as TabKey) : "profile";
}

export default async function CustomerDetailPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ tab?: string }>;
}) {
  const [{ id }, { tab }] = await Promise.all([params, searchParams]);
  const activeTab = resolveTab(tab);

  let customer;
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
        href="/"
        className="text-sm text-gray-500 hover:text-gray-800"
      >
        ← Customers
      </Link>
      <div className="mt-4">
        <CustomerDetail customer={customer} activeTab={activeTab} />
      </div>
    </main>
  );
}
