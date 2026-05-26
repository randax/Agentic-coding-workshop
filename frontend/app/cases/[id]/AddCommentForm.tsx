"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { addCaseComment, type Agent } from "@/lib/api";

export default function AddCommentForm({
  caseId,
  agents,
}: {
  caseId: number;
  agents: Agent[];
}) {
  const router = useRouter();
  const [body, setBody] = useState("");
  const [authorAgentId, setAuthorAgentId] = useState<number>(
    agents[0]?.id ?? 0,
  );
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!body.trim() || !authorAgentId) return;
    setBusy(true);
    setError(null);
    try {
      await addCaseComment(caseId, { body, authorAgentId });
      setBody("");
      router.refresh();
    } catch {
      setError("Could not add the comment. Is the backend running?");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="space-y-3 rounded-lg border border-gray-200 bg-gray-50 p-4"
    >
      <h3 className="text-sm font-semibold text-gray-900">Add a comment</h3>

      {error && (
        <div
          role="alert"
          className="rounded-md border border-red-200 bg-red-50 p-2 text-sm text-red-800"
        >
          {error}
        </div>
      )}

      <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
        Comment
        <textarea
          aria-label="Comment"
          value={body}
          rows={3}
          required
          onChange={(e) => setBody(e.target.value)}
          className="rounded-md border border-gray-300 px-3 py-2 text-sm normal-case text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
      </label>

      <div className="flex flex-wrap items-end gap-3">
        <label className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500">
          Author
          <select
            aria-label="Author"
            value={authorAgentId}
            onChange={(e) => setAuthorAgentId(Number(e.target.value))}
            className="rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
            {agents.map((a) => (
              <option key={a.id} value={a.id}>
                {a.name}
              </option>
            ))}
          </select>
        </label>

        <button
          type="submit"
          disabled={busy}
          className="rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-gray-800 disabled:opacity-50"
        >
          Add comment
        </button>
      </div>
    </form>
  );
}
