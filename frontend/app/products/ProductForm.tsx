"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  createProduct,
  updateProduct,
  type Product,
  type ProductInput,
  type ProductCategory,
} from "@/lib/api";

const CATEGORIES: { value: ProductCategory; label: string }[] = [
  { value: "fiber", label: "Fiber" },
  { value: "router", label: "Router" },
  { value: "tv", label: "TV" },
];

/**
 * Create/edit form for a catalog product. Passing `product` switches the form
 * into edit mode (fields pre-filled, submit calls updateProduct); otherwise it
 * creates a new product. The per-category attribute field changes with the
 * selected category. On success it refreshes the page and calls `onSaved`.
 */
export default function ProductForm({
  product,
  onSaved,
}: {
  product?: Product;
  onSaved?: () => void;
}) {
  const router = useRouter();
  const isEdit = product !== undefined;

  const [name, setName] = useState(product?.name ?? "");
  const [category, setCategory] = useState<ProductCategory>(
    product?.category ?? "fiber",
  );
  const [monthlyPrice, setMonthlyPrice] = useState(
    product ? String(product.monthlyPrice) : "",
  );
  const [speedMbps, setSpeedMbps] = useState(
    product?.speedMbps != null ? String(product.speedMbps) : "",
  );
  const [routerModel, setRouterModel] = useState(product?.routerModel ?? "");
  const [tvPackageTier, setTvPackageTier] = useState(
    product?.tvPackageTier ?? "",
  );
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    const input: ProductInput = {
      name,
      category,
      monthlyPrice: Number(monthlyPrice),
    };
    if (category === "fiber" && speedMbps) input.speedMbps = Number(speedMbps);
    if (category === "router" && routerModel) input.routerModel = routerModel;
    if (category === "tv" && tvPackageTier) input.tvPackageTier = tvPackageTier;
    try {
      if (isEdit && product) {
        await updateProduct(product.id, input);
      } else {
        await createProduct(input);
        setName("");
        setMonthlyPrice("");
        setSpeedMbps("");
        setRouterModel("");
        setTvPackageTier("");
      }
      router.refresh();
      onSaved?.();
    } catch {
      setError(
        "Could not save the product. Check the fields and that the backend is running.",
      );
    } finally {
      setBusy(false);
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="space-y-4 rounded-lg border border-gray-200 bg-gray-50 p-4"
    >
      {error && (
        <div
          role="alert"
          className="rounded-md border border-red-200 bg-red-50 p-2 text-sm text-red-800"
        >
          {error}
        </div>
      )}

      <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
        Name
        <input
          aria-label="Name"
          value={name}
          required
          onChange={(e) => setName(e.target.value)}
          className="rounded-md border border-gray-300 px-3 py-2 text-sm normal-case text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
      </label>

      <div className="flex flex-wrap gap-4">
        <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
          Category
          <select
            aria-label="Category"
            value={category}
            onChange={(e) => setCategory(e.target.value as ProductCategory)}
            className="rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
            {CATEGORIES.map((c) => (
              <option key={c.value} value={c.value}>
                {c.label}
              </option>
            ))}
          </select>
        </label>

        <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
          Monthly price
          <input
            aria-label="Monthly price"
            type="number"
            min={0}
            value={monthlyPrice}
            required
            onChange={(e) => setMonthlyPrice(e.target.value)}
            className="w-32 rounded-md border border-gray-300 px-3 py-2 text-sm normal-case text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          />
        </label>

        {category === "fiber" && (
          <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
            Speed (Mbps)
            <input
              aria-label="Speed (Mbps)"
              type="number"
              min={0}
              value={speedMbps}
              onChange={(e) => setSpeedMbps(e.target.value)}
              className="w-32 rounded-md border border-gray-300 px-3 py-2 text-sm normal-case text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </label>
        )}

        {category === "router" && (
          <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
            Router model
            <input
              aria-label="Router model"
              value={routerModel}
              onChange={(e) => setRouterModel(e.target.value)}
              className="rounded-md border border-gray-300 px-3 py-2 text-sm normal-case text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </label>
        )}

        {category === "tv" && (
          <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
            TV package tier
            <input
              aria-label="TV package tier"
              value={tvPackageTier}
              onChange={(e) => setTvPackageTier(e.target.value)}
              className="rounded-md border border-gray-300 px-3 py-2 text-sm normal-case text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </label>
        )}
      </div>

      <button
        type="submit"
        disabled={busy}
        className="rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-gray-800 disabled:opacity-50"
      >
        {isEdit ? "Save changes" : "Create product"}
      </button>
    </form>
  );
}
