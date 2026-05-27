# Write a PRD with the write-a-prd skill

**Goal:** Turn your grilled understanding of the conversion feature into a
structured Product Requirements Document an agent can act on.

**You'll produce:** `issues/prd.md` — problem, solution, user stories,
implementation decisions, testing decisions, and out-of-scope.

## Why this step

`/grill-me` got the decisions out of your head. A PRD writes them down in a shape
the rest of the pipeline consumes. The next step (`/prd-to-issues`) slices *this
document* into work items, and the agent in step 5 reads those slices. A vague
PRD produces vague issues produces an agent that builds the wrong thing. This is
where you make the feature concrete and complete.

## Before you start

- You've finished [step 2](02-use-the-grill-me-skill.md) and have a decision
  summary.
- The same context is still open (the brief + your grilling), so Claude can
  reuse it.

## Steps

### 1. Invoke the skill

```
/write-a-prd
```

The `write-a-prd` skill will:

1. Ask you for a detailed description of the problem and your ideas. Give it your
   decision summary from step 2.
2. Explore the repo to verify your assertions against the actual code (it'll find
   the existing `lead`, `customer`, `contact`, and `opportunity` domains).
3. Interview you further to close any remaining gaps.
4. Sketch the **modules** it expects to build or change, and check them with you
   — actively looking for **deep modules** (a lot of behavior behind a small,
   testable interface). Conversion is a natural one: "convert this lead into
   these records" is a simple interface hiding real logic.
5. Write the PRD to **`issues/prd.md`**.

### 2. Steer the module sketch

When it proposes modules, sanity-check them against what already exists. SaltCRM
already has separate `lead`, `customer` (the Account), `contact`, and
`opportunity` domains, plus `activity` and `metadata`/Studio for custom fields.
The conversion feature mostly **orchestrates** these — so expect a focused
conversion module/service with a clean interface, not a rewrite of each domain.
Tell Claude which modules you want **tests** written for (the conversion logic is
the obvious candidate).

### 3. Read the PRD critically

Open `issues/prd.md` and check it actually captures the brief. It should include:

- **Problem statement** — the handoff pain Marcus described.
- **Solution** — the Convert action and what it produces.
- **User stories** — a long list covering convert-from-qualified, the three
  records, field mapping, link-to-existing, optional opportunity, the
  `converted` status, can't-convert-twice, activity logging, custom-field
  carry-over, and visibility rules.
- **Implementation decisions** — schema change for `converted`, the conversion
  module's interface, how existing-account matching works. (No file paths or
  code — those go stale.)
- **Testing decisions** — what to test (external behavior, not internals) and
  prior art in the repo (e.g. the existing service-layer tests).
- **Out of scope** — no bulk/mass conversion.

### 4. Fix what's missing

If a decision from your grilling didn't make it in, ask Claude to add it. Iterate
until the PRD is something you'd be comfortable handing to another engineer.

## What good looks like

- `issues/prd.md` exists and reads like a complete spec, not a summary.
- The user-story list is **long** and covers every behavior in the brief,
  including the edge cases you pinned down in step 2.
- Out-of-scope explicitly names bulk conversion.
- Nothing in it surprises you — it's *your* decisions, written down.

## Common pitfalls

- **A thin PRD.** If the user-story list is five bullets, it's not done. The brief
  alone implies a dozen-plus stories.
- **Leaking code into the PRD.** Specific file paths and snippets rot fast; keep
  it at the decision/interface level.
- **Skipping testing decisions.** The agent uses `/tdd` in step 5 — it needs to
  know what "tested" means for this feature.
- **Forgetting it's local.** This `write-a-prd` writes a **file** to `issues/`. It
  does **not** open a GitHub issue.

## ✓ Checkpoint (done when…)

- [ ] `issues/prd.md` exists.
- [ ] It covers problem, solution, a long user-story list, implementation
      decisions, testing decisions, and out-of-scope.
- [ ] Every decision from your grilling is reflected in it.

---

← [Use the grill-me skill](02-use-the-grill-me-skill.md) · Next → [Turn a PRD to a kanban board with prd-to-issues](04-turn-a-prd-to-a-kanban-board-with-prd-to-issues.md)
