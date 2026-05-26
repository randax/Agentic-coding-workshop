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

## Structure

```
app/page.tsx   Customer list page (async Server Component)
app/layout.tsx Root layout
lib/api.ts     Typed API client — the single seam to the backend
```
