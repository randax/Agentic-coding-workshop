import { render, screen, within, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import LayoutEditor from "./LayoutEditor";
import { getModuleMeta, getLayouts, saveLayouts } from "@/lib/api";

const { refresh } = vi.hoisted(() => ({ refresh: vi.fn() }));
vi.mock("next/navigation", () => ({ useRouter: () => ({ refresh }) }));
vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return {
    ...actual,
    getModuleMeta: vi.fn(),
    getLayouts: vi.fn(),
    saveLayouts: vi.fn(),
  };
});

const rawMeta = {
  module: "accounts",
  label: "Accounts",
  labelSingular: "Account",
  fields: [
    { name: "name", type: "string", label: "Name" },
    { name: "email", type: "string", label: "Email" },
    { name: "phone", type: "string", label: "Phone" },
    { name: "status", type: "enum", label: "Status", options: ["active", "suspended"] },
  ],
  listView: { columns: ["name", "email", "status"] },
  detailView: {
    panels: [
      { label: "Profile", fields: ["name", "email", "phone"] },
      { label: "Account", fields: ["status"] },
    ],
  },
  editView: { fields: ["name", "email", "phone", "status"] },
};

/** Renders the editor and resolves once the (async) load has populated the views. */
async function renderLoaded() {
  render(<LayoutEditor module="accounts" />);
  const list = within(await screen.findByRole("region", { name: /list columns/i }));
  return { list };
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getModuleMeta).mockResolvedValue(rawMeta as never);
  vi.mocked(getLayouts).mockResolvedValue({});
  vi.mocked(saveLayouts).mockResolvedValue(undefined);
});

describe("LayoutEditor", () => {
  it("renders each field as a visibility checkbox, defaulting hidden fields off", async () => {
    const { list } = await renderLoaded();

    // Default list columns are checked; a field absent from them (phone) shows unchecked.
    expect(list.getByLabelText("Name")).toBeChecked();
    expect(list.getByLabelText("Email")).toBeChecked();
    expect(list.getByLabelText("Phone")).not.toBeChecked();
  });

  it("omits an unchecked field from the saved layout", async () => {
    const { list } = await renderLoaded();
    fireEvent.click(list.getByLabelText("Email"));

    fireEvent.click(screen.getByRole("button", { name: /save layout/i }));

    await waitFor(() => expect(saveLayouts).toHaveBeenCalled());
    const [mod, views] = vi.mocked(saveLayouts).mock.calls[0];
    expect(mod).toBe("accounts");
    expect(views.list).toEqual(["name", "status"]);
  });

  it("reorders a column with the up button before saving", async () => {
    const { list } = await renderLoaded();

    fireEvent.click(list.getByRole("button", { name: /move status up/i }));
    fireEvent.click(screen.getByRole("button", { name: /save layout/i }));

    await waitFor(() => expect(saveLayouts).toHaveBeenCalled());
    expect(vi.mocked(saveLayouts).mock.calls[0][1].list).toEqual(["name", "status", "email"]);
  });

  it("saves detail and edit fields, detail flattened in code-panel order", async () => {
    await renderLoaded();

    fireEvent.click(screen.getByRole("button", { name: /save layout/i }));

    await waitFor(() => expect(saveLayouts).toHaveBeenCalled());
    const views = vi.mocked(saveLayouts).mock.calls[0][1];
    expect(views.detail).toEqual(["name", "email", "phone", "status"]);
    expect(views.edit).toEqual(["name", "email", "phone", "status"]);
  });
});
