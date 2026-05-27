# Turn a PRD to a kanban board with prd-to-issues

**Goal:** Slice your PRD into small, independently-workable issues — and decide
which the agent can build alone (**AFK**) versus which need you (**HITL**).

**You'll produce:** A set of `issues/NNN-*.md` files (your kanban board), each a
thin vertical slice with acceptance criteria and dependencies.

## Why this step

An agent works best on **small, complete, verifiable** units of work — not "build
the whole feature." The `prd-to-issues` skill breaks the PRD into **tracer-bullet
vertical slices**: each one cuts through every layer (schema → API → UI → tests)
and is demoable on its own. The `issues/` folder is your board; `issues/done/` is
the "done" column. In [step 5](05-running-your-afk-agent.md), the agent reads
this board and works the slices one at a time.

## Before you start

- You've finished [step 3](03-write-a-prd-with-the-write-a-prd-skill.md) and
  `issues/prd.md` exists.
- Skim the existing `issues/` folder — issues are numbered, so the skill picks up
  from the next free number.

## HITL vs AFK — the key decision on this page

Every slice is tagged **HITL** or **AFK**:

- **AFK (away-from-keyboard)** — the agent can implement and verify it without
  you. These are slices where the behavior is well-specified and there's a clear
  test for "done."
- **HITL (human-in-the-loop)** — the slice needs a human: an architectural call,
  a schema decision, or a UX design review.

The agent in step 5 works **AFK slices only**. You drive the HITL ones yourself
(interactively with Claude). Prefer AFK where you honestly can — but don't
mislabel a real decision as AFK just to offload it.

For lead-conversion, a realistic split looks like:

| Likely **HITL** | Likely **AFK** |
|---|---|
| Add the `converted` terminal status + the schema/migration change (it changes the funnel model) | Carry lead fields → Account/Contact (name, email, phone, company) |
| The Convert modal UX: create-new vs link-to-existing account, and the optional-opportunity toggle | Create the Opportunity in `prospecting` stage |
| | Mark the lead `converted`; block converting a non-`qualified` or already-`converted` lead (can't-convert-twice guard) |
| | Carry the assigned user/team to the new records |
| | Log the conversion as an activity; relate the lead's existing notes/activities |
| | Carry matching Studio custom fields across |

(Your split may differ — this is a guide, not the answer.)

## Steps

### 1. Invoke the skill

```
/prd-to-issues
```

Point it at your PRD when asked: `issues/prd.md`.

### 2. Review the proposed slices

The skill presents a numbered breakdown. For each slice it shows the **title**,
**type** (HITL/AFK), **blocked-by**, and which **user stories** it covers. Quiz
yourself with the skill's own questions:

- Is the **granularity** right — many thin slices, not a few thick ones?
- Are the **dependencies** correct? (The `converted`-status schema slice almost
  certainly blocks the "mark lead converted" slice.)
- Should any slice be **split or merged**?
- Are the **HITL/AFK tags** honest?

Iterate until you're happy. Aim for a **tracer bullet first**: the thinnest
end-to-end Convert that creates *something* and proves the wiring, before the
slices that flesh out field mapping, opportunities, and history.

### 3. Generate the files

Once approved, the skill writes one markdown file per slice to
`issues/NNN-short-title.md`, in dependency order, cross-referencing real
filenames in each "Blocked by" field. Each file has: parent PRD, what to build,
acceptance criteria, blocked-by, and the user stories it addresses.

### 4. Eyeball the board

```bash
ls issues
```

Read a couple of the generated files. Each should be buildable on its own and
have a checkable definition of done.

## What good looks like

- `issues/` holds several `NNN-*.md` files, each a thin vertical slice.
- There's a clear **first** slice (a tracer bullet) with no blockers.
- HITL and AFK tags are honest, and dependencies point at real filenames.
- The slices, taken together, cover the whole PRD.

## Common pitfalls

- **Horizontal slices.** "Do all the schema," then "do all the API," then "do all
  the UI" is the wrong cut — nothing is demoable until the end. Slice
  *vertically*: one small behavior, all the way through.
- **Everything marked AFK.** If you tag the schema change and the modal UX as AFK
  to make the agent do everything, you'll get a confused agent and a feature you
  didn't design. Keep the genuine decisions HITL.
- **Slices too big.** "Implement conversion" is not a slice. "Convert a qualified
  lead into an Account (no contact/opportunity yet), behind the button" is.
- **GitHub confusion.** This skill writes **local files**. It does not create
  GitHub issues or reference issue numbers — only filenames.

## ✓ Checkpoint (done when…)

- [ ] `issues/` contains numbered slice files covering the whole PRD.
- [ ] At least one AFK tracer-bullet slice has no blockers and can start
      immediately.
- [ ] HITL/AFK tags and "blocked by" references are accurate.

---

← [Write a PRD with the write-a-prd skill](03-write-a-prd-with-the-write-a-prd-skill.md) · Next → [Running your AFK agent](05-running-your-afk-agent.md)
