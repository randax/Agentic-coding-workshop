import { render, screen, within, fireEvent } from "@testing-library/react";
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

const many: ModuleRecord[] = [
  { id: 1, name: "Fiber 500", category: "fiber", monthlyPrice: 499, available: true },
  { id: 2, name: "Mesh Pro", category: "router", monthlyPrice: 99, available: true },
  { id: 3, name: "TV Max", category: "tv", monthlyPrice: 299, available: false },
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

  it("links each row to its record view", () => {
    render(<ListView meta={productsMeta} records={records} />);
    expect(screen.getByRole("link", { name: "Fiber 500" })).toHaveAttribute(
      "href",
      "/m/products/1",
    );
  });

  it("filters rows by a free-text query across columns", () => {
    render(<ListView meta={productsMeta} records={many} />);
    expect(screen.getAllByRole("row")).toHaveLength(4); // header + 3

    fireEvent.change(screen.getByLabelText(/filter/i), { target: { value: "mesh" } });

    expect(screen.getByText("Mesh Pro")).toBeInTheDocument();
    expect(screen.queryByText("Fiber 500")).not.toBeInTheDocument();
  });

  it("shows a 'New <singular>' action linking to the create page when creation is enabled", () => {
    render(<ListView meta={productsMeta} records={[]} newHref="/m/products/new" />);
    expect(screen.getByRole("link", { name: /new product/i })).toHaveAttribute(
      "href",
      "/m/products/new",
    );
  });

  it("omits the create action when no newHref is given", () => {
    render(<ListView meta={productsMeta} records={[]} />);
    expect(screen.queryByRole("link", { name: /new product/i })).not.toBeInTheDocument();
  });

  it("sorts by a column when its header is clicked", () => {
    render(<ListView meta={productsMeta} records={many} />);

    fireEvent.click(screen.getByRole("button", { name: /name/i }));
    const firstDataRow = screen.getAllByRole("row")[1];
    expect(firstDataRow).toHaveTextContent("Fiber 500"); // asc: Fiber, Mesh, TV

    fireEvent.click(screen.getByRole("button", { name: /name/i }));
    const firstDataRowDesc = screen.getAllByRole("row")[1];
    expect(firstDataRowDesc).toHaveTextContent("TV Max"); // desc
  });
});
