import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import StudioForm from "./StudioForm";
import { addCustomField } from "@/lib/api";

const { refresh } = vi.hoisted(() => ({ refresh: vi.fn() }));
vi.mock("next/navigation", () => ({ useRouter: () => ({ refresh }) }));
vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return { ...actual, addCustomField: vi.fn() };
});

beforeEach(() => vi.clearAllMocks());

describe("StudioForm", () => {
  it("adds a custom field with the chosen module, name, label and type", async () => {
    vi.mocked(addCustomField).mockResolvedValue({
      id: 1,
      module: "accounts",
      name: "churnRisk",
      label: "Churn risk",
      type: "string",
    });
    render(<StudioForm modules={["accounts", "contacts"]} />);

    fireEvent.change(screen.getByLabelText(/module/i), { target: { value: "accounts" } });
    fireEvent.change(screen.getByLabelText(/field name/i), { target: { value: "churnRisk" } });
    fireEvent.change(screen.getByLabelText(/label/i), { target: { value: "Churn risk" } });
    fireEvent.change(screen.getByLabelText(/type/i), { target: { value: "string" } });
    fireEvent.click(screen.getByRole("button", { name: /add field/i }));

    await waitFor(() =>
      expect(addCustomField).toHaveBeenCalledWith(
        expect.objectContaining({
          module: "accounts",
          name: "churnRisk",
          label: "Churn risk",
          type: "string",
        }),
      ),
    );
    expect(await screen.findByText(/added/i)).toBeInTheDocument();
  });
});
