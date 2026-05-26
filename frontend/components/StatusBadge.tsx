import type { CustomerStatus } from "@/lib/api";

const statusStyles: Record<CustomerStatus, string> = {
  active: "bg-green-100 text-green-800 ring-green-600/20",
  suspended: "bg-amber-100 text-amber-800 ring-amber-600/20",
};

/** A pill showing a customer's account status. */
export default function StatusBadge({ status }: { status: CustomerStatus }) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${statusStyles[status]}`}
    >
      {status}
    </span>
  );
}
