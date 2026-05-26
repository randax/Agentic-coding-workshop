import { getProducts, type Product, API_BASE_URL } from "@/lib/api";
import ProductCatalog from "./ProductCatalog";
import ProductForm from "./ProductForm";

export default async function ProductsPage() {
  let products: Product[] = [];
  let error: string | null = null;

  try {
    products = await getProducts();
  } catch {
    error = `Could not reach the CRM API at ${API_BASE_URL}. Is the backend running?`;
  }

  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10">
      <header className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight text-gray-900">
          Products
        </h1>
        <p className="mt-1 text-sm text-gray-500">
          {error
            ? "—"
            : `${products.length} product${products.length === 1 ? "" : "s"} in the catalog`}
        </p>
      </header>

      {error ? (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
          {error}
        </div>
      ) : (
        <div className="space-y-8">
          <section>
            <h2 className="mb-3 text-sm font-semibold text-gray-900">
              New product
            </h2>
            <ProductForm />
          </section>

          {products.length === 0 ? (
            <div className="rounded-lg border border-dashed border-gray-300 p-10 text-center text-sm text-gray-500">
              No products yet.
            </div>
          ) : (
            <div className="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm">
              <ProductCatalog products={products} />
            </div>
          )}
        </div>
      )}
    </main>
  );
}
