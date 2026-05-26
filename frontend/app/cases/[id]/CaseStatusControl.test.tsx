import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import CaseStatusControl from "./CaseStatusControl";
import { updateCaseStatus } from "@/lib/api";

const { refresh } = vi.hoisted(() => ({ refresh: vi.fn() }));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh }),
}));

vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return {
    ...actual,
    updateCaseStatus: vi.fn(),
  };
});

beforeEach(() => {
  vi.clearAllMocks();
});

describe("CaseStatusControl", () => {
  it("offers only the legal next state from Open and advances it", async () => {
    vi.mocked(updateCaseStatus).mockResolvedValue({
      id: 5,
      customerId: 7,
      subject: "x",
      description: "",
      category: "general",
      priority: "low",
      status: "in_progress",
      createdAt: "2024-01-01T00:00:00Z",
      updatedAt: "2024-01-01T00:00:00Z",
    });
    render(<CaseStatusControl caseId={5} status="open" />);

    expect(
      screen.getByRole("button", { name: /in progress/i }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /resolved/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /closed/i }),
    ).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /in progress/i }));

    await waitFor(() =>
      expect(updateCaseStatus).toHaveBeenCalledWith(5, "in_progress"),
    );
    await waitFor(() => expect(refresh).toHaveBeenCalled());
  });

  it("offers reopen and close from Resolved", () => {
    render(<CaseStatusControl caseId={5} status="resolved" />);

    expect(
      screen.getByRole("button", { name: /in progress/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /closed/i }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /resolved/i }),
    ).not.toBeInTheDocument();
  });

  it("offers no transitions from Closed (terminal)", () => {
    render(<CaseStatusControl caseId={5} status="closed" />);

    expect(screen.queryByRole("button")).not.toBeInTheDocument();
    expect(screen.getByText(/closed/i)).toBeInTheDocument();
  });
});
