# Setting up the repo

**Goal:** Get SaltCRM running locally — backend + frontend — and confirm the
agentic tooling is in place, so every later step "just works."

**You'll produce:** A SaltCRM you can log into in the browser, plus a green test
baseline and verified skills/runner.

## Why this step

The whole assignment hinges on a working app and working tooling. The agent in
[step 5](05-running-your-afk-agent.md) builds and tests real code; if your
backend won't compile or your tests don't run, you won't be able to tell whether
the agent succeeded. A few minutes of setup now saves an hour of confusion later.

> **Commands are shown for both platforms.** Look for the **macOS / Linux** and
> **Windows (PowerShell)** labels. On Windows, use **PowerShell 7+** or **Git
> Bash** — in legacy *Windows PowerShell 5.1* the `&&` chaining operator isn't
> supported, so run those lines one at a time.

## Before you start

You need these installed. Check each (these commands are the same on every OS):

```bash
go version        # want go1.26 or newer
node --version    # want v20 or newer
git --version     # any recent version
claude --version  # Claude Code CLI
gcc --version     # a C compiler — see below (the SQLite driver needs cgo)
```

Install anything that's missing:

- **Go 1.26+** — https://go.dev/dl/ (Windows: the `.msi` installer).
- **Node.js 20+** — https://nodejs.org/ (Windows: the `.msi` installer; pick LTS).
- **Git** — https://git-scm.com/downloads (Windows: this also gives you **Git
  Bash**).
