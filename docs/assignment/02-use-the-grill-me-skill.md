# Use the grill-me skill

**Goal:** Turn Marcus's vague Slack message into a plan you actually understand
and can defend — by having Claude interrogate you about it.

**You'll produce:** A clear mental (and written) model of the conversion
feature: the decisions, the edge cases, and the open questions — ready to feed
into a PRD.

## Why this step

A client brief is full of hidden decisions. "Link to an existing account
instead of creating a duplicate" sounds simple until you ask: matched on what —
company name? domain? What if there are two matches? The `grill-me` skill makes
Claude **interview you relentlessly**, one question at a time, walking every
branch of the decision tree. You can't write a good PRD for something you
haven't been forced to think through. This is that forcing function.

## Before you start

- You've completed [step 1](01-setting-up-the-repo.md) and the app runs.
- Read [`client-brief.md`](../../client-brief.md) once, slowly. It's Marcus's
  full request — Convert action, the three records, field mapping, link-to-
  existing, optional opportunity, the `converted` status, activity logging,
  custom fields, and the explicit out-of-scope (no bulk conversion).

## Steps

### 1. Open the brief in context

Tell Claude what you're working on so the grilling is grounded:

> Read `client-brief.md`. I want to build the lead-to-account conversion feature
> it describes. Use the grill-me skill on my plan.

### 2. Invoke the skill

```
/grill-me
```

Claude will start asking you questions **one at a time**, each with a
recommended answer, exploring the codebase instead of asking when it can.

### 3. Answer honestly — and push back

Don't rubber-stamp the recommendations. The point is to *reach a decision*, not
to agree fast. Expect to resolve things like:

- **Convert gating** — only `qualified` leads? What happens if someone tries to
  convert a `new` or already-`converted` lead?
- **The three records** — Account is a Customer record, Contact links to it,
  Opportunity starts in `prospecting`. Which fields carry over to each?
- **Link to existing account** — how is a match detected, and who decides
  (auto-match vs the rep picks)? How do you avoid duplicates?
- **Optional opportunity** — default on, but skippable. What's the UI for that?
- **The `converted` terminal status** — leads today only go to
  `qualified`/`unqualified`. Adding `converted` is a schema change. A converted
  lead must drop out of the active funnel and **can't be converted twice**.
- **History** — log the conversion as an activity; relate the lead's existing
  notes/activities to the new records.
- **Studio custom fields** — carry values across when a field exists on both the
  lead and the account.
- **Guardrails** — no bulk conversion; respect record-level visibility (reps
  only convert leads they can see).

### 4. Capture the outcome

When the grilling reaches a shared understanding, ask Claude to summarize the
resolved decisions and any remaining open questions. Keep that summary — you'll
hand it straight to `/write-a-prd` in the next step.

> Summarize everything we just resolved as a decision list, and flag anything
> still open.

## What good looks like

- You can explain, without re-reading the brief, what Convert does, what it
  creates, and what it refuses to do.
- Every "it should just work" phrase in the brief has become a concrete rule
  (e.g. "match an existing account by exact company name; if multiple match, the
  rep chooses").
- You have a short written decision list to carry into the PRD.

## Common pitfalls

- **Agreeing too quickly.** If you accept every recommendation, you've learned
  nothing. Disagree when you actually disagree.
- **Designing the code.** This step is about the *feature's behavior*, not file
  layout or function signatures. Save implementation for later.
- **Skipping the edge cases.** Double-conversion, no-match vs multi-match,
  opportunity-skipped — these are exactly where the feature breaks. Pin them
  down now.

## ✓ Checkpoint (done when…)

- [ ] You ran `/grill-me` and worked through the conversion plan to a shared
      understanding.
- [ ] Every major decision from the brief is resolved or explicitly flagged as
      open.
- [ ] You have a written decision summary ready to feed into the PRD.

---

← [Setting up the repo](01-setting-up-the-repo.md) · Next → [Write a PRD with the write-a-prd skill](03-write-a-prd-with-the-write-a-prd-skill.md)
