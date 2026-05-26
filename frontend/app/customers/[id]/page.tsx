import Link from "next/link";
import { notFound } from "next/navigation";
import {
  getCustomer,
  getCustomerSubscriptions,
  getProducts,
  type Subscription,
  type Product,
  API_BASE_URL,
} from "@/lib/api";
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

  // Only the Subscriptions tab needs this data; fetch it lazily. The add form
  // offers products still available in the catalog (retired ones can't be sold).
  let subscriptions: Subscription[] = [];
  let availableProducts: Product[] = [];
  if (activeTab === "subscriptions") {
    try {
      const [subs, products] = await Promise.all([
        getCustomerSubscriptions(id),
        getProducts(),
      ]);
      subscriptions = subs;
      availableProducts = products.filter((p) => p.available);
    } catch {
      subscriptions = [];
      availableProducts = [];
    }
  }

  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10">
      <div className="flex items-center justify-between gap-4">
        <Link href="/" className="text-sm text-gray-500 hover:text-gray-800">
          ← Customers
        </Link>
        <Link
          href={`/customers/${customer.id}/edit`}
          className="rounded-md border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50"
        >
          Edit
        </Link>
      </div>
      <div className="mt-4">
        <CustomerDetail
          customer={customer}
          activeTab={activeTab}
          subscriptions={subscriptions}
          availableProducts={availableProducts}
        />
      </div>
    </main>
  );
}
