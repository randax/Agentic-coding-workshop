import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import ProductCatalog from "./ProductCatalog";
import type { Product } from "@/lib/api";

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
});
