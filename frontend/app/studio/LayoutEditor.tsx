"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  getLayouts,
  getModuleMeta,
  saveLayouts,
  type ModuleMeta,
} from "@/lib/api";

/** One field row in the editor: its name, display label, and whether it shows. */
type Row = { name: string; label: string; visible: boolean };
/** A detail-view panel (code-defined) with its editable rows. */
type PanelRows = { label: string; rows: Row[] };

/**
 * Admin editor for a module's view layouts. For each generic view (list, detail,
 * edit) it shows every field with a visibility checkbox and Up/Down reorder
 * buttons, pre-filled from the saved layout (or the module's code defaults). The
 * design-time palette comes from the module's raw metadata, so currently-hidden
 * fields still appear and can be re-enabled. Detail keeps code-defined panel
 * groupings; reordering is within a panel. Saving persists each view's ordered,
 * visible field names so the generic views honor them for everyone.
 */
export default function LayoutEditor({ module }: { module: string }) {
  const router = useRouter();
  const [list, setList] = useState<Row[]>([]);
  const [edit, setEdit] = useState<Row[]>([]);
  const [detail, setDetail] = useState<PanelRows[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    setSaved(false);
    Promise.all([getModuleMeta(module, undefined, { raw: true }), getLayouts(module)])
      .then(([meta, layouts]) => {
        if (cancelled) return;
        const labels = labelMap(meta);
        const allNames = meta.fields.map((f) => f.name);
        setList(flatRows(allNames, labels, layouts.list ?? meta.listView.columns));
        setEdit(flatRows(allNames, labels, layouts.edit ?? meta.editView?.fields ?? []));
        setDetail(panelRows(meta.detailView?.panels ?? [], labels, layouts.detail));
        setLoading(false);
      })
      .catch(() => {
        if (cancelled) return;
        setError("Could not load this module's layout.");
        setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [module]);

  async function handleSave() {
    setBusy(true);
    setError(null);
    setSaved(false);
    try {
      await saveLayouts(module, {
        list: visibleNames(list),
        edit: visibleNames(edit),
        detail: detail.flatMap((p) => visibleNames(p.rows)),
      });
      setSaved(true);
      router.refresh();
    } catch {
      setError("Could not save the layout. Admin access is required.");
    } finally {
      setBusy(false);
    }
  }

  if (loading) {
    return <p className="text-sm text-gray-500">Loading layout…</p>;
  }

  return (
    <div className="space-y-6">
      {error && (
        <div role="alert" className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800">
          {error}
        </div>
      )}
      {saved && (
        <div className="rounded-lg border border-green-200 bg-green-50 p-3 text-sm text-green-800">
          Layout saved. The {module} views now use it for everyone.
        </div>
      )}

      <ViewSection label="List columns">
        <RowList rows={list} onToggle={toggler(setList)} onMove={mover(setList)} />
      </ViewSection>

      <ViewSection label="Detail fields">
        <div className="space-y-4">
          {detail.map((panel, pi) => (
            <div key={panel.label}>
              <h4 className="mb-1 text-xs font-medium uppercase tracking-wide text-gray-400">
                {panel.label}
              </h4>
              <RowList
                rows={panel.rows}
                onToggle={(i) => setDetail((d) => updatePanel(d, pi, (rows) => toggle(rows, i)))}
                onMove={(i, dir) => setDetail((d) => updatePanel(d, pi, (rows) => move(rows, i, dir)))}
              />
            </div>
          ))}
        </div>
      </ViewSection>

      <ViewSection label="Edit fields">
        <RowList rows={edit} onToggle={toggler(setEdit)} onMove={mover(setEdit)} />
      </ViewSection>

      <button
        type="button"
        onClick={handleSave}
        disabled={busy}
        className="rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-gray-800 disabled:opacity-50"
      >
        Save layout
      </button>
    </div>
  );
}

function ViewSection({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <section aria-label={label} className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
      <h3 className="mb-2 text-sm font-semibold text-gray-900">{label}</h3>
      {children}
    </section>
  );
}

function RowList({
  rows,
  onToggle,
  onMove,
}: {
  rows: Row[];
  onToggle: (i: number) => void;
  onMove: (i: number, dir: number) => void;
}) {
  const arrow =
    "rounded border border-gray-300 px-2 py-0.5 text-xs text-gray-600 hover:bg-gray-100 disabled:opacity-30";
  return (
    <ul className="divide-y divide-gray-100">
      {rows.map((r, i) => (
        <li key={r.name} className="flex items-center gap-3 py-1.5">
          <input
            type="checkbox"
            aria-label={r.label}
            checked={r.visible}
            onChange={() => onToggle(i)}
            className="h-4 w-4"
          />
          <span className="flex-1 text-sm text-gray-700">{r.label}</span>
          <button type="button" aria-label={`Move ${r.label} up`} onClick={() => onMove(i, -1)} disabled={i === 0} className={arrow}>
            ↑
          </button>
          <button type="button" aria-label={`Move ${r.label} down`} onClick={() => onMove(i, 1)} disabled={i === rows.length - 1} className={arrow}>
            ↓
          </button>
        </li>
      ))}
    </ul>
  );
}

// --- pure helpers ---------------------------------------------------------

function labelMap(meta: ModuleMeta): Map<string, string> {
  return new Map(meta.fields.map((f) => [f.name, f.label]));
}

/** Visible fields (in saved/default order) first, then the remaining fields as
 * hidden rows so they can be re-enabled. */
function flatRows(allNames: string[], labels: Map<string, string>, order: string[]): Row[] {
  const shown = order.filter((n) => labels.has(n));
  const hidden = allNames.filter((n) => !shown.includes(n));
  return [
    ...shown.map((n) => ({ name: n, label: labels.get(n) ?? n, visible: true })),
    ...hidden.map((n) => ({ name: n, label: labels.get(n) ?? n, visible: false })),
  ];
}

/** Detail rows grouped by code-defined panel; a saved detail order reorders and
 * hides within each panel (cross-panel membership is fixed by code). */
function panelRows(
  panels: { label: string; fields: string[] }[],
  labels: Map<string, string>,
  savedDetail: string[] | undefined,
): PanelRows[] {
  return panels.map((p) => {
    const shown = savedDetail ? savedDetail.filter((n) => p.fields.includes(n)) : p.fields;
    const hidden = p.fields.filter((n) => !shown.includes(n));
    return {
      label: p.label,
      rows: [
        ...shown.map((n) => ({ name: n, label: labels.get(n) ?? n, visible: true })),
        ...hidden.map((n) => ({ name: n, label: labels.get(n) ?? n, visible: false })),
      ],
    };
  });
}

const visibleNames = (rows: Row[]) => rows.filter((r) => r.visible).map((r) => r.name);

function toggle(rows: Row[], i: number): Row[] {
  return rows.map((r, idx) => (idx === i ? { ...r, visible: !r.visible } : r));
}

function move(rows: Row[], i: number, dir: number): Row[] {
  const j = i + dir;
  if (j < 0 || j >= rows.length) return rows;
  const next = [...rows];
  [next[i], next[j]] = [next[j], next[i]];
  return next;
}

function updatePanel(panels: PanelRows[], pi: number, fn: (rows: Row[]) => Row[]): PanelRows[] {
  return panels.map((p, idx) => (idx === pi ? { ...p, rows: fn(p.rows) } : p));
}

const toggler = (setter: React.Dispatch<React.SetStateAction<Row[]>>) => (i: number) =>
  setter((rows) => toggle(rows, i));
const mover = (setter: React.Dispatch<React.SetStateAction<Row[]>>) => (i: number, dir: number) =>
  setter((rows) => move(rows, i, dir));
