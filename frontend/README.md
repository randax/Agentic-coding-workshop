# ISP CRM — Frontend

Next.js 16 (App Router) + Tailwind CSS. Renders the CRM UI by calling the Go
backend through a typed API client (`lib/api.ts`).

## Requirements

- Node.js 20+
- The backend running (see `../backend/README.md`)

## Run

```bash
cd frontend
npm install        # first time only
npm run dev        # dev server at http://localhost:3000
```

Or a production build:

```bash
npm run build
npm run start
```

By default the app calls the backend at `http://localhost:8080`. Override with:

```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080 npm run dev
```

If the backend is unreachable, the customer list shows a friendly error instead
of crashing.

## Test

```bash
npm test          # run once (Vitest + React Testing Library, jsdom)
npm run test:watch
```

Tests exercise presentational components through the rendered DOM (behavior, not
implementation) — e.g. `app/customers/[id]/CustomerDetail.test.tsx`,
`components/StatusBadge.test.tsx`. Async Server Component pages are verified via
build + end-to-end rather than unit tests.

## Structure

```
app/page.tsx                       Customer list page (async Server Component)
app/customers/[id]/page.tsx        Customer detail page (reads ?tab=, fetches, renders)
app/customers/[id]/CustomerDetail  Pure, tested presentational component (tabs + panels)
app/layout.tsx                     Root layout
components/StatusBadge.tsx         Shared status pill
lib/api.ts                         Typed API client — the single seam to the backend
lib/format.ts                      Shared formatting helpers
```
