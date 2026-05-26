import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import CaseDetail from "./CaseDetail";
import type { Case } from "@/lib/api";

const sample: Case = {
  id: 5,
  customerId: 7,
  subject: "No internet since this morning",
  description: "Connection dropped around 08:00 and has not come back.",
  category: "connectivity",
  priority: "high",
  status: "in_progress",
  assignedAgent: { id: 1, name: "Sam Carter", email: "sam@isp.example" },
  comments: [
    {
      id: 1,
      caseId: 5,
      body: "Looking into it now.",
      authorAgent: { id: 1, name: "Sam Carter", email: "sam@isp.example" },
      createdAt: "2024-01-01T09:00:00Z",
    },
    {
      id: 2,
      caseId: 5,
      body: "Line check done, stable again.",
      authorAgent: { id: 2, name: "Robin Diaz", email: "robin@isp.example" },
      createdAt: "2024-01-02T09:00:00Z",
    },
  ],
  createdAt: "2024-01-01T08:00:00Z",
  updatedAt: "2024-01-02T09:00:00Z",
};

describe("CaseDetail", () => {
  it("shows the subject, description, status and assigned agent", () => {
    render(<CaseDetail caseItem={sample} />);

    expect(
      screen.getByRole("heading", { name: "No internet since this morning" }),
    ).toBeInTheDocument();
    expect(
      screen.getByText("Connection dropped around 08:00 and has not come back."),
    ).toBeInTheDocument();
    expect(screen.getByText(/in progress/i)).toBeInTheDocument();
    expect(screen.getAllByText(/Sam Carter/).length).toBeGreaterThan(0);
  });

  it("renders the comment timeline in order, with author and body", () => {
    render(<CaseDetail caseItem={sample} />);

    const items = screen.getAllByRole("listitem");
    expect(items).toHaveLength(2);
    expect(items[0]).toHaveTextContent("Looking into it now.");
    expect(items[0]).toHaveTextContent("Sam Carter");
    expect(items[1]).toHaveTextContent("Line check done, stable again.");
    expect(items[1]).toHaveTextContent("Robin Diaz");
  });

  it("shows an empty-timeline message when there are no comments", () => {
    render(<CaseDetail caseItem={{ ...sample, comments: [] }} />);

    expect(screen.getByText(/no comments yet/i)).toBeInTheDocument();
  });
});
