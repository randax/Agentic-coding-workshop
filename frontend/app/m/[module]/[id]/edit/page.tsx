import { notFound } from "next/navigation";
import Link from "next/link";
import {
  getModuleMeta,
  getModuleRecord,
  type ModuleMeta,
  type ModuleRecord,
  API_BASE_URL,
} from "@/lib/api";
import EditView from "../../EditView";

export default async function ModuleEditPage({
  params,
}: {
  params: Promise<{ module: string; id: string }>;
}) {
  const { module, id } = await params;

  let meta: ModuleMeta;
  let record: ModuleRecord | null;

  try {
    [meta, record] = await Promise.all([
      getModuleMeta(module),
      getModuleRecord(module, id),
    ]);
    if (record === null) notFound();
  } catch (e) {
    if (e && typeof e === "object" && "digest" in e) throw e;
    return (
      <main className="mx-auto w-full max-w-5xl px-6 py-10">
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
          Could not load this {module} record from the API at {API_BASE_URL}.
        </div>
      </main>
    );
  }

  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10">
      <Link
        href={`/m/${module}/${id}`}
        className="text-sm text-gray-500 hover:text-gray-800"
      >
        ← Back
      </Link>
      <h1 className="mt-4 text-2xl font-semibold tracking-tight text-gray-900">
        Edit {meta.labelSingular.toLowerCase()}
      </h1>
      <div className="mt-6">
        <EditView meta={meta} record={record!} />
      </div>
    </main>
  );
}
