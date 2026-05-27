import { notFound } from "next/navigation";
import { cookies } from "next/headers";
import Link from "next/link";
import {
  getModuleMeta,
  getModuleRecord,
  getSubpanelRecords,
  type ModuleMeta,
  type ModuleRecord,
  API_BASE_URL,
} from "@/lib/api";
import RecordView, { type SubpanelData } from "../RecordView";

export default async function ModuleRecordPage({
  params,
}: {
  params: Promise<{ module: string; id: string }>;
}) {
  const { module, id } = await params;
  const cookie = (await cookies()).toString();

  let meta: ModuleMeta;
  let record: ModuleRecord | null;
  let subpanels: SubpanelData[] = [];

  try {
    [meta, record] = await Promise.all([
      getModuleMeta(module, cookie),
      getModuleRecord(module, id, cookie),
    ]);
    if (record === null) notFound();
    subpanels = await Promise.all(
      (meta.subpanels ?? []).map(async (sp) => ({
        meta: sp,
        records: await getSubpanelRecords(sp.path, id, cookie),
      })),
    );
  } catch (e) {
    // notFound() throws a special error that must propagate.
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
        href={`/m/${module}`}
        className="text-sm text-gray-500 hover:text-gray-800"
      >
        ← {meta.label}
      </Link>
      <div className="mt-4">
        <RecordView meta={meta} record={record!} subpanels={subpanels} />
      </div>
    </main>
  );
}
