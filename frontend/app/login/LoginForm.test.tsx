import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import LoginForm from "./LoginForm";
import { login, InvalidCredentialsError } from "@/lib/api";

const { push, refresh } = vi.hoisted(() => ({ push: vi.fn(), refresh: vi.fn() }));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push, refresh }),
}));
vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return { ...actual, login: vi.fn() };
});

beforeEach(() => vi.clearAllMocks());

describe("LoginForm", () => {
  it("logs in with the entered credentials and navigates home", async () => {
    vi.mocked(login).mockResolvedValue({
      id: 1,
      name: "Sam Carter",
      email: "sam@isp.example",
      role: "admin",
    });
    render(<LoginForm />);

    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: "sam@isp.example" },
    });
    fireEvent.change(screen.getByLabelText(/password/i), {
      target: { value: "s3cret" },
    });
    fireEvent.click(screen.getByRole("button", { name: /sign in/i }));

    await waitFor(() =>
      expect(login).toHaveBeenCalledWith("sam@isp.example", "s3cret"),
    );
    await waitFor(() => expect(push).toHaveBeenCalledWith("/"));
  });

  it("shows an error and does not navigate when credentials are rejected", async () => {
    vi.mocked(login).mockRejectedValue(new InvalidCredentialsError("nope"));
    render(<LoginForm />);

    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: "sam@isp.example" },
    });
    fireEvent.change(screen.getByLabelText(/password/i), {
      target: { value: "wrong" },
    });
    fireEvent.click(screen.getByRole("button", { name: /sign in/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/invalid/i);
    expect(push).not.toHaveBeenCalled();
  });
});
