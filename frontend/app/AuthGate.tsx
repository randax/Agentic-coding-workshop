"use client";

import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import Link from "next/link";
import { getCurrentUser, logout, type AuthUser } from "@/lib/api";

/**
 * Client-side auth guard + app chrome. Resolves the current user from the
 * session cookie; unauthenticated visitors are redirected to /login (which
 * renders without the nav). Authenticated pages get the nav, showing the
 * signed-in user and a logout control.
 */
export default function AuthGate({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);
  const onLoginPage = pathname === "/login";

  useEffect(() => {
    let active = true;
    getCurrentUser()
      .then((u) => active && (setUser(u), setLoading(false)))
      .catch(() => active && (setUser(null), setLoading(false)));
    return () => {
      active = false;
    };
  }, [pathname]);

  useEffect(() => {
    if (!loading && !user && !onLoginPage) router.replace("/login");
  }, [loading, user, onLoginPage, router]);

  if (onLoginPage) return <>{children}</>;
  if (loading) {
    return <div className="p-10 text-sm text-gray-500">Loading…</div>;
  }
  if (!user) return null; // redirecting to /login

  async function handleLogout() {
    await logout();
    setUser(null);
    router.replace("/login");
  }

  return (
    <>
      <header className="border-b border-gray-200 bg-white">
        <nav className="mx-auto flex max-w-5xl items-center gap-6 px-6 py-4">
          <span className="font-semibold text-gray-900">SaltCRM</span>
          <Link href="/" className="text-sm text-gray-600 hover:text-gray-900">
            Customers
          </Link>
          <Link
            href="/m/accounts"
            className="text-sm text-gray-600 hover:text-gray-900"
          >
            Accounts
          </Link>
          <Link
            href="/m/contacts"
            className="text-sm text-gray-600 hover:text-gray-900"
          >
            Contacts
          </Link>
          <Link
            href="/m/leads"
            className="text-sm text-gray-600 hover:text-gray-900"
          >
            Leads
          </Link>
          <Link
            href="/products"
            className="text-sm text-gray-600 hover:text-gray-900"
          >
            Products
          </Link>
          <div className="ml-auto flex items-center gap-3 text-sm text-gray-500">
            <span>
              {user.name}
              {user.role ? ` · ${user.role}` : ""}
            </span>
            <button
              type="button"
              onClick={handleLogout}
              className="font-medium text-gray-700 hover:text-gray-900"
            >
              Log out
            </button>
          </div>
        </nav>
      </header>
      {children}
    </>
  );
}
