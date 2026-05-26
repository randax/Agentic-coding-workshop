"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  createSubscription,
  cancelSubscription,
  type Subscription,
  type Product,
} from "@/lib/api";
import { formatDate, formatMonthlyPrice } from "@/lib/format";

function period(s: Subscription): string {
  const start = formatDate(s.startDate);
  return s.endDate ? `${start} – ${formatDate(s.endDate)}` : `${start} –`;
}

export default function SubscriptionList({
  customerId,
  subscriptions,
  availableProducts,
}: {
  customerId: number;
  subscriptions: Subscription[];
  availableProducts: Product[];
}) {
  const router = useRouter();
  const [productId, setProductId] = useState<number>(
    availableProducts[0]?.id ?? 0,
  );
  const [quantity, setQuantity] = useState<number>(1);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault();
    if (!productId || quantity < 1) return;
    setBusy(true);
    setError(null);
    try {
      await createSubscription(customerId, { productId, quantity });
      setQuantity(1);
      router.refresh();
    } catch {
      setError("Could not add the subscription. Is the backend running?");
    } finally {
      setBusy(false);
    }
  }

  async function handleCancel(id: number) {
    setBusy(true);
    setError(null);
    try {
      await cancelSubscription(id);
      router.refresh();
    } catch {
      setError("Could not cancel the subscription. Is the backend running?");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-6">
      <form
        onSubmit={handleAdd}
        className="flex flex-wrap items-end gap-3 rounded-lg border border-gray-200 bg-gray-50 p-4"
      >
        <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
          Product
          <select
            aria-label="Product"
            value={productId}
            onChange={(e) => setProductId(Number(e.target.value))}
            disabled={availableProducts.length === 0}
            className="min-w-56 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-normal normal-case text-gray-900"
          >
            {availableProducts.length === 0 ? (
              <option value={0}>No products available</option>
            ) : (
              availableProducts.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name} — {formatMonthlyPrice(p.monthlyPrice)}
                </option>
              ))
            )}
          </select>
        </label>

        <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
          Quantity
          <input
            aria-label="Quantity"
            type="number"
            min={1}
            value={quantity}
            onChange={(e) => setQuantity(Number(e.target.value))}
            className="w-24 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-normal text-gray-900"
          />
        </label>

        <button
          type="submit"
          disabled={busy || productId === 0}
          className="rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Add subscription
        </button>
      </form>

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800">
          {error}
        </div>
      )}

      {subscriptions.length === 0 ? (
        <div className="rounded-lg border border-dashed border-gray-300 p-10 text-center text-sm text-gray-500">
          This customer has no subscriptions.
        </div>
      ) : (
        <table className="min-w-full divide-y divide-gray-200 text-sm">
          <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
            <tr>
              <th className="px-4 py-3">Product</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Qty</th>
              <th className="px-4 py-3">Price</th>
              <th className="px-4 py-3">Period</th>
              <th className="px-4 py-3" />
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
                <td className="px-4 py-3 text-right">
                  {s.status !== "cancelled" && (
                    <button
                      type="button"
                      onClick={() => handleCancel(s.id)}
                      disabled={busy}
                      className="text-sm font-medium text-red-700 hover:text-red-900 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      Cancel
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
