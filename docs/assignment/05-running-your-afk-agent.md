# Running your AFK agent

**Goal:** Let an agent build the AFK slices of the conversion feature, while you
drive the HITL ones — until the Convert action works end-to-end again.

**You'll produce:** The rebuilt, working lead-to-account conversion feature, with
the backend and frontend test suites green.

## Why this step

This is the payoff. You've turned a Slack message into a board of small,
well-specified slices. Now you run **`ralph/once.sh`** — a thin loop that hands
the agent your issues, the recent commits, and a working prompt, and lets it pick
one AFK task, build it test-first, run the feedback loops, commit, and move the
issue to `issues/done/`. You repeat until the board is clear. The HITL slices you
handle yourself, interactively.

## Before you start

- You've finished [step 4](04-turn-a-prd-to-a-kanban-board-with-prd-to-issues.md);
  `issues/` holds your slice files.
- Your green baseline from [step 1](01-setting-up-the-repo.md) still holds.
- **Commit your issues first.** The agent will make commits; you want a clean
  starting point so you can see exactly what it changed.

## How the runner works

`ralph/once.sh` (run from the repo root) does one iteration:

1. Gathers your `issues/*.md`, the last few git commits, and `ralph/prompt.md`.
2. Starts Claude with `--permission-mode acceptEdits` (it edits files without
   prompting — that's why you run on a throwaway feature branch).
3. The prompt tells the agent to: work **AFK issues only**, pick the next task
   (bugfixes → infra → tracer bullets → polish → refactors), build it with
   `/tdd`, run the **feedback loops**, commit with a descriptive message, and
   move the finished issue to `issues/done/`.
4. When there are no AFK tasks left, the agent outputs
   `<promise>NO MORE TASKS</promise>`.

The feedback loops the agent must pass before committing (from `ralph/prompt.md`):

```bash
cd backend && go build ./... && go test ./...     # backend
cd frontend && npm test && npm run typecheck       # frontend
```

> **No Docker needed.** This assignment uses `ralph/once.sh`. The repo also ships
> `ralph/afk.sh`, which loops many iterations inside a Docker sandbox — that's an
> optional, more advanced way to run the same prompt unattended. Skip it unless
> you've set up `docker sandbox`.

## Steps

### 1. Unblock the board: do the HITL slices first

The agent won't touch HITL slices, and your AFK slices are probably **blocked by**
them (e.g. "mark lead converted" depends on the `converted`-status schema change).
So start by driving the HITL slices yourself, interactively with Claude:

> Let's implement `issues/00X-add-converted-status.md` together. Use /tdd.

Work through each HITL slice this way, committing as you go, and move each to
`issues/done/` when it's done. Now the AFK slices that depended on them are
unblocked.

### 2. Run one iteration

From the repo root:

```bash
bash ralph/once.sh
```

**Watch it.** It will pick one AFK task, explore, build it test-first, run the
feedback loops, commit, and move the issue to `issues/done/`. Read the commit it
makes — that's your window into what it decided.

### 3. Repeat until the board is clear

`once.sh` does **one** task per run. Run it again for the next AFK slice:

```bash
bash ralph/once.sh
```

Keep going. You'll know you're done when the agent reports
`<promise>NO MORE TASKS</promise>` — there are no AFK issues left to grab. If
some AFK slice got blocked by a HITL one you haven't done yet, do that HITL slice
(step 1) and resume.

### 4. Verify against the acceptance checklist

Restart the backend and frontend, log in as `sam@isp.example`, and exercise the
feature on the seeded qualified lead **Sofia Berg (Polar Foods)**. Check each:

- [ ] A **Convert** action is available on a `qualified` lead — and **not** on a
      `new` lead.
- [ ] Converting creates an **Account** (Customer; company → account name,
      email/phone carried), a **Contact** (linked to the account), and an
      **Opportunity** in the **`prospecting`** stage.
- [ ] Field mapping just works — nothing is re-typed; the assigned **user/team**
      carries over.
- [ ] You can **link to an existing account** instead of creating a new one — no
      duplicate account.
- [ ] The **Opportunity is optional** (default on, can be skipped); Account +
      Contact are always created.
- [ ] The lead is marked **`converted`**, drops out of the active funnel, and
      **can't be converted twice**.
- [ ] The conversion is logged as an **activity**; the lead's existing
      notes/activities are related to the new records.
- [ ] **Studio custom fields** present on both lead and account carry across.
- [ ] No bulk/mass conversion exists; reps can only convert leads they can see.

And confirm the suites are still green:

```bash
cd backend && go test ./...
cd frontend && npm test && npm run typecheck
```

## The answer key (use sparingly)

This feature was removed from the repo on purpose. The full original
implementation lives in git history, **one commit before the removal**:

```bash
# See what the removal deleted (the feature, in reverse):
git show bc866b2

# Inspect the feature as it existed, e.g. the conversion module:
git show 70ea910:backend/internal/conversion/conversion.go
```

(`70ea910` is the last commit that still had the feature; `bc866b2` removed it.)

Reach for this only when you're genuinely stuck — and remember the agent is
*supposed* to arrive at a different, equally valid implementation than the
original. Don't diff against it; verify against the **acceptance checklist**
above.

## What good looks like

- `issues/done/` contains all your slices; `issues/` has no AFK work left.
- The Convert action works in the browser and satisfies the checklist.
- `go test ./...`, `npm test`, and `npm run typecheck` all pass.
- The git log shows a series of small, descriptive commits — one per slice.

## Common pitfalls

- **Agent says "NO MORE TASKS" but the feature isn't done.** Remaining slices are
  probably **HITL** (the agent skips those) or **blocked by** an undone HITL
  slice. Do the HITL work (step 1) and resume.
- **Feedback loops fail and the agent loops on the same task.** Read its commit
  notes — it records blockers for the next iteration. Often a HITL decision is
  missing, or a slice was mis-specified in step 4. Fix the issue file and rerun.
- **Backend changes seem to do nothing in the UI.** Your `saltcrm.db` may predate
  the new schema. Re-seed: `rm backend/saltcrm.db` and restart the backend.
- **You can't tell what the agent changed.** You forgot to commit before running.
  Commit your issues *before* the first `once.sh` so each agent commit stands
  alone.

## ✓ Checkpoint (done when…)

- [ ] Every slice is in `issues/done/`.
- [ ] The Convert feature satisfies the full acceptance checklist above.
- [ ] Backend and frontend suites are green.

🎉 You've rebuilt a real feature by directing an agent through a full
spec-to-ship pipeline. That's the job.

---

← [Turn a PRD to a kanban board with prd-to-issues](04-turn-a-prd-to-a-kanban-board-with-prd-to-issues.md) · [Back to overview](00-overview.md)
