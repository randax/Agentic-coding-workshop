// Typed client for the SaltCRM backend API. This is the single seam through
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

export type Role = "admin" | "manager" | "agent";

export interface Agent {
  id: number;
  name: string;
  email: string;
  role?: Role;
  teamId?: number | null;
}

/** The authenticated user is an Agent (with a role). */
export type AuthUser = Agent;

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

// --- Authentication -------------------------------------------------------
// Auth requests use credentials:"include" so the HTTP-only session cookie is
// sent/received across the dev origins (Next :3000 ↔ API :8080).

/** Thrown by login when the credentials are rejected (HTTP 401). */
export class InvalidCredentialsError extends Error {}

/** Logs in with email/password, establishing a session cookie. */
export async function login(
  email: string,
  password: string,
): Promise<AuthUser> {
  const res = await fetch(`${API_BASE_URL}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ email, password }),
  });
  if (res.status === 401) {
    throw new InvalidCredentialsError("invalid email or password");
  }
  if (!res.ok) {
    throw new Error(`Login failed (HTTP ${res.status})`);
  }
  return res.json() as Promise<AuthUser>;
}

/** Ends the current session. */
export async function logout(): Promise<void> {
  await fetch(`${API_BASE_URL}/auth/logout`, {
    method: "POST",
    credentials: "include",
  });
}

/** Returns the currently-authenticated user, or null if not logged in. */
export async function getCurrentUser(): Promise<AuthUser | null> {
  const res = await fetch(`${API_BASE_URL}/auth/me`, {
    credentials: "include",
    cache: "no-store",
  });
  if (res.status === 401) return null;
  if (!res.ok) {
    throw new Error(`Failed to load current user (HTTP ${res.status})`);
  }
  return res.json() as Promise<AuthUser>;
}

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

/** Editable product fields, sent to the backend on create and edit. The
 * per-category attribute is optional and only meaningful for its category. */
export interface ProductInput {
  name: string;
  category: ProductCategory;
  monthlyPrice: number;
  speedMbps?: number;
  routerModel?: string;
  tvPackageTier?: string;
}

/** Creates a new product in the catalog. New products are available. */
export async function createProduct(input: ProductInput): Promise<Product> {
  const res = await fetch(`${API_BASE_URL}/products`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    throw new Error(`Failed to create product (HTTP ${res.status})`);
  }
  return res.json() as Promise<Product>;
}

/** Edits an existing product's catalog fields. Availability is server-managed
 * (via retire/unretire) and unaffected by edits. */
export async function updateProduct(
  id: string | number,
  input: ProductInput,
): Promise<Product> {
  const res = await fetch(`${API_BASE_URL}/products/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    throw new Error(`Failed to update product (HTTP ${res.status})`);
  }
  return res.json() as Promise<Product>;
}

/** Retires a product so it can no longer be subscribed to. */
export async function retireProduct(id: string | number): Promise<void> {
  const res = await fetch(`${API_BASE_URL}/products/${id}/retire`, {
    method: "POST",
  });
  if (!res.ok) {
    throw new Error(`Failed to retire product (HTTP ${res.status})`);
  }
}

/** Reactivates a retired product so it can be subscribed to again. */
export async function unretireProduct(id: string | number): Promise<void> {
  const res = await fetch(`${API_BASE_URL}/products/${id}/unretire`, {
    method: "POST",
  });
  if (!res.ok) {
    throw new Error(`Failed to unretire product (HTTP ${res.status})`);
  }
}

// --- Generic module metadata (drives the metadata-rendered views under /m) ---

export type FieldType = "string" | "enum" | "currency" | "bool" | "date";

export interface FieldMeta {
  name: string;
  type: FieldType;
  label: string;
  /** Present for `enum` fields. */
  options?: string[];
}

export interface PanelMeta {
  label: string;
  fields: string[];
}

export interface SubpanelMeta {
  label: string;
  /** Records endpoint with `{id}` replaced by the parent record's id. */
  path: string;
  columns: FieldMeta[];
}

