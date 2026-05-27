import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import EditView from "./EditView";
import type { ModuleMeta, ModuleRecord } from "@/lib/api";
import { updateModuleRecord } from "@/lib/api";

const { push, refresh } = vi.hoisted(() => ({ push: vi.fn(), refresh: vi.fn() }));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push, refresh }),
}));
vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return { ...actual, updateModuleRecord: vi.fn() };
});

const meta: ModuleMeta = {
  module: "accounts",
  label: "Accounts",
  labelSingular: "Account",
  fields: [
    { name: "name", type: "string", label: "Name" },
    { name: "status", type: "enum", label: "Status", options: ["active", "suspended"] },
  ],
  listView: { columns: ["name"] },
  editView: { fields: ["name", "status"] },
};

const record: ModuleRecord = { id: 7, name: "Ada Lovelace", status: "active" };

beforeEach(() => vi.clearAllMocks());

describe("generic EditView", () => {
  it("pre-fills fields from the record and renders an input per editView field", () => {
    render(<EditView meta={meta} record={record} />);
    expect(screen.getByLabelText("Name")).toHaveValue("Ada Lovelace");
    expect(screen.getByLabelText("Status")).toHaveValue("active");
  });

  it("submits edited values via updateModuleRecord, then navigates to the record", async () => {
    vi.mocked(updateModuleRecord).mockResolvedValue({ ...record, name: "Ada B. Lovelace" });
    render(<EditView meta={meta} record={record} />);

    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Ada B. Lovelace" } });
    fireEvent.change(screen.getByLabelText("Status"), { target: { value: "suspended" } });
    fireEvent.click(screen.getByRole("button", { name: /save/i }));

    await waitFor(() => {
      expect(updateModuleRecord).toHaveBeenCalledWith(
        "accounts",
        7,
        expect.objectContaining({ name: "Ada B. Lovelace", status: "suspended" }),
      );
    });
    await waitFor(() => expect(push).toHaveBeenCalledWith("/m/accounts/7"));
  });
});
