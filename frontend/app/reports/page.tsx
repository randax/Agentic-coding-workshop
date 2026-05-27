import { cookies } from "next/headers";
import {
  getModuleMeta,
  getSavedReports,
  type SavedReport,
  API_BASE_URL,
} from "@/lib/api";
import ReportBuilder, { type ReportModule } from "./ReportBuilder";
import SavedReports from "./SavedReports";

// The modules a report can run over (kept in sync with the backend's report
// handler). Their fields — including any Studio custom fields — are loaded from
// each module's metadata, so custom fields are reportable without code.
const REPORT_MODULES: { name: string; label: string }[] = [
  { name: "accounts", label: "Accounts" },
  { name: "contacts", label: "Contacts" },
  { name: "leads", label: "Leads" },
  { name: "opportunities", label: "Opportunities" },
];

/**
 * Reports landing page. This server component forwards the session cookie so
 * metadata and saved reports are scoped to the signed-in user, then renders the
 * builder (to compose and run reports) and the saved-report list (to re-run them).
 */
export default async function ReportsPage() {
  const cookie = (await cookies()).toString();

  let modules: ReportModule[] = [];
  let saved: SavedReport[] = [];
  let error: string | null = null;
  try {
    modules = await Promise.all(
      REPORT_MODULES.map(async (m) => ({
        name: m.name,
        label: m.label,
        fields: (await getModuleMeta(m.name, cookie)).fields,
      })),
    );
    saved = await getSavedReports(cookie);
  } catch {
    error = `Could not reach the API at ${API_BASE_URL}. Is the backend running?`;
  }

  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10">
      <header className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight text-gray-900">
          Reports
        </h1>
        <p className="mt-1 text-sm text-gray-500">
          Build a report over a module with filters, grouping, and an aggregation;
          chart the result and save it to re-run later.
        </p>
      </header>

      {error ? (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
          {error}
        </div>
      ) : (
        <div className="space-y-12">
          <section>
            <h2 className="mb-4 text-lg font-semibold text-gray-900">Build a report</h2>
            <ReportBuilder modules={modules} />
          </section>

          <section>
            <h2 className="mb-4 text-lg font-semibold text-gray-900">Saved reports</h2>
            <SavedReports reports={saved} />
          </section>
        </div>
      )}
    </main>
  );
}
