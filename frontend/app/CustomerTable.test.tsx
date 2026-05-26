import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import CustomerTable from "./CustomerTable";
import type { Customer } from "@/lib/api";

const ada: Customer = {
  id: 1,
  name: "Ada Lovelace",
  email: "ada@analytical.example",
  phone: "",
  serviceAddress: "",
  accountNumber: "ACME-001",
  customerSince: "2024-01-15T00:00:00Z",
  status: "active",
};

describe("CustomerTable", () => {
  it("renders a customer's name, account number, email and status", () => {
    render(<CustomerTable customers={[ada]} />);

    expect(screen.getByText("Ada Lovelace")).toBeInTheDocument();
    expect(screen.getByText("ACME-001")).toBeInTheDocument();
    expect(screen.getByText("ada@analytical.example")).toBeInTheDocument();
    expect(screen.getByText("active")).toBeInTheDocument();
  });

  it("links each customer name to their detail page", () => {
    render(<CustomerTable customers={[ada]} />);

    expect(screen.getByRole("link", { name: "Ada Lovelace" })).toHaveAttribute(
      "href",
      "/customers/1",
    );
  });

  it("shows an empty-state message when there are no customers", () => {
    render(<CustomerTable customers={[]} />);

    expect(screen.getByText(/no customers/i)).toBeInTheDocument();
  });
});
