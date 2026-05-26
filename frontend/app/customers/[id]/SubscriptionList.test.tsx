import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import SubscriptionList from "./SubscriptionList";
import type { Subscription, Product } from "@/lib/api";
import { createSubscription, cancelSubscription } from "@/lib/api";

const { refresh } = vi.hoisted(() => ({ refresh: vi.fn() }));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh }),
}));

vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return {
    ...actual,
    createSubscription: vi.fn(),
    cancelSubscription: vi.fn(),
  };
});

const fiber: Product = {
  id: 1,
  name: "Fiber 500",
  category: "fiber",
  monthlyPrice: 499,
  available: true,
  speedMbps: 500,
};

const meshRouter: Product = {
  id: 2,
  name: "Mesh Pro",
  category: "router",
  monthlyPrice: 99,
  available: true,
};

const active: Subscription = {
  id: 1,
  customerId: 7,
  productId: 1,
  status: "active",
  startDate: "2024-01-15T00:00:00Z",
  monthlyPriceSnapshot: 499,
  quantity: 1,
  product: fiber,
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe("SubscriptionList", () => {
  it("shows the product name, status, quantity and snapshot price", () => {
    render(
      <SubscriptionList
        customerId={7}
        subscriptions={[active]}
        availableProducts={[]}
      />,
    );

    expect(screen.getByText("Fiber 500")).toBeInTheDocument();
    expect(screen.getByText("active")).toBeInTheDocument();
    expect(screen.getByText("kr 499 / mo")).toBeInTheDocument();
    expect(screen.getByText("1")).toBeInTheDocument();
  });

  it("shows a message when there are no subscriptions", () => {
    render(
      <SubscriptionList
        customerId={7}
        subscriptions={[]}
        availableProducts={[fiber]}
      />,
    );

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
    render(
      <SubscriptionList
        customerId={7}
        subscriptions={[cancelled]}
        availableProducts={[]}
      />,
    );

    expect(screen.getByText(/Mar 2023/)).toBeInTheDocument();
    expect(screen.getByText(/Feb 2024/)).toBeInTheDocument();
  });

  it("lists the available products to subscribe to", () => {
    render(
      <SubscriptionList
        customerId={7}
        subscriptions={[]}
        availableProducts={[fiber, meshRouter]}
      />,
    );

    const select = screen.getByLabelText("Product") as HTMLSelectElement;
    const labels = Array.from(select.options).map((o) => o.textContent);
    expect(labels.some((l) => l?.includes("Fiber 500"))).toBe(true);
    expect(labels.some((l) => l?.includes("Mesh Pro"))).toBe(true);
  });

  it("assigns the selected product with the chosen quantity, then refreshes", async () => {
    vi.mocked(createSubscription).mockResolvedValue({
      ...active,
      id: 9,
      productId: 2,
      quantity: 3,
      product: meshRouter,
    });
    render(
      <SubscriptionList
        customerId={7}
        subscriptions={[]}
        availableProducts={[fiber, meshRouter]}
      />,
    );

    fireEvent.change(screen.getByLabelText("Product"), {
      target: { value: "2" },
    });
    fireEvent.change(screen.getByLabelText("Quantity"), {
      target: { value: "3" },
    });
    fireEvent.click(screen.getByRole("button", { name: /add/i }));

    await waitFor(() => {
      expect(createSubscription).toHaveBeenCalledWith(7, {
        productId: 2,
        quantity: 3,
      });
    });
    await waitFor(() => expect(refresh).toHaveBeenCalled());
  });

  it("cancels an active subscription, then refreshes", async () => {
    vi.mocked(cancelSubscription).mockResolvedValue({
      ...active,
      status: "cancelled",
      endDate: "2026-05-26T00:00:00Z",
    });
    render(
      <SubscriptionList
        customerId={7}
        subscriptions={[active]}
        availableProducts={[]}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: /cancel/i }));

    await waitFor(() => expect(cancelSubscription).toHaveBeenCalledWith(1));
    await waitFor(() => expect(refresh).toHaveBeenCalled());
  });

  it("does not offer to cancel an already-cancelled subscription", () => {
    const cancelled: Subscription = {
      ...active,
      id: 2,
      status: "cancelled",
      endDate: "2024-02-01T00:00:00Z",
    };
    render(
      <SubscriptionList
        customerId={7}
        subscriptions={[cancelled]}
        availableProducts={[]}
      />,
    );

    expect(
      screen.queryByRole("button", { name: /cancel/i }),
    ).not.toBeInTheDocument();
  });
});
