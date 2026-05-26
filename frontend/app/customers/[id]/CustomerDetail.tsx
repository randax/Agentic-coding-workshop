import Link from "next/link";
import type { Customer, Subscription } from "@/lib/api";
import { formatDate } from "@/lib/format";
import StatusBadge from "@/components/StatusBadge";
import SubscriptionList from "./SubscriptionList";

export type TabKey = "profile" | "subscriptions" | "cases";

const TABS: { key: TabKey; label: string }[] = [
  { key: "profile", label: "Profile" },
  { key: "subscriptions", label: "Subscriptions" },
  { key: "cases", label: "Cases" },
];

function TabNav({
  customerId,
  active,
}: {
  customerId: number;
  active: TabKey;
}) {
  return (
    <nav className="mt-6 flex gap-6 border-b border-gray-200">
      {TABS.map(({ key, label }) => {
        const isActive = key === active;
        return (
          <Link
            key={key}
            href={`/customers/${customerId}?tab=${key}`}
            aria-current={isActive ? "page" : undefined}
            className={`-mb-px border-b-2 px-1 pb-3 text-sm font-medium ${
              isActive
                ? "border-gray-900 text-gray-900"
                : "border-transparent text-gray-500 hover:text-gray-800"
            }`}
          >
            {label}
          </Link>
        );
      })}
    </nav>
  );
}

function Field({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs font-medium uppercase tracking-wide text-gray-500">
        {label}
      </dt>
      <dd className="mt-1 text-sm text-gray-900">{value}</dd>
    </div>
  );
}

export default function CustomerDetail({
  customer,
  activeTab,
  subscriptions = [],
}: {
  customer: Customer;
  activeTab: TabKey;
  subscriptions?: Subscription[];
}) {
  return (
    <div>
      <h1 className="text-2xl font-semibold tracking-tight text-gray-900">
        {customer.name}
      </h1>
      <TabNav customerId={customer.id} active={activeTab} />
      <div className="mt-6">
        {activeTab === "profile" && <ProfilePanel customer={customer} />}
        {activeTab === "subscriptions" && (
          <SubscriptionList subscriptions={subscriptions} />
        )}
        {activeTab === "cases" && (
          <Placeholder>No support cases to display yet.</Placeholder>
        )}
      </div>
    </div>
  );
}

function ProfilePanel({ customer }: { customer: Customer }) {
  return (
    <dl className="grid grid-cols-1 gap-x-8 gap-y-6 sm:grid-cols-2">
      <Field label="Email" value={customer.email} />
      <Field label="Phone" value={customer.phone} />
      <Field label="Service address" value={customer.serviceAddress} />
      <Field label="Account number" value={customer.accountNumber} />
      <Field label="Customer since" value={formatDate(customer.customerSince)} />
      <div>
        <dt className="text-xs font-medium uppercase tracking-wide text-gray-500">
          Status
        </dt>
        <dd className="mt-1">
          <StatusBadge status={customer.status} />
        </dd>
      </div>
    </dl>
  );
}

function Placeholder({ children }: { children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-dashed border-gray-300 p-10 text-center text-sm text-gray-500">
      {children}
    </div>
  );
}
