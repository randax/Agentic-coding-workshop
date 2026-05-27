import { render, screen, within } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import SearchResults from "./SearchResults";
import type { SearchGroup } from "@/lib/api";

const groups: SearchGroup[] = [
  {
    module: "accounts",
    hits: [{ module: "accounts", id: 1, title: "Northwind Traders" }],
  },
  {
    module: "cases",
    hits: [{ module: "cases", id: 7, title: "Fiber outage downtown" }],
  },
];

const section = (name: RegExp) =>
  within(screen.getByRole("region", { name }));

describe("SearchResults", () => {
  it("renders a labeled section per module, each hit linking to its record", () => {
    render(<SearchResults query="north" groups={groups} />);

    const accountLink = section(/accounts/i).getByRole("link", {
      name: "Northwind Traders",
    });
    expect(accountLink).toHaveAttribute("href", "/m/accounts/1");

    // Cases link to the dedicated case route, not the generic module route.
    const caseLink = section(/cases/i).getByRole("link", {
      name: "Fiber outage downtown",
    });
    expect(caseLink).toHaveAttribute("href", "/cases/7");
  });

  it("shows an empty state when there are no matches", () => {
    render(<SearchResults query="zzz" groups={[]} />);
    expect(screen.getByText(/no results for/i)).toBeInTheDocument();
    expect(screen.getByText(/zzz/)).toBeInTheDocument();
  });
});
