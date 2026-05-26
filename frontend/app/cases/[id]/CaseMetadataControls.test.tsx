import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import CaseMetadataControls from "./CaseMetadataControls";
import type { Agent } from "@/lib/api";
import { updateCaseMetadata } from "@/lib/api";

const { refresh } = vi.hoisted(() => ({ refresh: vi.fn() }));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh }),
}));

vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return {
    ...actual,
    updateCaseMetadata: vi.fn(),
  };
});

const agents: Agent[] = [
  { id: 1, name: "Sam Carter", email: "sam@isp.example" },
  { id: 2, name: "Robin Diaz", email: "robin@isp.example" },
];

const baseCase = {
  id: 5,
  customerId: 7,
  subject: "x",
  description: "",
  category: "general" as const,
  priority: "low" as const,
  status: "open" as const,
  createdAt: "2024-01-01T00:00:00Z",
  updatedAt: "2024-01-01T00:00:00Z",
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(updateCaseMetadata).mockResolvedValue(baseCase);
});

describe("CaseMetadataControls", () => {
  function renderControls() {
    render(
      <CaseMetadataControls
        caseId={5}
        priority="low"
        category="general"
        assignedAgentId={1}
        agents={agents}
      />,
    );
  }

  it("changes priority via updateCaseMetadata, then refreshes", async () => {
    renderControls();

    fireEvent.change(screen.getByLabelText("Priority"), {
      target: { value: "urgent" },
    });

    await waitFor(() =>
      expect(updateCaseMetadata).toHaveBeenCalledWith(5, { priority: "urgent" }),
    );
    await waitFor(() => expect(refresh).toHaveBeenCalled());
  });

  it("changes category via updateCaseMetadata", async () => {
    renderControls();

    fireEvent.change(screen.getByLabelText("Category"), {
      target: { value: "billing" },
    });

    await waitFor(() =>
      expect(updateCaseMetadata).toHaveBeenCalledWith(5, { category: "billing" }),
    );
  });

  it("reassigns the case to a chosen agent via updateCaseMetadata", async () => {
    renderControls();

    fireEvent.change(screen.getByLabelText("Assigned to"), {
      target: { value: "2" },
    });

    await waitFor(() =>
      expect(updateCaseMetadata).toHaveBeenCalledWith(5, { assignedAgentId: 2 }),
    );
    await waitFor(() => expect(refresh).toHaveBeenCalled());
  });
});
