import StudioForm from "./StudioForm";
import LayoutStudio from "./LayoutStudio";

// Modules that support custom fields (those with a custom_fields column today).
const CUSTOM_FIELD_MODULES = ["accounts"];

// Modules whose layouts can be arranged. Layouts need no custom_fields column,
// so every module rendered by the generic /m views qualifies (kept in sync with
// the backend metadata registry).
const LAYOUT_MODULES = [
  "accounts",
  "contacts",
  "leads",
  "opportunities",
  "products",
  "subscriptions",
  "activities",
];

export default function StudioPage() {
  return (
    <main className="mx-auto w-full max-w-5xl px-6 py-10">
      <header className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight text-gray-900">
          Studio
        </h1>
        <p className="mt-1 text-sm text-gray-500">
          Configure modules without code: add custom fields and arrange how each
          view shows its fields. Changes apply for everyone.
        </p>
      </header>

      <div className="space-y-12">
        <section>
          <h2 className="mb-1 text-lg font-semibold text-gray-900">Custom fields</h2>
          <p className="mb-4 text-sm text-gray-500">
            Add a custom field to a module. New fields appear in that module&apos;s
            list, record, and edit views immediately.
          </p>
          <StudioForm modules={CUSTOM_FIELD_MODULES} />
        </section>

        <section>
          <h2 className="mb-1 text-lg font-semibold text-gray-900">Layouts</h2>
          <p className="mb-4 text-sm text-gray-500">
            Choose which fields appear and in what order for a module&apos;s list,
            record, and edit views.
          </p>
          <LayoutStudio modules={LAYOUT_MODULES} />
        </section>
      </div>
    </main>
  );
}
