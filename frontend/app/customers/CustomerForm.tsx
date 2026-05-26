"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  createCustomer,
  updateCustomer,
  type Customer,
  type CustomerInput,
  type CustomerStatus,
} from "@/lib/api";

/**
 * Create/edit form for a customer. Passing `customer` switches the form into
 * edit mode (fields are pre-filled and submit calls updateCustomer); otherwise
 * it creates a new customer. On success it navigates to the customer's page.
 */
export default function CustomerForm({ customer }: { customer?: Customer }) {
  const router = useRouter();
  const isEdit = customer !== undefined;

  const [name, setName] = useState(customer?.name ?? "");
  const [email, setEmail] = useState(customer?.email ?? "");
  const [phone, setPhone] = useState(customer?.phone ?? "");
  const [serviceAddress, setServiceAddress] = useState(
    customer?.serviceAddress ?? "",
  );
  const [accountNumber, setAccountNumber] = useState(
    customer?.accountNumber ?? "",
  );
  const [status, setStatus] = useState<CustomerStatus>(
    customer?.status ?? "active",
  );
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    const input: CustomerInput = {
      name,
      email,
      phone,
      serviceAddress,
      accountNumber,
      status,
    };
    try {
      const saved =
        isEdit && customer
          ? await updateCustomer(customer.id, input)
          : await createCustomer(input);
      router.push(`/customers/${saved.id}`);
      router.refresh();
    } catch {
      setError(
        "Could not save the customer. Check the fields and that the backend is running.",
      );
      setBusy(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="max-w-lg space-y-5">
      {error && (
        <div
          role="alert"
          className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800"
        >
          {error}
        </div>
      )}

      <TextField label="Name" value={name} onChange={setName} required />
      <TextField
        label="Email"
        type="email"
        value={email}
        onChange={setEmail}
        required
      />
      <TextField label="Phone" value={phone} onChange={setPhone} />
      <TextField
        label="Service address"
        value={serviceAddress}
        onChange={setServiceAddress}
      />
      <TextField
        label="Account number"
        value={accountNumber}
        onChange={setAccountNumber}
        required
      />

      <label className="flex flex-col gap-1 text-sm font-medium text-gray-700">
        Status
        <select
          aria-label="Status"
          value={status}
          onChange={(e) => setStatus(e.target.value as CustomerStatus)}
          className="rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        >
          <option value="active">Active</option>
          <option value="suspended">Suspended</option>
        </select>
      </label>

      <button
        type="submit"
        disabled={busy}
        className="rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-gray-800 disabled:opacity-50"
      >
        {isEdit ? "Save changes" : "Create customer"}
      </button>
    </form>
  );
}

function TextField({
  label,
  value,
  onChange,
  type = "text",
  required = false,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
  required?: boolean;
}) {
  return (
    <label className="flex flex-col gap-1 text-sm font-medium text-gray-700">
      {label}
      <input
        type={type}
        aria-label={label}
        value={value}
        required={required}
        onChange={(e) => onChange(e.target.value)}
        className="rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
      />
    </label>
  );
}
