import StudioForm from "./StudioForm";

// Modules that support custom fields (those with a custom_fields column today).
const MODULES = ["accounts"];

export default function StudioPage() {
  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10">
      <header className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight text-gray-900">
          Studio
        </h1>
        <p className="mt-1 text-sm text-gray-500">
          Add custom fields to a module. New fields appear in that module&apos;s
          list, record, and edit views immediately.
        </p>
      </header>
      <StudioForm modules={MODULES} />
    </main>
  );
}