export interface ActionMeta {
  label: string;
  method: string;
  /** Endpoint with `{id}` replaced by the record's id. */
  path: string;
}

export interface ModuleMeta {
  module: string;
  label: string;
  labelSingular: string;
  fields: FieldMeta[];
  listView: { columns: string[] };
  detailView?: { panels: PanelMeta[] };
  editView?: { fields: string[] };
  subpanels?: SubpanelMeta[];
  actions?: ActionMeta[];
}

/** A module record is an open bag of fields; the metadata says how to render them. */
export type ModuleRecord = Record<string, unknown>;

// Authenticated GET options. The session cookie is sent automatically in the
// browser via credentials:"include"; in server components there is no ambient
// cookie, so callers forward it explicitly as `cookie`.
function authGet(cookie?: string): RequestInit {
  const init: RequestInit = { cache: "no-store", credentials: "include" };
  if (cookie) init.headers = { Cookie: cookie };
  return init;
}

/** Fetches a module's metadata (fields + view layouts). Always fresh (uncached).
 * With `{ raw: true }` the backend returns the code+custom defaults *without*
 * saved layouts applied — the design-time palette the layout editor needs to
 * show currently-hidden fields. */
export async function getModuleMeta(
  module: string,
  cookie?: string,
  opts?: { raw?: boolean },
): Promise<ModuleMeta> {
  const qs = opts?.raw ? "?raw=1" : "";
  const res = await fetch(`${API_BASE_URL}/metadata/${module}${qs}`, authGet(cookie));
  if (!res.ok) {
    throw new Error(`Failed to load metadata for ${module} (HTTP ${res.status})`);
  }
  return res.json() as Promise<ModuleMeta>;
}

/** Fetches a module's records from its list endpoint (`/{module}`). Always fresh. */
export async function getModuleRecords(
  module: string,
  cookie?: string,
): Promise<ModuleRecord[]> {
  const res = await fetch(`${API_BASE_URL}/${module}`, authGet(cookie));
  if (!res.ok) {
    throw new Error(`Failed to load ${module} (HTTP ${res.status})`);
  }
  return res.json() as Promise<ModuleRecord[]>;
}

/** Fetches a single module record (`/{module}/{id}`). Returns null if not found. */
export async function getModuleRecord(
  module: string,
  id: string | number,
  cookie?: string,
): Promise<ModuleRecord | null> {
  const res = await fetch(`${API_BASE_URL}/${module}/${id}`, authGet(cookie));
  if (res.status === 404) return null;
  if (!res.ok) {
    throw new Error(`Failed to load ${module}/${id} (HTTP ${res.status})`);
  }
  return res.json() as Promise<ModuleRecord>;
}

