import { cookies } from "next/headers";
import { getDashboard, type DashboardData, API_BASE_URL } from "@/lib/api";
import Dashboard from "./Dashboard";

/**
 * Landing page: the signed-in user's dashboard. Fetches the dashlets on the
 * server (forwarding the session cookie so visibility is enforced) and renders
 * them; falls back to a friendly message if the backend is unreachable.
 */
export default async function HomePage() {
  const cookie = (await cookies()).toString();

  let data: DashboardData | null = null;
  let error: string | null = null;
  try {
    data = await getDashboard(cookie);
  } catch {
    error = `Could not load your dashboard from the API at ${API_BASE_URL}. Is the backend running?`;
  }

  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10">
      <header className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight text-gray-900">
          Dashboard
        </h1>
        <p className="mt-1 text-sm text-gray-500">What needs your attention.</p>
      </header>

      {error || !data ? (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800">
          {error}
        </div>
      ) : (
        <Dashboard data={data} />
      )}
    </main>
  );
}
