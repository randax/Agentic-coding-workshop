# Improve the codebase architecture

**Goal:** Now that conversion works, make the code behind it *deeper* — a smaller
interface hiding more behavior — so it's more testable and easier for the next
agent to navigate.

**You'll produce:** A refactor RFC written as an `issues/NNN-*.md` file: a chosen
candidate, a designed interface, and the trade-offs behind it.

## Why this step

Shipping a feature and shipping *good* architecture are different skills. The
conversion feature you rebuilt orchestrates several domains (lead → account,
contact, opportunity, activity, custom fields). Orchestration code tends to
sprawl: logic leaks into handlers, modules stay shallow (their interface is
nearly as complex as their implementation), and the real bugs hide in the seams.

A **deep module** (John Ousterhout, *A Philosophy of Software Design*) has a
small interface hiding a large implementation. Deep modules are more testable —
you test at the boundary instead of inside — and more AI-navigable, because
understanding one concept doesn't mean bouncing between ten tiny files. The
`improve-codebase-architecture` skill finds these deepening opportunities and
turns one into a concrete refactor plan.

This is a capstone: it **loops back into the pipeline**. The RFC you produce is
just another issue — you can later run it through the AFK agent or `/tdd` exactly
like the slices in steps 4–5.

## Before you start

- You've finished [step 5](05-running-your-afk-agent.md): conversion works and the
  test suites are green. A green baseline matters — a refactor must preserve
  behavior, and your tests are how you'll know it did.
- Commit your work first, so the RFC (and any later refactor) starts from a clean
  tree.
- Note: this skill's `REFERENCE.md` is a stub in this repo, so it will use its own
  judgment for the dependency categories and RFC shape. That's fine — the value
  is in the exploration and the interface designs.

## Steps

### 1. Invoke the skill

```
/improve-codebase-architecture
```

Tell it where to look:

> Focus on the lead-to-account conversion feature I just rebuilt and the modules
> it touches.

### 2. Read the candidates it surfaces

The skill explores the code the way an agent would and presents a numbered list
of **deepening opportunities**. For each, it shows the **cluster** of modules
involved, **why they're coupled**, a **dependency category**, and the **test
impact** (which existing tests would become boundary tests). It does *not* propose
solutions yet.

Look for friction you recognize from the rebuild — e.g. conversion logic spread
across a handler and several services, or a module so thin it's just plumbing.

### 3. Pick one candidate

Choose the opportunity that would most improve testability. For this feature, a
conversion module that exposes something like "convert this lead into these
records" — hiding the field mapping, existing-account matching, opportunity
creation, and activity logging behind one interface — is a strong candidate.

### 4. Frame the problem, then compare interface designs

The skill writes up the **problem space** (constraints, dependencies, a rough
sketch to ground them), then spawns **several sub-agents in parallel**, each
designing a *radically different* interface for the deepened module — e.g. one
minimal (1–3 entry points), one maximally flexible, one optimized for the common
caller, one using ports & adapters. It presents each (signature, usage example,
what it hides, dependency strategy, trade-offs), compares them, and gives an
**opinionated recommendation**.

Read the designs against this codebase's rules (see the root `CLAUDE.md`):
business logic belongs in a domain Service behind a `Repository` interface; the
`api` layer stays HTTP-only; `store` is the only GORM consumer. The best design
deepens the conversion module *without* breaking that dependency rule.

### 5. Pick an interface and let it write the RFC

Accept the recommendation or pick another design. The skill writes the refactor
RFC to `issues/NNN-*.md` and shares the path.

### 6. (Optional) feed it back through the pipeline

The RFC is now a normal issue. If you want to actually do the refactor, tag it
AFK and run [step 5](05-running-your-afk-agent.md)'s `ralph/once.sh`, or implement
it interactively with `/tdd`. The test suites staying green is your proof that
behavior was preserved.

## What good looks like

- `issues/` has a new RFC file describing a deeper module for the conversion
  seam: the chosen interface, what it hides, how dependencies are handled, and
  which tests move to the boundary.
- The proposed interface respects this repo's dependency rule (Service →
  `Repository` interface; HTTP-only `api`; GORM only in `store`).
- You can articulate *why* the chosen design is deeper than what exists today.

## Common pitfalls

- **Skipping exploration.** The friction the skill encounters *is* the signal.
  Don't jump to a solution before you've felt where the code is awkward.
- **Refactoring for aesthetics.** The goal is testability and a smaller
  interface, not moving files around. If a change doesn't shrink the interface or
  improve the seam, it's not this exercise.
- **Picking a non-shallow candidate.** A module that's already deep doesn't need
  deepening. Choose where the interface is nearly as complex as the
  implementation.
- **Treating the RFC as the build.** This step stops at the *plan*. Executing it
  is optional (step 6 above) — don't start hand-coding mid-skill.
- **Breaking the dependency rule.** A "clever" interface that pulls GORM into a
  service or business logic into `api` is a regression, not an improvement.

## ✓ Checkpoint (done when…)

- [ ] The skill surfaced candidates and you picked one tied to the conversion
      feature.
- [ ] You compared the parallel interface designs and chose one.
- [ ] A refactor RFC exists as `issues/NNN-*.md`, consistent with the repo's
      architecture rules.

---

← [Running your AFK agent](05-running-your-afk-agent.md) · [Back to overview](00-overview.md)
