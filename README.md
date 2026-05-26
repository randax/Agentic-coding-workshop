# ISP CRM

A simple internal CRM for an ISP that sells fiber internet, mesh routers, and TV subscriptions.
Customer-service agents can view customers, the products/subscriptions each customer has, and the
support cases they've filed.

See the [product requirements](../../issues/1) for the full scope.

## Stack

- **Backend:** Go (Gin + GORM) over SQLite
- **Frontend:** Next.js 16 (App Router) + Tailwind CSS

## Layout

```
backend/    Go API server
frontend/   Next.js web app
```

## Running locally

See `backend/README.md` and `frontend/README.md` (added as the project takes shape).
