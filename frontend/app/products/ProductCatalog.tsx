import type { Product } from "@/lib/api";
import { formatMonthlyPrice } from "@/lib/format";

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
  return (
    <table className="min-w-full divide-y divide-gray-200 text-sm">
      <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
        <tr>
          <th className="px-4 py-3">Name</th>
          <th className="px-4 py-3">Category</th>
          <th className="px-4 py-3">Details</th>
          <th className="px-4 py-3">Price</th>
          <th className="px-4 py-3">Status</th>
        </tr>
      </thead>
      <tbody className="divide-y divide-gray-100">
        {products.map((p) => (
          <tr key={p.id} className="hover:bg-gray-50">
            <td className="px-4 py-3 font-medium text-gray-900">{p.name}</td>
            <td className="px-4 py-3 capitalize text-gray-600">{p.category}</td>
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
          </tr>
        ))}
      </tbody>
    </table>
  );
}
