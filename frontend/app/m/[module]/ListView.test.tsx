import { render, screen, within } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import ListView from "./ListView";
import type { ModuleMeta, ModuleRecord } from "@/lib/api";

const productsMeta: ModuleMeta = {
  module: "products",
  label: "Products",
  labelSingular: "Product",
  fields: [
    { name: "name", type: "string", label: "Name" },
    {
      name: "category",
      type: "enum",
      label: "Category",
      options: ["fiber", "router", "tv"],
    },
    { name: "monthlyPrice", type: "currency", label: "Monthly price" },
    { name: "available", type: "bool", label: "Status" },
  ],
  listView: { columns: ["name", "category", "monthlyPrice", "available"] },
};

const records: ModuleRecord[] = [
  {
    id: 1,
    name: "Fiber 500",
    category: "fiber",
    monthlyPrice: 499,
    available: true,
  },
];

describe("generic ListView", () => {
  it("renders a column header per listView column, using field labels", () => {
    render(<ListView meta={productsMeta} records={[]} />);

    const headers = screen.getAllByRole("columnheader").map((h) => h.textContent);
    expect(headers).toEqual(["Name", "Category", "Monthly price", "Status"]);
  });

  it("renders each record formatted by field type", () => {
    render(<ListView meta={productsMeta} records={records} />);

    const row = screen.getByText("Fiber 500").closest("tr")!;
    const cells = within(row).getAllByRole("cell");
    expect(cells[0]).toHaveTextContent("Fiber 500"); // string
    expect(cells[1]).toHaveTextContent("Fiber"); // enum capitalized
    expect(cells[2]).toHaveTextContent("kr 499"); // currency
    expect(cells[3]).toHaveTextContent("Yes"); // bool
  });
});
