import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import RecordActions from "./RecordActions";
import type { ActionMeta } from "@/lib/api";
import { runRecordAction } from "@/lib/api";

const { refresh } = vi.hoisted(() => ({ refresh: vi.fn() }));
vi.mock("next/navigation", () => ({ useRouter: () => ({ refresh }) }));
vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return { ...actual, runRecordAction: vi.fn() };
});

const convert: ActionMeta = { label: "Convert", method: "POST", path: "/leads/{id}/convert" };

beforeEach(() => vi.clearAllMocks());

describe("RecordActions", () => {
  it("renders a button per action and runs it for the record, then refreshes", async () => {
    vi.mocked(runRecordAction).mockResolvedValue();
    render(<RecordActions actions={[convert]} recordId={7} />);

    fireEvent.click(screen.getByRole("button", { name: /convert/i }));

    await waitFor(() =>
      expect(runRecordAction).toHaveBeenCalledWith(convert, 7),
    );
    await waitFor(() => expect(refresh).toHaveBeenCalled());
  });

  it("renders nothing when there are no actions", () => {
    const { container } = render(<RecordActions actions={[]} recordId={7} />);
    expect(container).toBeEmptyDOMElement();
  });
});
