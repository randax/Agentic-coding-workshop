import { render, screen, within } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import Dashboard from "./Dashboard";
import type { DashboardData } from "@/lib/api";

const data: DashboardData = {
  myOpenCases: [
    {
      id: 1, customerId: 1, subject: "No internet", description: "",
      category: "connectivity", priority: "high", status: "open",
      createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z",
    },
  ],
  myTasks: [{ id: 2, subject: "Call Ada back", status: "open", occurredAt: "2026-01-02T00:00:00Z" }],
  recentLeads: [{ id: 3, name: "Priya Patel", company: "Fjord Logistics", status: "new" }],
  pipelineByStage: [
    { stage: "prospecting", count: 2, totalAmount: 300, items: [] },
    { stage: "closed_won", count: 0, totalAmount: 0, items: [] },
  ],
};

const region = (name: RegExp) => within(screen.getByRole("region", { name }));

describe("Dashboard", () => {
  it("renders each dashlet with its items", () => {
    render(<Dashboard data={data} />);

    expect(region(/my open cases/i).getByText("No internet")).toBeInTheDocument();
    expect(region(/my tasks/i).getByText("Call Ada back")).toBeInTheDocument();
    expect(region(/recent leads/i).getByText(/Priya Patel/)).toBeInTheDocument();

    const pipeline = region(/pipeline by stage/i);
    expect(pipeline.getByText(/prospecting/i)).toBeInTheDocument();
    expect(pipeline.getByText("2")).toBeInTheDocument(); // count for prospecting
  });

  it("shows an empty state for a dashlet with no items", () => {
    render(
      <Dashboard
        data={{ myOpenCases: [], myTasks: [], recentLeads: [], pipelineByStage: [] }}
      />,
    );

    expect(region(/my open cases/i).getByText(/no open cases/i)).toBeInTheDocument();
    expect(region(/my tasks/i).getByText(/no open tasks/i)).toBeInTheDocument();
  });
});
