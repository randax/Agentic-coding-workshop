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

export type ProductCategory = "fiber" | "router" | "tv";

export interface Product {
  id: number;
  name: string;
  category: ProductCategory;
  monthlyPrice: number;
  available: boolean;
  speedMbps?: number;
  routerModel?: string;
  tvPackageTier?: string;
}

export type SubscriptionStatus = "active" | "pending" | "cancelled";

export interface Subscription {
  id: number;
  customerId: number;
  productId: number;
  status: SubscriptionStatus;
  startDate: string;
  endDate?: string | null;
  monthlyPriceSnapshot: number;
  quantity: number;
  product: Product;
}

export interface Agent {
  id: number;
  name: string;
  email: string;
}

export type CaseCategory =
  | "billing"
  | "connectivity"
  | "hardware"
  | "tv"
  | "general";
export type CasePriority = "low" | "medium" | "high" | "urgent";
export type CaseStatus = "open" | "in_progress" | "resolved" | "closed";

export interface CaseComment {
  id: number;
  caseId: number;
  body: string;
  authorAgentId?: number | null;
  authorAgent?: Agent | null;
  /** ISO-8601 timestamp. */
  createdAt: string;
}

export interface Case {
  id: number;
  customerId: number;
  subject: string;
  description: string;
  category: CaseCategory;
  priority: CasePriority;
  status: CaseStatus;
  assignedAgentId?: number | null;
  assignedAgent?: Agent | null;
  /** Chronological timeline (oldest first), present on the case-detail response. */
  comments?: CaseComment[];
  /** ISO-8601 timestamps. */
  createdAt: string;
  updatedAt: string;
}

export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

/** Optional filters for the customer list, forwarded to the backend as query params. */
export interface CustomerFilters {
  /** Partial, case-insensitive term matched against name, email, or account number. */
  search?: string;
  status?: CustomerStatus;
}

/** Fetches customers, optionally filtered by search term and status. Always fresh (uncached). */
export async function getCustomers(
  filters: CustomerFilters = {},
): Promise<Customer[]> {
  const params = new URLSearchParams();
  if (filters.search) params.set("search", filters.search);
  if (filters.status) params.set("status", filters.status);
  const qs = params.toString();
  const url = `${API_BASE_URL}/customers${qs ? `?${qs}` : ""}`;

  const res = await fetch(url, { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`Failed to load customers (HTTP ${res.status})`);
  }
  return res.json() as Promise<Customer[]>;
}

/** Fetches the product catalog. Data is always fresh (uncached). */
export async function getProducts(): Promise<Product[]> {
  const res = await fetch(`${API_BASE_URL}/products`, { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`Failed to load products (HTTP ${res.status})`);
  }
  return res.json() as Promise<Product[]>;
}

/** Fetches all support agents. Data is always fresh (uncached). */
export async function getAgents(): Promise<Agent[]> {
  const res = await fetch(`${API_BASE_URL}/agents`, { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`Failed to load agents (HTTP ${res.status})`);
  }
  return res.json() as Promise<Agent[]>;
}

