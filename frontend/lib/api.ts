// Typed client for the ISP CRM backend API. This is the single seam through
// which the frontend talks to the Go backend; all data access goes through here.

export type CustomerStatus = "active" | "suspended";

export interface Customer {
  id: number;
  name: string;
  email: string;
  phone: string;
  serviceAddress: string;
  accountNumber: string;
  /** ISO-8601 timestamp. */
  customerSince: string;
  status: CustomerStatus;
}

export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

/** Fetches all customers. Data is always fresh (uncached). */
export async function getCustomers(): Promise<Customer[]> {
  const res = await fetch(`${API_BASE_URL}/customers`, { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`Failed to load customers (HTTP ${res.status})`);
  }
  return res.json() as Promise<Customer[]>;
}

/** Fetches a single customer by id. Returns null if no such customer exists. */
export async function getCustomer(
  id: string | number,
): Promise<Customer | null> {
  const res = await fetch(`${API_BASE_URL}/customers/${id}`, {
    cache: "no-store",
  });
  if (res.status === 404) {
    return null;
  }
  if (!res.ok) {
    throw new Error(`Failed to load customer (HTTP ${res.status})`);
  }
  return res.json() as Promise<Customer>;
}
