import { cookies } from "next/headers";
import Link from "next/link";
import { getModuleMeta, type ModuleMeta, API_BASE_URL } from "@/lib/api";
import EditView from "../EditView";

/**
 * Generic create page. Loads the module's metadata and renders the shared
 * create/edit form with no record, so EditView POSTs a new record and then
 * navigates to it. Knows nothing about any specific module.
 */
export default async function ModuleNewPage({
  params,
}: {
  params: Promise<{ module: string }>;
}) {
  const { module } = await params;
  const cookie = (await cookies()).toString();

  let meta: ModuleMeta;
  try {
    meta = await getModuleMeta(module, cookie);
  } catch {
    return (
      <main className="mx-auto w-full max-w-5xl px-6 py-10">
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
          Could not load the “{module}” module from the API at {API_BASE_URL}.
        </div>
      </main>
    );
  }

  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10">
      <Link href={`/m/${module}`} className="text-sm text-gray-500 hover:text-gray-800">
        ← Back
      </Link>
      <h1 className="mt-4 text-2xl font-semibold tracking-tight text-gray-900">
        New {meta.labelSingular.toLowerCase()}
      </h1>
      <div className="mt-6">
        <EditView meta={meta} />
      </div>
    </main>
  );
}
