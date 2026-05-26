import type { Subscription } from "@/lib/api";
import { formatDate, formatMonthlyPrice } from "@/lib/format";

function period(s: Subscription): string {
  const start = formatDate(s.startDate);
  return s.endDate ? `${start} – ${formatDate(s.endDate)}` : `${start} –`;
}

export default function SubscriptionList({
  subscriptions,
}: {
  subscriptions: Subscription[];
}) {
  if (subscriptions.length === 0) {
    return (
      <div className="rounded-lg border border-dashed border-gray-300 p-10 text-center text-sm text-gray-500">
        This customer has no subscriptions.
      </div>
    );
  }

  return (
    <table className="min-w-full divide-y divide-gray-200 text-sm">
      <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
        <tr>
          <th className="px-4 py-3">Product</th>
          <th className="px-4 py-3">Status</th>
          <th className="px-4 py-3">Qty</th>
          <th className="px-4 py-3">Price</th>
          <th className="px-4 py-3">Period</th>
        </tr>
      </thead>
      <tbody className="divide-y divide-gray-100">
        {subscriptions.map((s) => (
          <tr key={s.id} className="hover:bg-gray-50">
            <td className="px-4 py-3 font-medium text-gray-900">
              {s.product.name}
            </td>
            <td className="px-4 py-3 text-gray-600">{s.status}</td>
            <td className="px-4 py-3 text-gray-600">{s.quantity}</td>
            <td className="px-4 py-3 text-gray-600">
              {formatMonthlyPrice(s.monthlyPriceSnapshot)}
            </td>
            <td className="px-4 py-3 text-gray-600">{period(s)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
