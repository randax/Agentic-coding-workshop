# ISP CRM — Backend

Go API server (Gin + GORM over SQLite).

## Requirements

- Go 1.26+
- A C compiler (cgo is required by the SQLite driver; preinstalled on macOS/Linux dev machines)

## Run

```bash
cd backend
go run ./cmd/server
```

The server listens on `:8080`, creating `ispcrm.db` (SQLite) in the working
directory and seeding a few demo customers on first run.

Configuration (environment variables):

| Variable      | Default      | Description                     |
| ------------- | ------------ | ------------------------------- |
| `ISPCRM_ADDR` | `:8080`      | Address the HTTP server binds   |
| `ISPCRM_DB`   | `ispcrm.db`  | SQLite database file (DSN)      |

## API

| Method | Path         | Description         |
| ------ | ------------ | ------------------- |
| GET    | `/customers` | List all customers  |

```bash
curl http://localhost:8080/customers
```

## Test

```bash
cd backend
go test ./...
```

- Service-layer **unit tests** run against an in-memory fake repository
  (`internal/customer/service_test.go`).
- **API integration tests** exercise the real router over a throwaway SQLite
  database via `httptest` (`internal/api/router_test.go`).

## Layout

```
cmd/server/        Entry point (composition root): open DB, migrate, seed, serve
internal/customer/ Domain model + service (business logic) + Repository interface
internal/store/    GORM-backed persistence (Open, Migrate, repositories)
internal/api/      Gin router + handlers (HTTP <-> service translation)
internal/seed/     Idempotent demo-data seeding
```

The service layer depends on the `Repository` interface, not on GORM, so business
logic can be unit-tested in isolation. HTTP and persistence are thin layers around it.
