import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import StatusBadge from "./StatusBadge";

describe("StatusBadge", () => {
  it("renders the status label", () => {
    render(<StatusBadge status="active" />);
    expect(screen.getByText("active")).toBeInTheDocument();
  });

  it("uses distinct styling per status", () => {
    const { rerender } = render(<StatusBadge status="active" />);
    const activeClass = screen.getByText("active").className;

    rerender(<StatusBadge status="suspended" />);
    const suspendedClass = screen.getByText("suspended").className;

    expect(activeClass).not.toEqual(suspendedClass);
  });
});
