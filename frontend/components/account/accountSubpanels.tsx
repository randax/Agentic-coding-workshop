import {
  getCustomerCases,
  getCustomerSubscriptions,
  getProducts,
} from "@/lib/api";
import NewCaseForm from "./NewCaseForm";
import CaseList from "./CaseList";
import SubscriptionList from "./SubscriptionList";

/**
 * Builds the bespoke subpanel bodies for an account record view — the two
 * panels that carry write actions the generic read-only subpanels can't:
 *   • Cases — an "open case" form above the case list.
 *   • Subscriptions — assign a product and cancel a subscription.
 * Returned keyed by subpanel label for RecordView's `subpanelOverrides`.
 */
export async function accountSubpanelOverrides(
  accountId: string,
): Promise<Record<string, React.ReactNode>> {
  const [cases, subscriptions, products] = await Promise.all([
    getCustomerCases(accountId),
    getCustomerSubscriptions(accountId),
    getProducts(),
  ]);
  const id = Number(accountId);
  return {
    Cases: (
      <div className="space-y-6">
        <NewCaseForm customerId={id} />
        <CaseList cases={cases} />
      </div>
    ),
    Subscriptions: (
      <SubscriptionList
        customerId={id}
        subscriptions={subscriptions}
        availableProducts={products.filter((p) => p.available)}
      />
    ),
  };
}
