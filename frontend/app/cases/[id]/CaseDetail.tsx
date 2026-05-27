import type { Case } from "@/lib/api";
import { formatDate } from "@/lib/format";
import { STATUS_LABELS, STATUS_STYLES } from "@/app/cases/status";

export default function CaseDetail({ caseItem }: { caseItem: Case }) {
  const comments = caseItem.comments ?? [];

  return (
    <div className="space-y-8">
      <div>
        <div className="flex flex-wrap items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight text-gray-900">
            {caseItem.subject}
          </h1>
          <span
            className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${STATUS_STYLES[caseItem.status]}`}
          >
            {STATUS_LABELS[caseItem.status]}
          </span>
        </div>
        <dl className="mt-4 grid grid-cols-2 gap-x-8 gap-y-3 sm:grid-cols-4">
          <Meta label="Category" value={caseItem.category} capitalize />
          <Meta label="Priority" value={caseItem.priority} capitalize />
          <Meta
            label="Assigned to"
            value={caseItem.assignedAgent?.name ?? "Unassigned"}
          />
          <Meta label="Opened" value={formatDate(caseItem.createdAt)} />
        </dl>
      </div>

      <section>
        <h2 className="text-xs font-medium uppercase tracking-wide text-gray-500">
          Description
        </h2>
        <p className="mt-2 whitespace-pre-line text-sm text-gray-900">
          {caseItem.description}
        </p>
      </section>

      <section>
        <h2 className="text-xs font-medium uppercase tracking-wide text-gray-500">
          Timeline
        </h2>
        {comments.length === 0 ? (
          <p className="mt-2 text-sm text-gray-500">No comments yet.</p>
        ) : (
          <ol className="mt-3 space-y-4">
            {comments.map((comment) => (
              <li
                key={comment.id}
                className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm"
              >
                <div className="flex items-baseline justify-between gap-3">
                  <span className="text-sm font-medium text-gray-900">
                    {comment.authorAgent?.name ?? "Unknown agent"}
                  </span>
                  <span className="text-xs text-gray-500">
                    {formatDate(comment.createdAt)}
                  </span>
                </div>
                <p className="mt-2 whitespace-pre-line text-sm text-gray-700">
                  {comment.body}
                </p>
              </li>
            ))}
          </ol>
        )}
      </section>
    </div>
  );
}

function Meta({
  label,
  value,
  capitalize = false,
}: {
  label: string;
  value: string;
  capitalize?: boolean;
}) {
  return (
    <div>
      <dt className="text-xs font-medium uppercase tracking-wide text-gray-500">
        {label}
      </dt>
      <dd className={`mt-1 text-sm text-gray-900 ${capitalize ? "capitalize" : ""}`}>
        {value}
      </dd>
    </div>
  );
}
