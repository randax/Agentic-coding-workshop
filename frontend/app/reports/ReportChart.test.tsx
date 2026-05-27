import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import ReportChart from "./ReportChart";
import type { ReportResult } from "@/lib/api";

const result: ReportResult = {
  rows: [
    { group: "new", count: 2, value: 2 },
    { group: "working", count: 1, value: 1 },
  ],
};

describe("ReportChart", () => {
  it("renders a labeled bar per row showing its value", () => {
    render(<ReportChart result={result} />);

    const newBar = screen.getByTestId("report-bar-new");
    expect(newBar).toHaveTextContent(/new/i);
    expect(newBar).toHaveTextContent(/2/);

    expect(screen.getByTestId("report-bar-working")).toHaveTextContent(/working/i);
  });

  it("shows an empty state when there are no rows", () => {
    render(<ReportChart result={{ rows: [] }} />);
    expect(screen.getByText(/no data/i)).toBeInTheDocument();
  });
});
