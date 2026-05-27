"use client";

import { useState } from "react";
import { runSavedReport, type ReportResult, type SavedReport } from "@/lib/api";
import ReportChart from "./ReportChart";

/**
 * Lists the saved reports and lets the user re-run any of them, charting the
 * fresh result inline. The data (and its access scoping) is produced server-side
 * on each run.
 */
export default function SavedReports({ reports }: { reports: SavedReport[] }) {
  const [activeId, setActiveId] = useState<number | null>(null);
  const [result, setResult] = useState<ReportResult | null>(null);
  const [busyId, setBusyId] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);

  if (reports.length === 0) {
    return <p className="text-sm text-gray-500">No saved reports yet.</p>;
  }

  async function run(report: SavedReport) {
    setBusyId(report.id);
    setError(null);
    try {
      const res = await runSavedReport(report.id);
      setActiveId(report.id);
      setResult(res);
    } catch {
      setError("Could not run this report.");
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div className="flex flex-col gap-3">
      {error && (
        <div role="alert" className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800">
          {error}
        </div>
      )}
      <ul className="divide-y divide-gray-100 rounded-lg border border-gray-200 bg-white">
        {reports.map((report) => (
          <li key={report.id} className="flex flex-col gap-3 px-4 py-3">
            <div className="flex items-center justify-between">
              <div className="text-sm">
                <span className="font-medium text-gray-900">{report.name}</span>
                <span className="ml-2 text-gray-500">{report.definition.module}</span>
              </div>
              <button
                type="button"
                onClick={() => run(report)}
                disabled={busyId === report.id}
                className="rounded-md border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 disabled:opacity-50"
              >
                Run
              </button>
            </div>
            {activeId === report.id && result && <ReportChart result={result} />}
          </li>
        ))}
      </ul>
    </div>
  );
}