- **Claude Code** — you're already in it if you're reading this from the CLI.
- **A C compiler** — required because the SQLite driver (`mattn/go-sqlite3`) uses
  cgo. This is the step most likely to trip you up on Windows:

  **macOS:**
  ```bash
  xcode-select --install
  ```

  **Linux (Debian/Ubuntu):**
  ```bash
  sudo apt-get install build-essential
  ```

  **Windows:** install a GCC toolchain and make sure `gcc` is on your `PATH`.
  Easiest is **TDM-GCC** (https://jmeubank.github.io/tdm-gcc/) — run the
  installer and it adds `gcc` to your `PATH`. (Alternative: **MSYS2**
  (https://www.msys2.org/), then `pacman -S mingw-w64-ucrt-x86_64-gcc` and add
  `C:\msys64\ucrt64\bin` to your `PATH`.) Open a **new** terminal and confirm:

  ```powershell
  gcc --version
  go env -w CGO_ENABLED=1   # ensure cgo is on
  ```

You do **not** need Docker for this assignment.

## Steps

### 1. Get the code

If you haven't already, clone the repo and open it:

```bash
git clone <the-repo-url> Agentic-coding-workshop
cd Agentic-coding-workshop
```

Run everything below **from the repo root** unless a step says otherwise.

### 2. Start the backend (Go API)

```bash
cd backend
go run ./cmd/server
```

The first run downloads dependencies, compiles the SQLite driver (this is where a
missing C compiler shows up as a build error — see *Before you start*), creates
`saltcrm.db` in `backend/`, and **seeds demo data**. It listens on
**http://localhost:8080**.

Leave it running. Sanity-check it from a second terminal — or just open the URL
in your browser:

**macOS / Linux:**
```bash
curl http://localhost:8080/customers
```

**Windows (PowerShell):**
```powershell
curl.exe http://localhost:8080/customers
```

> On Windows, plain `curl` in PowerShell is an alias for `Invoke-WebRequest`,
> which formats output differently — use `curl.exe` (or just the browser).

You should get back a JSON array of customers.

### 3. Start the frontend (Next.js)

In a **second terminal**, from the repo root:

```bash
cd frontend
npm install      # first time only
npm run dev      # http://localhost:3000
```

Open **http://localhost:3000** in your browser.

### 4. Log in

The seed creates demo agents, all with the password `password`:

| Email | Role |
|-------|------|
| `sam@isp.example` | Admin |
| `robin@isp.example` | Manager |
| `lee@isp.example` | Agent |

Log in as **`sam@isp.example` / `password`**. Click into **Leads** — you'll see a
few seeded leads, including **Sofia Berg (Polar Foods)**, whose status is
`qualified`. That's the lead you'll convert later.

### 5. Establish the green test baseline

You should be able to run both test suites and see them pass *before* you change
anything.

**macOS / Linux:**
```bash
# Backend
cd backend && go build ./... && go test ./...

# Frontend (from repo root again)
cd frontend && npm test && npm run typecheck
```

**Windows (PowerShell):**
```powershell
# Backend
cd backend; go build ./...; go test ./...

# Frontend (from repo root again)
cd frontend; npm test; npm run typecheck
```

Everything should pass. These are the exact **feedback loops** the agent runs in
step 5 — if they're green now, you'll know any later red is your feature, not
your environment.

### 6. Confirm the tooling is present

**macOS / Linux:**
```bash
ls .claude/skills    # grill-me, write-a-prd, prd-to-issues, tdd, ...
ls ralph             # afk.sh, once.sh, prompt.md
ls issues            # .gitkeep, done/   (this is your kanban board)
```

**Windows (PowerShell):**
```powershell
ls .claude\skills
ls ralph
ls issues
```

The skills are **project skills** — Claude Code loads them automatically in this
repo. You'll invoke them in the next steps as `/grill-me`, `/write-a-prd`, and
`/prd-to-issues` (or just by asking Claude to "use the … skill").

## What good looks like

- `http://localhost:3000` shows the SaltCRM UI and you can log in as Sam.
- The **Leads** list shows seeded leads, with Sofia Berg as `qualified`.
- `go test ./...` and `npm test` both pass; `npm run typecheck` exits cleanly.
- `.claude/skills/`, `ralph/`, and `issues/` all exist at the repo root.

## Common pitfalls

- **`go run` fails to build the SQLite driver** (errors mentioning `gcc`, `cc`,
  `cgo`, or `exec: "gcc": not found`). You're missing a C compiler or it's not on
  your `PATH`. Revisit *Before you start*: macOS `xcode-select --install`, Linux
  `build-essential`, Windows TDM-GCC/MSYS2. On Windows, open a **new** terminal
  after installing so the updated `PATH` takes effect, and check
  `go env CGO_ENABLED` is `1`.

- **Login returns 401, or the data looks stale/wrong.** Your `saltcrm.db` is out
  of date with the current schema. Stop the backend, delete the database, and
  restart to re-seed:

  **macOS / Linux:**
  ```bash
  rm backend/saltcrm.db
  cd backend && go run ./cmd/server
  ```

  **Windows (PowerShell):**
  ```powershell
  Remove-Item backend\saltcrm.db
  cd backend; go run ./cmd/server
  ```

- **Frontend can't reach the backend** ("friendly error" on the customer list).
  Make sure the backend is running on `:8080`. If you moved it, point the
  frontend at it:

  **macOS / Linux:**
  ```bash
  NEXT_PUBLIC_API_BASE_URL=http://localhost:8080 npm run dev
  ```

  **Windows (PowerShell):**
  ```powershell
  $env:NEXT_PUBLIC_API_BASE_URL="http://localhost:8080"; npm run dev
  ```

- **`npm install` is slow or fails.** Make sure you're on Node 20+; delete
  `frontend/node_modules` and retry if it's in a weird state.

- **`&&` gives an error on Windows.** You're on legacy PowerShell 5.1. Use
  PowerShell 7+ or Git Bash, or run each command on its own line.

## ✓ Checkpoint (done when…)

- [ ] Backend running on `:8080`, frontend on `:3000`.
- [ ] You logged into the UI as `sam@isp.example` and saw the seeded leads.
- [ ] `go test ./...`, `npm test`, and `npm run typecheck` all pass.
- [ ] `.claude/skills/`, `ralph/`, and `issues/` are present at the repo root.

---

← [Overview](00-overview.md) · Next → [Use the grill-me skill](02-use-the-grill-me-skill.md)
