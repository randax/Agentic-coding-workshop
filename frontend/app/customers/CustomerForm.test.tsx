import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import CustomerForm from "./CustomerForm";
import type { Customer } from "@/lib/api";
import { createCustomer, updateCustomer } from "@/lib/api";

const { push, refresh } = vi.hoisted(() => ({
  push: vi.fn(),
  refresh: vi.fn(),
}));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push, refresh }),
}));

vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return {
    ...actual,
    createCustomer: vi.fn(),
    updateCustomer: vi.fn(),
  };
});

const existing: Customer = {
  id: 7,
  name: "Grace Hopper",
  email: "grace@navy.example",
  phone: "555-0100",
  serviceAddress: "1 Navy Yard",
  accountNumber: "ACME-9",
  customerSince: "2020-01-02T00:00:00Z",
  status: "active",
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe("CustomerForm (create mode)", () => {
  it("submits the entered fields via createCustomer, then navigates to the new customer", async () => {
    vi.mocked(createCustomer).mockResolvedValue({ ...existing, id: 42 });
    render(<CustomerForm />);

    fireEvent.change(screen.getByLabelText("Name"), {
      target: { value: "Ada Lovelace" },
    });
    fireEvent.change(screen.getByLabelText("Email"), {
      target: { value: "ada@analytical.example" },
    });
    fireEvent.change(screen.getByLabelText(/account number/i), {
      target: { value: "ACME-100" },
    });
    fireEvent.change(screen.getByLabelText("Status"), {
      target: { value: "suspended" },
    });
    fireEvent.click(screen.getByRole("button", { name: /create/i }));

    await waitFor(() => {
      expect(createCustomer).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "Ada Lovelace",
          email: "ada@analytical.example",
          accountNumber: "ACME-100",
          status: "suspended",
        }),
      );
    });
    await waitFor(() => expect(push).toHaveBeenCalledWith("/customers/42"));
  });

  it("shows an error message when the save fails", async () => {
    vi.mocked(createCustomer).mockRejectedValue(new Error("nope"));
    render(<CustomerForm />);

    fireEvent.change(screen.getByLabelText("Name"), {
      target: { value: "Ada" },
    });
    fireEvent.change(screen.getByLabelText("Email"), {
      target: { value: "ada@x.example" },
    });
    fireEvent.change(screen.getByLabelText(/account number/i), {
      target: { value: "ACME-100" },
    });
    fireEvent.click(screen.getByRole("button", { name: /create/i }));

    expect(await screen.findByRole("alert")).toBeInTheDocument();
    expect(push).not.toHaveBeenCalled();
  });
});

describe("CustomerForm (edit mode)", () => {
  it("pre-fills the existing customer and submits via updateCustomer", async () => {
    vi.mocked(updateCustomer).mockResolvedValue({
      ...existing,
      name: "Grace B. Hopper",
    });
    render(<CustomerForm customer={existing} />);

    expect(screen.getByLabelText("Name")).toHaveValue("Grace Hopper");
    expect(screen.getByLabelText(/account number/i)).toHaveValue("ACME-9");

    fireEvent.change(screen.getByLabelText("Name"), {
      target: { value: "Grace B. Hopper" },
    });
    fireEvent.click(screen.getByRole("button", { name: /save/i }));

    await waitFor(() => {
      expect(updateCustomer).toHaveBeenCalledWith(
        7,
        expect.objectContaining({ name: "Grace B. Hopper" }),
      );
    });
    await waitFor(() => expect(push).toHaveBeenCalledWith("/customers/7"));
  });
});
