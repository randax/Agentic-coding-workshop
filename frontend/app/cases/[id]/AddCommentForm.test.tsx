import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import AddCommentForm from "./AddCommentForm";
import type { Agent } from "@/lib/api";
import { addCaseComment } from "@/lib/api";

const { refresh } = vi.hoisted(() => ({ refresh: vi.fn() }));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh }),
}));

vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return {
    ...actual,
    addCaseComment: vi.fn(),
  };
});

const agents: Agent[] = [
  { id: 1, name: "Sam Carter", email: "sam@isp.example" },
  { id: 2, name: "Robin Diaz", email: "robin@isp.example" },
];

beforeEach(() => {
  vi.clearAllMocks();
});

describe("AddCommentForm", () => {
  it("adds a comment attributed to the chosen agent, then refreshes", async () => {
    vi.mocked(addCaseComment).mockResolvedValue({
      id: 9,
      caseId: 5,
      body: "On it",
      authorAgentId: 2,
      createdAt: "2024-01-01T00:00:00Z",
    });
    render(<AddCommentForm caseId={5} agents={agents} />);

    fireEvent.change(screen.getByLabelText("Comment"), {
      target: { value: "On it" },
    });
    fireEvent.change(screen.getByLabelText("Author"), {
      target: { value: "2" },
    });
    fireEvent.click(screen.getByRole("button", { name: /add comment/i }));

    await waitFor(() => {
      expect(addCaseComment).toHaveBeenCalledWith(5, {
        body: "On it",
        authorAgentId: 2,
      });
    });
    await waitFor(() => expect(refresh).toHaveBeenCalled());
  });

  it("shows an error when adding the comment fails", async () => {
    vi.mocked(addCaseComment).mockRejectedValue(new Error("nope"));
    render(<AddCommentForm caseId={5} agents={agents} />);

    fireEvent.change(screen.getByLabelText("Comment"), {
      target: { value: "On it" },
    });
    fireEvent.click(screen.getByRole("button", { name: /add comment/i }));

    expect(await screen.findByRole("alert")).toBeInTheDocument();
    expect(refresh).not.toHaveBeenCalled();
  });
});
