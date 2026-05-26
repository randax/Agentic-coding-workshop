import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import SubscriptionList from "./SubscriptionList";
import type { Subscription } from "@/lib/api";

const active: Subscription = {
  id: 1,
  customerId: 7,
  productId: 1,
  status: "active",
  startDate: "2024-01-15T00:00:00Z",
  monthlyPriceSnapshot: 499,
  quantity: 1,
  product: {
    id: 1,
    name: "Fiber 500",
    category: "fiber",
    monthlyPrice: 499,
    available: true,
    speedMbps: 500,
  },
};

describe("SubscriptionList", () => {
  it("shows the product name, status, quantity and snapshot price", () => {
    render(<SubscriptionList subscriptions={[active]} />);

    expect(screen.getByText("Fiber 500")).toBeInTheDocument();
    expect(screen.getByText("active")).toBeInTheDocument();
    expect(screen.getByText("kr 499 / mo")).toBeInTheDocument();
    expect(screen.getByText("1")).toBeInTheDocument();
  });

  it("shows a message when there are no subscriptions", () => {
    render(<SubscriptionList subscriptions={[]} />);

    expect(
      screen.getByText("This customer has no subscriptions."),
    ).toBeInTheDocument();
  });

  it("shows the start date and, for cancelled subscriptions, the end date", () => {
    const cancelled: Subscription = {
      ...active,
      id: 2,
      status: "cancelled",
      startDate: "2023-03-01T00:00:00Z",
      endDate: "2024-02-01T00:00:00Z",
    };
    render(<SubscriptionList subscriptions={[cancelled]} />);

    expect(screen.getByText(/Mar 2023/)).toBeInTheDocument();
    expect(screen.getByText(/Feb 2024/)).toBeInTheDocument();
  });
});
