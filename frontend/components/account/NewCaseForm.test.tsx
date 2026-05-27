import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import NewCaseForm from "./NewCaseForm";
import { createCase } from "@/lib/api";

const { refresh } = vi.hoisted(() => ({ refresh: vi.fn() }));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh }),
}));

vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return {
    ...actual,
    createCase: vi.fn(),
  };
});

beforeEach(() => {
  vi.clearAllMocks();
});

describe("NewCaseForm", () => {
  it("opens a case with the entered fields, then refreshes", async () => {
    vi.mocked(createCase).mockResolvedValue({
      id: 9,
      customerId: 7,
      subject: "No internet",
      description: "Down since 8am",
      category: "connectivity",
      priority: "high",
      status: "open",
      createdAt: "2024-01-01T00:00:00Z",
      updatedAt: "2024-01-01T00:00:00Z",
    });
    render(<NewCaseForm customerId={7} />);

    fireEvent.change(screen.getByLabelText("Subject"), {
      target: { value: "No internet" },
    });
    fireEvent.change(screen.getByLabelText("Description"), {
      target: { value: "Down since 8am" },
    });
    fireEvent.change(screen.getByLabelText("Category"), {
      target: { value: "connectivity" },
    });
    fireEvent.change(screen.getByLabelText("Priority"), {
      target: { value: "high" },
    });
    fireEvent.click(screen.getByRole("button", { name: /open case/i }));

    await waitFor(() => {
      expect(createCase).toHaveBeenCalledWith(7, {
        subject: "No internet",
        description: "Down since 8am",
        category: "connectivity",
        priority: "high",
      });
    });
    await waitFor(() => expect(refresh).toHaveBeenCalled());
  });

  it("shows an error when opening the case fails", async () => {
    vi.mocked(createCase).mockRejectedValue(new Error("nope"));
    render(<NewCaseForm customerId={7} />);

    fireEvent.change(screen.getByLabelText("Subject"), {
      target: { value: "Help" },
    });
    fireEvent.click(screen.getByRole("button", { name: /open case/i }));

    expect(await screen.findByRole("alert")).toBeInTheDocument();
    expect(refresh).not.toHaveBeenCalled();
  });
});
