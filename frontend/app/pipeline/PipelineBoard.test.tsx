import { render, screen, within } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import PipelineBoard from "./PipelineBoard";
import type { PipelineStage } from "@/lib/api";

const stages: PipelineStage[] = [
  {
    stage: "prospecting",
    count: 2,
    totalAmount: 1500,
    items: [
      { id: 1, name: "Globex fiber", amount: 1000 },
      { id: 2, name: "Acme fiber", amount: 500 },
    ],
  },
  { stage: "closed_won", count: 1, totalAmount: 2000, items: [{ id: 3, name: "Won deal", amount: 2000 }] },
];

describe("PipelineBoard", () => {
  it("renders a column per stage with a readable label and rolled-up total", () => {
    render(<PipelineBoard stages={stages} />);

    const prospecting = screen.getByTestId("stage-prospecting");
    expect(prospecting).toHaveTextContent(/Prospecting/);
    expect(prospecting).toHaveTextContent(/kr 1500/);
    expect(within(prospecting).getByText("Globex fiber")).toBeInTheDocument();

    const won = screen.getByTestId("stage-closed_won");
    expect(won).toHaveTextContent(/Closed won/);
  });
});
