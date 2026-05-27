import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import ReportBuilder, { type ReportModule } from "./ReportBuilder";
import { runReport, saveReport } from "@/lib/api";

vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return { ...actual, runReport: vi.fn(), saveReport: vi.fn() };
});

beforeEach(() => vi.clearAllMocks());

const modules: ReportModule[] = [
  {
    name: "leads",
    label: "Leads",
    fields: [
      { name: "status", type: "enum", label: "Status" },
      { name: "company", type: "string", label: "Company" },
    ],
  },
  {
    name: "opportunities",
    label: "Opportunities",
    fields: [
      { name: "stage", type: "enum", label: "Stage" },
      { name: "amount", type: "currency", label: "Amount" },
    ],
  },
];

describe("ReportBuilder", () => {
  it("runs the built report and renders its chart", async () => {
    vi.mocked(runReport).mockResolvedValue({ rows: [{ group: "new", count: 3, value: 3 }] });
    render(<ReportBuilder modules={modules} />);

    fireEvent.click(screen.getByRole("button", { name: /run report/i }));

    await waitFor(() =>
      expect(runReport).toHaveBeenCalledWith(
        expect.objectContaining({ module: "leads", groupBy: "status", aggregation: "count" }),
      ),
    );
    expect(await screen.findByTestId("report-bar-new")).toBeInTheDocument();
  });

  it("includes an added filter in the run definition", async () => {
    vi.mocked(runReport).mockResolvedValue({ rows: [] });
    render(<ReportBuilder modules={modules} />);

    fireEvent.click(screen.getByRole("button", { name: /add filter/i }));
    fireEvent.change(screen.getByLabelText(/filter field/i), { target: { value: "company" } });
    fireEvent.change(screen.getByLabelText(/filter operator/i), { target: { value: "contains" } });
    fireEvent.change(screen.getByLabelText(/filter value/i), { target: { value: "acme" } });
    fireEvent.click(screen.getByRole("button", { name: /run report/i }));

    await waitFor(() =>
      expect(runReport).toHaveBeenCalledWith(
        expect.objectContaining({
          filters: [{ field: "company", operator: "contains", value: "acme" }],
        }),
      ),
    );
  });

  it("saves the current report under a name", async () => {
    vi.mocked(saveReport).mockResolvedValue({
      id: 1,
      name: "My report",
      definition: { module: "leads", groupBy: "status", aggregation: "count" },
    });
    render(<ReportBuilder modules={modules} />);

    fireEvent.change(screen.getByLabelText(/report name/i), { target: { value: "My report" } });
    fireEvent.click(screen.getByRole("button", { name: /save report/i }));

    await waitFor(() =>
      expect(saveReport).toHaveBeenCalledWith(
        "My report",
        expect.objectContaining({ module: "leads", groupBy: "status", aggregation: "count" }),
      ),
    );
  });
});
