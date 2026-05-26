import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import CaseList from "./CaseList";
import type { Case } from "@/lib/api";

const sampleCase: Case = {
  id: 1,
  customerId: 7,
  subject: "No internet since this morning",
  description: "Connection dropped around 08:00.",
  category: "connectivity",
  priority: "high",
  status: "in_progress",
  assignedAgent: { id: 1, name: "Sam Carter", email: "sam@isp.example" },
  createdAt: "2024-01-01T00:00:00Z",
  updatedAt: "2024-01-02T00:00:00Z",
};

describe("CaseList", () => {
  it("lists a case's subject, category, priority, status and assigned agent", () => {
    render(<CaseList cases={[sampleCase]} />);

    expect(
      screen.getByText("No internet since this morning"),
    ).toBeInTheDocument();
    expect(screen.getByText(/connectivity/i)).toBeInTheDocument();
    expect(screen.getByText(/high/i)).toBeInTheDocument();
    expect(screen.getByText(/in progress/i)).toBeInTheDocument();
    expect(screen.getByText("Sam Carter")).toBeInTheDocument();
  });

  it("links each case to its detail page", () => {
    render(<CaseList cases={[sampleCase]} />);

    expect(
      screen.getByRole("link", { name: "No internet since this morning" }),
    ).toHaveAttribute("href", "/cases/1");
  });

  it("shows an unassigned label when no agent is assigned", () => {
    render(<CaseList cases={[{ ...sampleCase, assignedAgent: null }]} />);

    expect(screen.getByText(/unassigned/i)).toBeInTheDocument();
  });

  it("shows an empty state when there are no cases", () => {
    render(<CaseList cases={[]} />);

    expect(screen.getByText(/no support cases/i)).toBeInTheDocument();
  });
});
