import Link from "next/link";
import CustomerForm from "@/app/customers/CustomerForm";

export default function NewCustomerPage() {
  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10">
      <Link href="/" className="text-sm text-gray-500 hover:text-gray-800">
        ← Customers
      </Link>
      <h1 className="mt-4 text-2xl font-semibold tracking-tight text-gray-900">
        New customer
      </h1>
      <div className="mt-6">
        <CustomerForm />
      </div>
    </main>
  );
}
