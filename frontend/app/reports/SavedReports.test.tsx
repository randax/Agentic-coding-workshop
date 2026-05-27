import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import SavedReports from "./SavedReports";
import { runSavedReport, type SavedReport } from "@/lib/api";

vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return { ...actual, runSavedReport: vi.fn() };
});

beforeEach(() => vi.clearAllMocks());

const reports: SavedReport[] = [
  { id: 7, name: "Leads by status", definition: { module: "leads", groupBy: "status", aggregation: "count" } },
];

describe("SavedReports", () => {
  it("re-runs a saved report and charts the result", async () => {
    vi.mocked(runSavedReport).mockResolvedValue({ rows: [{ group: "new", count: 2, value: 2 }] });
    render(<SavedReports reports={reports} />);

    expect(screen.getByText("Leads by status")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /run/i }));

    await waitFor(() => expect(runSavedReport).toHaveBeenCalledWith(7));
    expect(await screen.findByTestId("report-bar-new")).toBeInTheDocument();
  });

  it("shows an empty state when there are no saved reports", () => {
    render(<SavedReports reports={[]} />);
    expect(screen.getByText(/no saved reports/i)).toBeInTheDocument();
  });
});
