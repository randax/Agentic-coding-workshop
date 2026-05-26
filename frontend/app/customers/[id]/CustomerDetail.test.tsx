import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import CustomerDetail from "./CustomerDetail";
import type { Customer } from "@/lib/api";

const sample: Customer = {
  id: 3,
  name: "Grace Hopper",
  email: "grace@example.com",
  phone: "+47 900 00 003",
  serviceAddress: "Havnegata 12, 7010 Trondheim",
  accountNumber: "ISP-1003",
  customerSince: "2020-12-09T00:00:00Z",
  status: "suspended",
};

describe("CustomerDetail", () => {
  it("shows the customer's profile fields on the profile tab", () => {
    render(<CustomerDetail customer={sample} activeTab="profile" />);

    expect(
      screen.getByRole("heading", { name: "Grace Hopper" }),
    ).toBeInTheDocument();
    expect(screen.getByText("grace@example.com")).toBeInTheDocument();
    expect(screen.getByText("+47 900 00 003")).toBeInTheDocument();
    expect(screen.getByText("Havnegata 12, 7010 Trondheim")).toBeInTheDocument();
    expect(screen.getByText("ISP-1003")).toBeInTheDocument();
  });

  it("shows the status and customer-since date on the profile tab", () => {
    render(<CustomerDetail customer={sample} activeTab="profile" />);

    expect(screen.getByText("suspended")).toBeInTheDocument();
    // 2020-12-09 formatted en-GB short month
    expect(screen.getByText(/Dec 2020/)).toBeInTheDocument();
  });

  it("renders Profile, Subscriptions and Cases tabs linking to this customer", () => {
    render(<CustomerDetail customer={sample} activeTab="cases" />);

    expect(screen.getByRole("link", { name: "Profile" })).toHaveAttribute(
      "href",
      "/customers/3?tab=profile",
    );
    expect(screen.getByRole("link", { name: "Subscriptions" })).toHaveAttribute(
      "href",
      "/customers/3?tab=subscriptions",
    );
    expect(screen.getByRole("link", { name: "Cases" })).toHaveAttribute(
      "href",
      "/customers/3?tab=cases",
    );
  });

  it("marks the active tab with aria-current", () => {
    render(<CustomerDetail customer={sample} activeTab="cases" />);

    expect(screen.getByRole("link", { name: "Cases" })).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(screen.getByRole("link", { name: "Profile" })).not.toHaveAttribute(
      "aria-current",
    );
  });

  it("shows a placeholder on the subscriptions tab instead of profile fields", () => {
    render(<CustomerDetail customer={sample} activeTab="subscriptions" />);

    expect(screen.queryByText("grace@example.com")).not.toBeInTheDocument();
    expect(screen.getByText("No subscriptions to display yet.")).toBeInTheDocument();
  });

  it("shows a placeholder on the cases tab instead of profile fields", () => {
    render(<CustomerDetail customer={sample} activeTab="cases" />);

    expect(screen.queryByText("grace@example.com")).not.toBeInTheDocument();
    expect(screen.getByText("No support cases to display yet.")).toBeInTheDocument();
  });
});