/** Fetches a customer's support cases. Data is always fresh (uncached). */
export async function getCustomerCases(
  id: string | number,
): Promise<Case[]> {
  const res = await fetch(`${API_BASE_URL}/customers/${id}/cases`, {
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error(`Failed to load cases (HTTP ${res.status})`);
  }
  return res.json() as Promise<Case[]>;
}

/** Fields needed to open a new case. Status is always Open (server-assigned). */
export interface CaseInput {
  subject: string;
  description: string;
  category: CaseCategory;
  priority: CasePriority;
}

/** Opens a new case for a customer. */
export async function createCase(
  customerId: string | number,
  input: CaseInput,
): Promise<Case> {
  const res = await fetch(`${API_BASE_URL}/customers/${customerId}/cases`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    throw new Error(`Failed to open case (HTTP ${res.status})`);
  }
  return res.json() as Promise<Case>;
}

/** Appends a comment to a case's timeline, attributed to the chosen agent. */
export async function addCaseComment(
  caseId: string | number,
  input: { body: string; authorAgentId: number },
): Promise<CaseComment> {
  const res = await fetch(`${API_BASE_URL}/cases/${caseId}/comments`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    throw new Error(`Failed to add comment (HTTP ${res.status})`);
  }
  return res.json() as Promise<CaseComment>;
}

/** Advances a case to a new status. Rejected by the backend if the transition is illegal. */
export async function updateCaseStatus(
  caseId: string | number,
  status: CaseStatus,
): Promise<Case> {
  const res = await fetch(`${API_BASE_URL}/cases/${caseId}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ status }),
  });
  if (!res.ok) {
    throw new Error(`Failed to change case status (HTTP ${res.status})`);
  }
  return res.json() as Promise<Case>;
}

/** Updates a case's metadata: assign/reassign an agent, adjust priority/category. */
export async function updateCaseMetadata(
  caseId: string | number,
  patch: {
    priority?: CasePriority;
    category?: CaseCategory;
    assignedAgentId?: number;
  },
): Promise<Case> {
  const res = await fetch(`${API_BASE_URL}/cases/${caseId}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(patch),
  });
  if (!res.ok) {
    throw new Error(`Failed to update case (HTTP ${res.status})`);
  }
  return res.json() as Promise<Case>;
}

/** Fetches a single case with its comment timeline. Returns null if not found. */
export async function getCase(id: string | number): Promise<Case | null> {
  const res = await fetch(`${API_BASE_URL}/cases/${id}`, { cache: "no-store" });
  if (res.status === 404) {
    return null;
  }
  if (!res.ok) {
    throw new Error(`Failed to load case (HTTP ${res.status})`);
  }
  return res.json() as Promise<Case>;
}

/** Fetches a customer's subscriptions. Data is always fresh (uncached). */
export async function getCustomerSubscriptions(
  id: string | number,
): Promise<Subscription[]> {
  const res = await fetch(`${API_BASE_URL}/customers/${id}/subscriptions`, {
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error(`Failed to load subscriptions (HTTP ${res.status})`);
  }
  return res.json() as Promise<Subscription[]>;
}

/** Assigns a catalog product to a customer, creating a new subscription. */
export async function createSubscription(
  customerId: string | number,
  input: { productId: number; quantity: number },
): Promise<Subscription> {
  const res = await fetch(`${API_BASE_URL}/customers/${customerId}/subscriptions`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    throw new Error(`Failed to create subscription (HTTP ${res.status})`);
  }
  return res.json() as Promise<Subscription>;
}

/** Cancels a subscription, setting its status to cancelled and an end date. */
export async function cancelSubscription(
  id: string | number,
): Promise<Subscription> {
  const res = await fetch(`${API_BASE_URL}/subscriptions/${id}/cancel`, {
    method: "POST",
  });
  if (!res.ok) {
    throw new Error(`Failed to cancel subscription (HTTP ${res.status})`);
  }
  return res.json() as Promise<Subscription>;
}

/** Editable customer profile fields, sent to the backend on create and edit. */
export interface CustomerInput {
  name: string;
  email: string;
  phone: string;
  serviceAddress: string;
  accountNumber: string;
  status: CustomerStatus;
}

/** Creates a new customer. */
export async function createCustomer(input: CustomerInput): Promise<Customer> {
  const res = await fetch(`${API_BASE_URL}/customers`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    throw new Error(`Failed to create customer (HTTP ${res.status})`);
  }
  return res.json() as Promise<Customer>;
}

/** Edits an existing customer's profile fields and status. */
export async function updateCustomer(
  id: string | number,
  input: CustomerInput,
): Promise<Customer> {
  const res = await fetch(`${API_BASE_URL}/customers/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    throw new Error(`Failed to update customer (HTTP ${res.status})`);
  }
  return res.json() as Promise<Customer>;
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
