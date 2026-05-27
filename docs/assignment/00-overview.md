# Assignment: Rebuild lead-to-account conversion, the agentic way

Welcome. In this assignment you'll act as an AI engineer on **SaltCRM** and
ship a real feature using an agentic coding pipeline — the same one taught in
the [AI Hero AI Engineer Workshop 2026](https://www.aihero.dev/ai-engineer-workshop-2026~dwnll).

You won't hand-write the feature. You'll **direct an agent** to build it, by
turning a vague client request into a sharp plan, the plan into a spec, the spec
into bite-sized issues, and then letting the agent work through them.

## The scenario

Marcus Webb (VP Sales) sent the team a Slack message asking for a
**lead-to-account conversion** workflow — one "Convert" button on a qualified
lead that spins up an Account, a Contact, and an Opportunity in one shot. His
full message is in [`client-brief.md`](../../client-brief.md) at the repo root.
That message is your **only** input. Everything else you produce.

This feature *used* to exist in SaltCRM and was deliberately removed so you can
rebuild it. That means there's a built-in answer key (see
[step 5](05-running-your-afk-agent.md)) — but try not to peek.

## The pipeline

Each step has its own page. Do them in order; each one's output is the next
one's input.

```
client-brief.md
      │
      ▼
┌─────────────────┐   stress-test your understanding of the request
│  /grill-me      │   → a plan you can defend                 (step 2)
└─────────────────┘
      │
      ▼
┌─────────────────┐   turn the request into a structured spec
│  /write-a-prd   │   → issues/prd.md                         (step 3)
└─────────────────┘
      │
      ▼
┌─────────────────┐   slice the PRD into independently-workable issues
│  /prd-to-issues │   → issues/NNN-*.md  (your "kanban board") (step 4)
└─────────────────┘
      │
      ▼
┌─────────────────┐   let the agent build the AFK slices; you drive the HITL ones
│  ralph/once.sh  │   → working feature, tests green          (step 5)
└─────────────────┘
      │
      ▼
   verify against the acceptance checklist
      │
      ▼
┌──────────────────────────────┐  deepen a shallow module for testability
│ /improve-codebase-architecture│  → a refactor RFC issue       (step 6, capstone)
└──────────────────────────────┘
```

## What you'll learn

- How to **stress-test a plan** with `/grill-me` before writing a line of spec.
- How to produce a **PRD** that an agent can actually act on.
- How to break a feature into **tracer-bullet vertical slices** and decide which
  are safe to hand off (**AFK**, away-from-keyboard) versus which need you in the
  loop (**HITL**, human-in-the-loop).
- How to **run an unattended-ish agent** (`ralph/once.sh`) that picks a task,
  builds it test-first, runs the feedback loops, and commits.
- How to **find and plan an architecture improvement** — deepening a shallow
  module for testability — and turn it back into a pipeline issue.

## Steps

| # | Page | You'll produce | Est. |
|---|------|----------------|------|
| 1 | [Setting up the repo](01-setting-up-the-repo.md) | A running SaltCRM (backend + frontend) and verified tooling | 20–30 min |
| 2 | [Use the grill-me skill](02-use-the-grill-me-skill.md) | A defensible plan for the conversion feature | 20–30 min |
| 3 | [Write a PRD with the write-a-prd skill](03-write-a-prd-with-the-write-a-prd-skill.md) | `issues/prd.md` | 20–30 min |
| 4 | [Turn a PRD to a kanban board with prd-to-issues](04-turn-a-prd-to-a-kanban-board-with-prd-to-issues.md) | `issues/NNN-*.md` issue files | 30–40 min |
| 5 | [Running your AFK agent](05-running-your-afk-agent.md) | The rebuilt, working feature | 60–120 min |
| 6 | [Improve the codebase architecture](06-improve-codebase-architecture.md) *(capstone)* | A refactor RFC issue | 30–45 min |

## Prerequisites (summary — details in step 1)

- **Go 1.26+** and a C compiler (for the SQLite driver)
- **Node.js 20+**
- **Git**
- **Claude Code** (the CLI you're reading this with)
- A terminal you're comfortable in

You do **not** need Docker for this assignment — we use the no-sandbox runner.

## The finish line

You're done when the **Convert** action works end-to-end again and the feature
satisfies the acceptance checklist in [step 5](05-running-your-afk-agent.md):
a qualified lead can be converted into an Account + Contact + Opportunity, the
lead becomes `converted` and can't be converted twice, and the backend and
frontend test suites are green.

Then, as a **capstone**, [step 6](06-improve-codebase-architecture.md) has you
improve the architecture of what you built — finding a shallow module and
planning a deeper one as a refactor RFC.

→ Start with [**Setting up the repo**](01-setting-up-the-repo.md).
