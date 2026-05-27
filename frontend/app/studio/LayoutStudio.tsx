"use client";

import { useState } from "react";
import LayoutEditor from "./LayoutEditor";

/** Module picker that drives the LayoutEditor for the selected module. */
export default function LayoutStudio({ modules }: { modules: string[] }) {
  const [module, setModule] = useState(modules[0] ?? "");

  return (
    <div className="space-y-5">
      <label className="flex max-w-xs flex-col gap-1 text-sm font-medium text-gray-700">
        Module
        <select
          aria-label="Layout module"
          value={module}
          onChange={(e) => setModule(e.target.value)}
          className="rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        >
          {modules.map((m) => (
            <option key={m} value={m}>
              {m}
            </option>
          ))}
        </select>
      </label>

      {module && <LayoutEditor key={module} module={module} />}
    </div>
  );
}