/** Updates a single module record (`PUT /{module}/{id}`) with the given field values. */
export async function updateModuleRecord(
  module: string,
  id: string | number,
  values: ModuleRecord,
): Promise<ModuleRecord> {
  const res = await fetch(`${API_BASE_URL}/${module}/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(values),
  });
  if (!res.ok) {
    throw new Error(`Failed to update ${module}/${id} (HTTP ${res.status})`);
  }
  return res.json() as Promise<ModuleRecord>;
}

/** One stage column of the sales pipeline, with rolled-up count and total. */
export interface PipelineStage {
  stage: string;
  count: number;
  totalAmount: number;
  items: ModuleRecord[];
}

// --- Dashboard ------------------------------------------------------------

/** A task on the "My Tasks" dashlet. */
export interface DashboardTask {
  id: number;
  subject: string;
  status: string;
  /** ISO-8601 timestamp. */
  occurredAt: string;
}

/** A lead on the "Recent Leads" dashlet. */
export interface DashboardLead {
  id: number;
  name: string;
  company: string;
  status: string;
}

/** The signed-in user's dashboard: each dashlet already scoped to what they see. */
export interface DashboardData {
  myOpenCases: Case[];
  myTasks: DashboardTask[];
  recentLeads: DashboardLead[];
  pipelineByStage: PipelineStage[];
}

/** Fetches the signed-in user's dashboard (all dashlets). Always fresh. */
export async function getDashboard(cookie?: string): Promise<DashboardData> {
  const res = await fetch(`${API_BASE_URL}/dashboard`, authGet(cookie));
  if (!res.ok) {
    throw new Error(`Failed to load dashboard (HTTP ${res.status})`);
  }
  return res.json() as Promise<DashboardData>;
}

// --- Global search --------------------------------------------------------

/** The modules global search spans. */
export type SearchModule =
  | "accounts"
  | "contacts"
  | "leads"
  | "opportunities"
  | "cases";

/** A single matched record. */
export interface SearchHit {
  module: SearchModule;
  id: number;
  title: string;
}

/** Matched records for one module (already ranked by the backend). */
export interface SearchGroup {
  module: SearchModule;
  hits: SearchHit[];
}

/** Runs a global search across modules, returning hits grouped by module —
 * already scoped to what the signed-in user may see. Always fresh. */
export async function globalSearch(
  query: string,
  cookie?: string,
): Promise<SearchGroup[]> {
  const res = await fetch(
    `${API_BASE_URL}/search?q=${encodeURIComponent(query)}`,
    authGet(cookie),
  );
  if (!res.ok) {
    throw new Error(`Search failed (HTTP ${res.status})`);
  }
  const body = (await res.json()) as { groups: SearchGroup[] };
  return body.groups;
}

// --- Studio (custom field definitions) -----------------------------------

export interface FieldDef {
  id: number;
  module: string;
  name: string;
  type: FieldType;
  label: string;
  options?: string[];
}

export interface FieldDefInput {
  module: string;
  name: string;
  type: FieldType;
  label: string;
  options?: string[];
}

/** Lists the custom fields defined on a module. */
export async function getCustomFields(
  module: string,
  cookie?: string,
): Promise<FieldDef[]> {
  const res = await fetch(`${API_BASE_URL}/studio/fields?module=${module}`, authGet(cookie));
  if (!res.ok) {
    throw new Error(`Failed to load custom fields (HTTP ${res.status})`);
  }
  return res.json() as Promise<FieldDef[]>;
}

/** Defines a new custom field on a module (admin only). */
export async function addCustomField(input: FieldDefInput): Promise<FieldDef> {
  const res = await fetch(`${API_BASE_URL}/studio/fields`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    throw new Error(`Failed to add custom field (HTTP ${res.status})`);
  }
  return res.json() as Promise<FieldDef>;
}

// --- Studio (view layouts) ------------------------------------------------

/** A module's saved layouts: the ordered, visible field names per view. A view
 * with no saved layout is absent (the generic views fall back to its default). */
export interface ModuleLayouts {
  list?: string[];
  detail?: string[];
  edit?: string[];
}

/** Fetches a module's saved view layouts (empty object if none saved). */
export async function getLayouts(
  module: string,
  cookie?: string,
): Promise<ModuleLayouts> {
  const res = await fetch(`${API_BASE_URL}/studio/layouts?module=${module}`, authGet(cookie));
  if (!res.ok) {
    throw new Error(`Failed to load layouts (HTTP ${res.status})`);
  }
  return res.json() as Promise<ModuleLayouts>;
}

/** Saves a module's view layouts (admin only). Each provided view replaces its
 * saved layout; the values are the ordered, visible field names for that view. */
export async function saveLayouts(
  module: string,
  views: ModuleLayouts,
): Promise<void> {
  const res = await fetch(`${API_BASE_URL}/studio/layouts`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ module, views }),
  });
  if (!res.ok) {
    throw new Error(`Failed to save layout (HTTP ${res.status})`);
  }
}

/** Runs a record action (e.g. convert a lead), substituting the id into its path. */
export async function runRecordAction(
  action: ActionMeta,
  id: string | number,
): Promise<void> {
  const url = `${API_BASE_URL}${action.path.replace("{id}", String(id))}`;
  const res = await fetch(url, { method: action.method, credentials: "include" });
  if (!res.ok) {
    throw new Error(`Action "${action.label}" failed (HTTP ${res.status})`);
  }
}

/** Fetches the sales pipeline grouped by stage. */
export async function getPipeline(cookie?: string): Promise<PipelineStage[]> {
  const res = await fetch(`${API_BASE_URL}/opportunities/pipeline`, authGet(cookie));
  if (!res.ok) {
    throw new Error(`Failed to load pipeline (HTTP ${res.status})`);
  }
  return res.json() as Promise<PipelineStage[]>;
}

/** Fetches a subpanel's related records, substituting the parent id into its path. */
export async function getSubpanelRecords(
  path: string,
  parentId: string | number,
  cookie?: string,
): Promise<ModuleRecord[]> {
  const url = `${API_BASE_URL}${path.replace("{id}", String(parentId))}`;
  const res = await fetch(url, authGet(cookie));
  if (!res.ok) {
    throw new Error(`Failed to load related records (HTTP ${res.status})`);
  }
  return res.json() as Promise<ModuleRecord[]>;
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

// --- Reports --------------------------------------------------------------

/** How a report rolls each group up. */
export type ReportAggregation = "count" | "sum" | "avg";

/** How a filter compares a field against its value. */
export type ReportOperator = "eq" | "contains" | "gt" | "lt";

/** One filter condition. `field` may name a core or custom field. */
export interface ReportFilter {
  field: string;
  operator: ReportOperator;
  value: string | number;
}

/** A report query: the module, conditions, grouping, and aggregation. */
export interface ReportDefinition {
  module: string;
  filters?: ReportFilter[];
  groupBy: string;
  aggregation: ReportAggregation;
  /** The field summed/averaged; ignored for `count`. */
  aggField?: string;
}

/** One aggregated group of a report result. */
export interface ReportRow {
  group: string;
  count: number;
  value: number;
}

/** The aggregated rows a report run produces. */
export interface ReportResult {
  rows: ReportRow[];
}

/** A saved, re-runnable report. */
export interface SavedReport {
  id: number;
  name: string;
  definition: ReportDefinition;
}

/** Runs an ad-hoc report; results are scoped to what the signed-in user may see. */
export async function runReport(
  def: ReportDefinition,
  cookie?: string,
): Promise<ReportResult> {
  const res = await fetch(`${API_BASE_URL}/reports/run`, {
    method: "POST",
    headers: { "Content-Type": "application/json", ...(cookie ? { Cookie: cookie } : {}) },
    credentials: "include",
    body: JSON.stringify(def),
  });
  if (!res.ok) {
    throw new Error(`Failed to run report (HTTP ${res.status})`);
  }
  return res.json() as Promise<ReportResult>;
}

/** Saves a named report so it can be re-run later. */
export async function saveReport(
  name: string,
  definition: ReportDefinition,
): Promise<SavedReport> {
  const res = await fetch(`${API_BASE_URL}/reports`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ name, definition }),
  });
  if (!res.ok) {
    throw new Error(`Failed to save report (HTTP ${res.status})`);
  }
  return res.json() as Promise<SavedReport>;
}

/** Lists the saved reports. Always fresh. */
export async function getSavedReports(cookie?: string): Promise<SavedReport[]> {
  const res = await fetch(`${API_BASE_URL}/reports`, authGet(cookie));
  if (!res.ok) {
    throw new Error(`Failed to load saved reports (HTTP ${res.status})`);
  }
  return res.json() as Promise<SavedReport[]>;
}

/** Re-runs a saved report by id; results are scoped to the signed-in user. */
export async function runSavedReport(
  id: number,
  cookie?: string,
): Promise<ReportResult> {
  const res = await fetch(`${API_BASE_URL}/reports/${id}/run`, {
    method: "POST",
    credentials: "include",
    headers: cookie ? { Cookie: cookie } : undefined,
  });
  if (!res.ok) {
    throw new Error(`Failed to run saved report (HTTP ${res.status})`);
  }
  return res.json() as Promise<ReportResult>;
}
