import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import ProductCatalog from "./ProductCatalog";
import type { Product } from "@/lib/api";
import { retireProduct, unretireProduct } from "@/lib/api";

const { refresh } = vi.hoisted(() => ({ refresh: vi.fn() }));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh }),
}));

vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return {
    ...actual,
    retireProduct: vi.fn(),
    unretireProduct: vi.fn(),
    createProduct: vi.fn(),
    updateProduct: vi.fn(),
  };
});

beforeEach(() => {
  vi.clearAllMocks();
});

const fiber: Product = {
  id: 1,
  name: "Fiber 500",
  category: "fiber",
  monthlyPrice: 499,
  available: true,
  speedMbps: 500,
};

describe("ProductCatalog", () => {
  it("shows a fiber product's name and speed", () => {
    render(<ProductCatalog products={[fiber]} />);

    expect(screen.getByText("Fiber 500")).toBeInTheDocument();
    expect(screen.getByText(/500\s*Mbps/)).toBeInTheDocument();
  });

  it("shows the router model and TV package tier as the category detail", () => {
    const router: Product = {
      id: 2,
      name: "Mesh Router Pro",
      category: "router",
      monthlyPrice: 99,
      available: true,
      routerModel: "MeshPro X6",
    };
    const tv: Product = {
      id: 3,
      name: "TV Premium",
      category: "tv",
      monthlyPrice: 399,
      available: true,
      tvPackageTier: "Premium",
    };
    render(<ProductCatalog products={[router, tv]} />);

    expect(screen.getByText("MeshPro X6")).toBeInTheDocument();
    expect(screen.getByText("Premium")).toBeInTheDocument();
  });

  it("shows the monthly price", () => {
    render(<ProductCatalog products={[fiber]} />);

    expect(screen.getByText("kr 499 / mo")).toBeInTheDocument();
  });

  it("flags a retired product as unavailable", () => {
    const retired: Product = { ...fiber, id: 9, available: false };
    render(<ProductCatalog products={[retired]} />);

    expect(screen.getByText("Retired")).toBeInTheDocument();
  });

  it("retires an available product via retireProduct, then refreshes", async () => {
    vi.mocked(retireProduct).mockResolvedValue();
    render(<ProductCatalog products={[fiber]} />);

    fireEvent.click(screen.getByRole("button", { name: /^retire$/i }));

    await waitFor(() => expect(retireProduct).toHaveBeenCalledWith(1));
    await waitFor(() => expect(refresh).toHaveBeenCalled());
    expect(unretireProduct).not.toHaveBeenCalled();
  });

  it("reactivates a retired product via unretireProduct, then refreshes", async () => {
    vi.mocked(unretireProduct).mockResolvedValue();
    const retired: Product = { ...fiber, id: 9, available: false };
    render(<ProductCatalog products={[retired]} />);

    // A retired row offers reactivation, not retirement.
    expect(
      screen.queryByRole("button", { name: /^retire$/i }),
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /unretire/i }));

    await waitFor(() => expect(unretireProduct).toHaveBeenCalledWith(9));
    await waitFor(() => expect(refresh).toHaveBeenCalled());
  });

  it("reveals an inline edit form pre-filled with the product when Edit is clicked", () => {
    render(<ProductCatalog products={[fiber]} />);

    // No edit form is shown until Edit is clicked.
    expect(
      screen.queryByRole("button", { name: /save changes/i }),
    ).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /^edit$/i }));

    expect(
      screen.getByRole("button", { name: /save changes/i }),
    ).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toHaveValue("Fiber 500");
  });
});
