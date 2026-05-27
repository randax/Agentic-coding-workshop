import {
  getModuleMeta,
  getModuleRecords,
  type ModuleMeta,
  type ModuleRecord,
  API_BASE_URL,
} from "@/lib/api";
import ListView from "./ListView";

/**
 * Generic module list page. Renders any registered module's records as a table
 * driven entirely by the module's metadata — no module-specific code here.
 */
export default async function ModuleListPage({
  params,
}: {
  params: Promise<{ module: string }>;
}) {
  const { module } = await params;

  let meta: ModuleMeta;
  let records: ModuleRecord[] = [];

  try {
    [meta, records] = await Promise.all([
      getModuleMeta(module),
      getModuleRecords(module),
    ]);
  } catch {
    return (
      <main className="mx-auto w-full max-w-5xl px-6 py-10">
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
          Could not load the “{module}” module from the API at {API_BASE_URL}.
          Is it a known module and is the backend running?
        </div>
      </main>
    );
  }

  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10">
      <header className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight text-gray-900">
          {meta.label}
        </h1>
        <p className="mt-1 text-sm text-gray-500">
          {`${records.length} ${records.length === 1 ? meta.labelSingular.toLowerCase() : meta.label.toLowerCase()}`}
        </p>
      </header>

      {records.length === 0 ? (
        <div className="rounded-lg border border-dashed border-gray-300 p-10 text-center text-sm text-gray-500">
          No {meta.label.toLowerCase()} yet.
        </div>
      ) : (
        <div className="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm">
          <ListView meta={meta} records={records} />
        </div>
      )}
    </main>
  );
}
