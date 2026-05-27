"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { convertLead } from "@/lib/api";

/**
 * Header action for a lead record. It owns the convert lifecycle's visibility:
 *   • already converted → a banner linking to the account it became;
 *   • qualified & unconverted → the Convert button, which converts the lead and
 *     redirects to the new account;
 *   • anything else → nothing (you can't convert a raw/working/unqualified lead).
 */
export default function ConvertLeadButton({
  leadId,
  status,
  convertedAccountId,
}: {
  leadId: number;
  status: string;
  convertedAccountId?: number | null;
}) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (convertedAccountId) {
    return (
      <Link
        href={`/m/accounts/${convertedAccountId}`}
        className="rounded-md border border-green-200 bg-green-50 px-3 py-2 text-sm font-medium text-green-800 hover:bg-green-100"
      >
        Converted → account #{convertedAccountId}
      </Link>
    );
  }

  if (status !== "qualified") {
    return null;
  }

  async function handleConvert() {
    setBusy(true);
    setError(null);
    try {
      const { accountId } = await convertLead(leadId);
      router.push(`/m/accounts/${accountId}`);
    } catch {
      setError("Could not convert this lead. Is the backend running?");
      setBusy(false);
    }
  }

  return (
    <div className="flex flex-col items-end gap-1">
      <button
        type="button"
        onClick={handleConvert}
        disabled={busy}
        className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-500 disabled:opacity-50"
      >
        {busy ? "Converting…" : "Convert"}
      </button>
      {error && (
        <p role="alert" className="text-xs text-red-700">
          {error}
        </p>
      )}
    </div>
  );
}
