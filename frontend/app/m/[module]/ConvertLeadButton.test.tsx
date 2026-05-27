import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import ConvertLeadButton from "./ConvertLeadButton";
import { convertLead } from "@/lib/api";

const { push } = vi.hoisted(() => ({ push: vi.fn() }));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push }),
}));

vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return {
    ...actual,
    convertLead: vi.fn(),
  };
});

beforeEach(() => {
  vi.clearAllMocks();
});

describe("ConvertLeadButton", () => {
  it("shows the Convert button for a qualified, unconverted lead", () => {
    render(<ConvertLeadButton leadId={7} status="qualified" />);
    expect(
      screen.getByRole("button", { name: /convert/i }),
    ).toBeInTheDocument();
  });

  it("shows nothing for a lead that is not qualified", () => {
    const { container } = render(
      <ConvertLeadButton leadId={7} status="working" />,
    );
    expect(screen.queryByRole("button", { name: /convert/i })).toBeNull();
    expect(container).toBeEmptyDOMElement();
  });

  it("shows a banner linking to the account (not the button) once converted", () => {
    render(
      <ConvertLeadButton
        leadId={7}
        status="converted"
        convertedAccountId={42}
      />,
    );
    expect(screen.queryByRole("button", { name: /convert/i })).toBeNull();
    const link = screen.getByRole("link", { name: /converted/i });
    expect(link).toHaveAttribute("href", "/m/accounts/42");
  });

  it("converts on click and redirects to the new account", async () => {
    vi.mocked(convertLead).mockResolvedValue({ accountId: 100, contactId: 200 });
    render(<ConvertLeadButton leadId={7} status="qualified" />);

    fireEvent.click(screen.getByRole("button", { name: /convert/i }));

    await waitFor(() => expect(convertLead).toHaveBeenCalledWith(7));
    await waitFor(() =>
      expect(push).toHaveBeenCalledWith("/m/accounts/100"),
    );
  });

  it("surfaces the backend's conflict message instead of a generic error", async () => {
    vi.mocked(convertLead).mockRejectedValue(new Error("lead already converted"));
    render(<ConvertLeadButton leadId={7} status="qualified" />);

    fireEvent.click(screen.getByRole("button", { name: /convert/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "lead already converted",
    );
    expect(push).not.toHaveBeenCalled();
  });
});
