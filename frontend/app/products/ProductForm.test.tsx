import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import ProductForm from "./ProductForm";
import type { Product } from "@/lib/api";
import { createProduct, updateProduct } from "@/lib/api";

const { refresh } = vi.hoisted(() => ({ refresh: vi.fn() }));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh }),
}));

vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return {
    ...actual,
    createProduct: vi.fn(),
    updateProduct: vi.fn(),
  };
});

beforeEach(() => {
  vi.clearAllMocks();
});

describe("ProductForm (create mode)", () => {
  it("submits the entered fields via createProduct, then refreshes", async () => {
    vi.mocked(createProduct).mockResolvedValue({
      id: 42,
      name: "Fiber 1000",
      category: "fiber",
      monthlyPrice: 699,
      available: true,
      speedMbps: 1000,
    });
    render(<ProductForm />);

    fireEvent.change(screen.getByLabelText("Name"), {
      target: { value: "Fiber 1000" },
    });
    fireEvent.change(screen.getByLabelText("Monthly price"), {
      target: { value: "699" },
    });
    fireEvent.change(screen.getByLabelText(/speed/i), {
      target: { value: "1000" },
    });
    fireEvent.click(screen.getByRole("button", { name: /create product/i }));

    await waitFor(() => {
      expect(createProduct).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "Fiber 1000",
          category: "fiber",
          monthlyPrice: 699,
          speedMbps: 1000,
        }),
      );
    });
    await waitFor(() => expect(refresh).toHaveBeenCalled());
  });

  it("swaps the per-category attribute field and submits it when category changes", async () => {
    vi.mocked(createProduct).mockResolvedValue({
      id: 5,
      name: "Mesh Pro",
      category: "router",
      monthlyPrice: 99,
      available: true,
      routerModel: "MeshPro X6",
    });
    render(<ProductForm />);

    // Fiber is the default, so the speed field is shown.
    expect(screen.getByLabelText(/speed/i)).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("Category"), {
      target: { value: "router" },
    });

    // Now the router-model field replaces the speed field.
    expect(screen.queryByLabelText(/speed/i)).not.toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Name"), {
      target: { value: "Mesh Pro" },
    });
    fireEvent.change(screen.getByLabelText("Monthly price"), {
      target: { value: "99" },
    });
    fireEvent.change(screen.getByLabelText(/router model/i), {
      target: { value: "MeshPro X6" },
    });
    fireEvent.click(screen.getByRole("button", { name: /create product/i }));

    await waitFor(() => {
      expect(createProduct).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "Mesh Pro",
          category: "router",
          monthlyPrice: 99,
          routerModel: "MeshPro X6",
        }),
      );
    });
    expect(vi.mocked(createProduct).mock.calls[0][0]).not.toHaveProperty(
      "speedMbps",
    );
  });
});

describe("ProductForm (edit mode)", () => {
  const existing: Product = {
    id: 7,
    name: "Fiber 500",
    category: "fiber",
    monthlyPrice: 499,
    available: true,
    speedMbps: 500,
  };

  it("pre-fills the product and submits changes via updateProduct, then calls onSaved", async () => {
    vi.mocked(updateProduct).mockResolvedValue({
      ...existing,
      name: "Fiber 1000",
      monthlyPrice: 699,
      speedMbps: 1000,
    });
    const onSaved = vi.fn();
    render(<ProductForm product={existing} onSaved={onSaved} />);

    expect(screen.getByLabelText("Name")).toHaveValue("Fiber 500");
    expect(screen.getByLabelText("Monthly price")).toHaveValue(499);
    expect(screen.getByLabelText(/speed/i)).toHaveValue(500);

    fireEvent.change(screen.getByLabelText("Name"), {
      target: { value: "Fiber 1000" },
    });
    fireEvent.change(screen.getByLabelText("Monthly price"), {
      target: { value: "699" },
    });
    fireEvent.click(screen.getByRole("button", { name: /save changes/i }));

    await waitFor(() => {
      expect(updateProduct).toHaveBeenCalledWith(
        7,
        expect.objectContaining({ name: "Fiber 1000", monthlyPrice: 699 }),
      );
    });
    await waitFor(() => expect(onSaved).toHaveBeenCalled());
    expect(createProduct).not.toHaveBeenCalled();
  });
});
