import Link from "next/link";
import { notFound } from "next/navigation";
import {
  getCase,
  getAgents,
  type Case,
  type Agent,
  API_BASE_URL,
} from "@/lib/api";
import CaseDetail from "./CaseDetail";
import AddCommentForm from "./AddCommentForm";

export default async function CaseDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let caseItem: Case | null;
  try {
    caseItem = await getCase(id);
  } catch {
    return (
      <main className="mx-auto w-full max-w-3xl px-6 py-10">
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
          Could not reach the CRM API at {API_BASE_URL}. Is the backend running?
        </div>
      </main>
    );
  }

  if (!caseItem) {
    notFound();
  }

  let agents: Agent[] = [];
  try {
    agents = await getAgents();
  } catch {
    agents = [];
  }

  return (
    <main className="mx-auto w-full max-w-3xl px-6 py-10">
      <Link
        href={`/customers/${caseItem.customerId}?tab=cases`}
        className="text-sm text-gray-500 hover:text-gray-800"
      >
        ← Back to cases
      </Link>
      <div className="mt-4">
        <CaseDetail caseItem={caseItem} />
      </div>
      <div className="mt-8">
        <AddCommentForm caseId={caseItem.id} agents={agents} />
      </div>
    </main>
  );
}
