import { cookies } from "next/headers";
import { getPipeline, type PipelineStage, API_BASE_URL } from "@/lib/api";
import PipelineBoard from "./PipelineBoard";

export default async function PipelinePage() {
  const cookie = (await cookies()).toString();

  let stages: PipelineStage[] = [];
  let error: string | null = null;
  try {
    stages = await getPipeline(cookie);
  } catch {
    error = `Could not load the pipeline from the API at ${API_BASE_URL}.`;
  }

  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10">
      <header className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight text-gray-900">
          Pipeline
        </h1>
        <p className="mt-1 text-sm text-gray-500">Opportunities by sales stage</p>
      </header>

      {error ? (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
          {error}
        </div>
      ) : (
        <PipelineBoard stages={stages} />
      )}
    </main>
  );
}
