import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import CustomerSearch from "./CustomerSearch";

const h = vi.hoisted(() => ({
  replace: vi.fn(),
  params: new URLSearchParams(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: h.replace }),
  usePathname: () => "/",
  useSearchParams: () => h.params,
}));

beforeEach(() => {
  h.replace.mockClear();
  h.params = new URLSearchParams();
});

describe("CustomerSearch", () => {
  it("seeds the search box and status select from the current URL", () => {
    h.params = new URLSearchParams("search=ada&status=suspended");
    render(<CustomerSearch />);

    expect(screen.getByLabelText(/search/i)).toHaveValue("ada");
    expect(screen.getByLabelText(/status/i)).toHaveValue("suspended");
  });

  it("writes the typed search term to the URL query", () => {
    render(<CustomerSearch />);

    fireEvent.change(screen.getByLabelText(/search/i), {
      target: { value: "globex" },
    });

    expect(h.replace).toHaveBeenCalledWith("/?search=globex");
  });

  it("writes the selected status to the URL query", () => {
    render(<CustomerSearch />);

    fireEvent.change(screen.getByLabelText(/status/i), {
      target: { value: "active" },
    });

    expect(h.replace).toHaveBeenCalledWith("/?status=active");
  });

  it("removes a query param when its field is cleared", () => {
    h.params = new URLSearchParams("search=ada");
    render(<CustomerSearch />);

    fireEvent.change(screen.getByLabelText(/search/i), {
      target: { value: "" },
    });

    expect(h.replace).toHaveBeenCalledWith("/");
  });
});
