# SaltCRM — Agent Guidelines

Internal CRM for an ISP (fiber, mesh routers, TV subscriptions). Go API backend,
Next.js frontend, SQLite. This file is the project-wide rulebook for agents.
Structure inspired by [waddle's AGENTS.md](https://github.com/waddle-social/waddle/blob/main/AGENTS.md).

> **Nested guidance wins for its subtree.** `frontend/AGENTS.md` governs frontend
> code and overrides anything here when they conflict. Read it before touching
> `frontend/`.

## Active technologies

- **Backend** (`backend/`): Go 1.26 · Gin (HTTP) · GORM over SQLite · module `saltcrm`
- **Frontend** (`frontend/`): Next.js 16 (App Router) · React 19 · TypeScript 5 (strict) · Tailwind 4
- **Tests**: Go `testing` (backend) · Vitest + React Testing Library + jsdom (frontend)

## Commands

Run each from its own directory. The backend must be running for the frontend to load data.

```bash
# Backend (from backend/)
go run ./cmd/server        # serve on :8080; creates+seeds saltcrm.db on first run
go build ./...             # compile check
go test ./...              # all tests

# Frontend (from frontend/)
npm install                # first time only
npm run dev                # dev server on :3000
npm test                   # Vitest (run once)
npm run typecheck          # tsc --noEmit
npm run build              # production build
```

Default seeded login: `sam@isp.example` / `password` (admin). Also `robin@` (manager), `lee@` (agent).

## Backend architecture — the dependency rule

The backend is strictly layered. Dependencies point **inward**; never the reverse.

```
cmd/server        composition root: open DB, migrate, seed, wire services, serve
   │ wires
internal/api      HTTP only — translate request→service call→JSON. NO business logic.
   │ calls
internal/<domain> domain model + Service + a Repository INTERFACE. Imports no DB.
   ▲ implemented by
internal/store    GORM repositories implementing each domain's Repository interface.
```

**Hard rules:**

- **Business logic lives in `internal/<domain>` Services**, behind a `Repository`
  interface the service defines. A domain package must **not** import GORM, Gin,
  or `internal/store`. (See `internal/lead/lead.go`, `internal/customer/customer.go`.)
- **`internal/store` is the only place that touches GORM.** Each repo implements
  a domain's `Repository` interface. Add persistence here, not in services.
- **`internal/api` holds no business logic.** Handlers parse input, call a
  service, map errors to status codes (e.g. `ErrNotFound` → 404), and serialize.
- **Prefer deep modules**: a small, stable interface hiding real behavior. New
  cross-domain behavior (e.g. lead→account conversion) belongs in its own domain
  service that orchestrates others through their interfaces — not smeared across
  handlers.
- **Wire new services in `cmd/server` and pass them to `api.NewRouter`.** That is
  the single composition root.

### Auth, roles & visibility

- Authenticated routes go through `requireAuth` (session cookie → `currentUser`).
- Role-gate writes with `requireRole(agent.RoleManager, agent.RoleAdmin, …)`.
- Records are **visibility-scoped (own-or-team)** via `internal/access` — new
  record surfaces must respect it. Don't add unauthenticated data routes.
- "**Accounts**" is the SaltCRM-facing name for customer records; the generic
  `/m/<module>` views read through the authenticated, scoped surfaces.

### Data & schema

- SQLite, auto-migrated and seeded on first run. `*.db` is gitignored.
- **Stale DB / 401 / wrong data after a schema change?** Delete `backend/saltcrm.db`
  and restart to re-seed.
- No production data exists. Prefer clean schema changes over compatibility shims.

## Frontend architecture

- **`lib/api.ts` is the single seam to the backend.** All data access goes through
  its typed functions; TS types mirror the Go JSON shapes. New endpoint → add a
  typed function and its interfaces here, never `fetch` ad hoc in a component.
- Components colocate under `app/`; the `@` alias maps to the frontend root.
- **This is Next.js 16 with breaking changes from older versions.** Per
  `frontend/AGENTS.md`, read the relevant guide in `node_modules/next/dist/docs/`
  before writing frontend code — do not assume APIs from memory.

## Testing

Test **external behavior, not implementation details.** A test should survive a
refactor that preserves behavior.

- **Backend unit tests** run a Service against an **in-memory fake repository** in
  the same package (`fakeRepo` pattern; see `internal/customer/service_test.go`).
  No database.
- **Backend integration tests** exercise the real router over a throwaway SQLite
  DB via `httptest` (`internal/api/router_test.go`).
- **Frontend tests** are `*.test.tsx` colocated with the component (Vitest + RTL).
  Mock `@/lib/api` and `next/navigation`; assert on rendered output and the API
  calls made, not internals (see `app/products/ProductForm.test.tsx`).
- New behavior ships with tests. Prefer test-first (`/tdd`).

## Feedback loops — run before every commit

Run the loops for whichever layer(s) you touched; both if you touched both:

```bash
cd backend  && go build ./... && go test ./...
cd frontend && npm test && npm run typecheck
```

All loops you ran must pass before committing. (This matches `ralph/prompt.md`,
which the AFK agent obeys.)

## Branch, commit & PR workflow

- Branch per feature (`<user>/saltcrm-<slug>`); open a PR to `main`. Avoid
  committing straight to `main`.
- Keep commits small and focused — ideally one vertical slice each.
- Commit messages: imperative subject, then a body covering **key decisions**,
  **files changed**, and **blockers/notes for next time** (this is the repo's
  established convention; the AFK agent follows it too).

## Repo map

```
backend/   Go API (cmd/server, internal/{<domain>,api,store,access,identity,metadata,studio,seed})
frontend/  Next.js app (app/, lib/api.ts, components/) — see frontend/AGENTS.md
issues/    Local kanban board for the agentic pipeline (issues/done/ = done)
ralph/     AFK agent runner (once.sh, afk.sh, prompt.md)
docs/      Assignment + design docs
.claude/   Project skills (grill-me, write-a-prd, prd-to-issues, tdd, …)
```
