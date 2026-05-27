import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import LayoutStudio from "./LayoutStudio";

// The editor itself is tested in LayoutEditor.test.tsx; here we only verify that
// LayoutStudio drives it with the picked module.
vi.mock("./LayoutEditor", () => ({
  default: ({ module }: { module: string }) => <div data-testid="editor">editing {module}</div>,
}));

describe("LayoutStudio", () => {
  it("edits the first module by default and switches when another is picked", () => {
    render(<LayoutStudio modules={["accounts", "contacts"]} />);
    expect(screen.getByTestId("editor")).toHaveTextContent("editing accounts");

    fireEvent.change(screen.getByLabelText(/layout module/i), { target: { value: "contacts" } });

    expect(screen.getByTestId("editor")).toHaveTextContent("editing contacts");
  });
});
