import { render, screen, within } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import RecordView from "./RecordView";
import type { ModuleMeta, ModuleRecord, SubpanelMeta } from "@/lib/api";

// RecordView renders RecordActions, which uses the router.
vi.mock("next/navigation", () => ({ useRouter: () => ({ refresh: vi.fn() }) }));

const meta: ModuleMeta = {
  module: "accounts",
  label: "Accounts",
  labelSingular: "Account",
  fields: [
    { name: "name", type: "string", label: "Name" },
    { name: "accountNumber", type: "string", label: "Account number" },
    { name: "status", type: "enum", label: "Status", options: ["active", "suspended"] },
  ],
  listView: { columns: ["name"] },
  detailView: {
    panels: [
      { label: "Profile", fields: ["name"] },
      { label: "Account", fields: ["accountNumber", "status"] },
    ],
  },
};

const record: ModuleRecord = {
  id: 7,
  name: "Ada Lovelace",
  accountNumber: "ACME-001",
  status: "suspended",
};

const casesSubpanel: SubpanelMeta = {
  label: "Cases",
  path: "/customers/{id}/cases",
  columns: [
    { name: "subject", type: "string", label: "Subject" },
    { name: "status", type: "string", label: "Status" },
  ],
};

describe("generic RecordView", () => {
  it("renders detailView panels with field labels and formatted values", () => {
    render(<RecordView meta={meta} record={record} subpanels={[]} />);

    expect(screen.getByText("Profile")).toBeInTheDocument();
    expect(screen.getByText("Account number")).toBeInTheDocument();
    expect(screen.getByText("ACME-001")).toBeInTheDocument();
    expect(screen.getByText("Suspended")).toBeInTheDocument(); // enum capitalized
  });

  it("renders a subpanel as a labelled table of related records", () => {
    render(
      <RecordView
        meta={meta}
        record={record}
        subpanels={[
          { meta: casesSubpanel, records: [{ id: 1, subject: "No internet", status: "open" }] },
        ]}
      />,
    );

    expect(screen.getByText("Cases")).toBeInTheDocument();
    const row = screen.getByText("No internet").closest("tr")!;
    expect(within(row).getAllByRole("cell")[0]).toHaveTextContent("No internet");
  });

  it("links to the edit view", () => {
    render(<RecordView meta={meta} record={record} subpanels={[]} />);
    expect(screen.getByRole("link", { name: /edit/i })).toHaveAttribute(
      "href",
      "/m/accounts/7/edit",
    );
  });
});
