"use client";

import { Fragment, useState } from "react";
import { useRouter } from "next/navigation";
import { retireProduct, unretireProduct, type Product } from "@/lib/api";
import { formatMonthlyPrice } from "@/lib/format";
import ProductForm from "./ProductForm";

function detail(p: Product): string {
  switch (p.category) {
    case "fiber":
      return p.speedMbps ? `${p.speedMbps} Mbps` : "—";
    case "router":
      return p.routerModel ?? "—";
    case "tv":
      return p.tvPackageTier ?? "—";
    default:
      return "—";
  }
}

export default function ProductCatalog({ products }: { products: Product[] }) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);

  async function handleRetire(id: number) {
    setBusy(true);
    try {
      await retireProduct(id);
      router.refresh();
    } finally {
      setBusy(false);
    }
  }

  async function handleUnretire(id: number) {
    setBusy(true);
    try {
      await unretireProduct(id);
      router.refresh();
    } finally {
      setBusy(false);
    }
  }

  return (
    <table className="min-w-full divide-y divide-gray-200 text-sm">
      <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
        <tr>
          <th className="px-4 py-3">Name</th>
          <th className="px-4 py-3">Category</th>
          <th className="px-4 py-3">Details</th>
          <th className="px-4 py-3">Price</th>
          <th className="px-4 py-3">Status</th>
          <th className="px-4 py-3" />
        </tr>
      </thead>
      <tbody className="divide-y divide-gray-100">
        {products.map((p) => (
          <Fragment key={p.id}>
            <tr className="hover:bg-gray-50">
              <td className="px-4 py-3 font-medium text-gray-900">{p.name}</td>
              <td className="px-4 py-3 capitalize text-gray-600">
                {p.category}
              </td>
              <td className="px-4 py-3 text-gray-600">{detail(p)}</td>
              <td className="px-4 py-3 text-gray-600">
                {formatMonthlyPrice(p.monthlyPrice)}
              </td>
              <td className="px-4 py-3">
                {p.available ? (
                  <span className="inline-flex items-center rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-800 ring-1 ring-inset ring-green-600/20">
                    Available
                  </span>
                ) : (
                  <span className="inline-flex items-center rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/20">
                    Retired
                  </span>
                )}
              </td>
              <td className="px-4 py-3 text-right">
                <div className="flex justify-end gap-4">
                  <button
                    type="button"
                    onClick={() =>
                      setEditingId(editingId === p.id ? null : p.id)
                    }
                    className="text-sm font-medium text-gray-700 hover:text-gray-900"
                  >
                    Edit
                  </button>
                  {p.available ? (
                    <button
                      type="button"
                      onClick={() => handleRetire(p.id)}
                      disabled={busy}
                      className="text-sm font-medium text-red-700 hover:text-red-900 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      Retire
                    </button>
                  ) : (
                    <button
                      type="button"
                      onClick={() => handleUnretire(p.id)}
                      disabled={busy}
                      className="text-sm font-medium text-green-700 hover:text-green-900 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      Unretire
                    </button>
                  )}
                </div>
              </td>
            </tr>
            {editingId === p.id && (
              <tr>
                <td colSpan={6} className="bg-gray-50 px-4 py-4">
                  <ProductForm
                    product={p}
                    onSaved={() => setEditingId(null)}
                  />
                </td>
              </tr>
            )}
          </Fragment>
        ))}
      </tbody>
    </table>
  );
}
