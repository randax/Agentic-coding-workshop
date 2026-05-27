# Slack Message from Marcus Webb (VP Sales, SaltCRM)

#product-requests · today

Hey team,

Our reps are losing deals in the handoff. When a lead finally gets **qualified**, there's no clean way to turn it into a real account — they end up **re-typing the same name, email, phone, and company** into a new customer record, then a contact, then an opportunity. It's slow, it's error-prone, and half the time the opportunity never gets created at all, so the deal just **falls out of the pipeline**.

I want a proper **lead-to-account conversion workflow**, like the "Convert" action in SugarCRM. One button on a qualified lead that **spins up the whole account in one shot** and automates the rest of the sales process.

Here's what I'm picturing:

- A **Convert** action on a lead. **Only available once the lead is `qualified`** — you shouldn't be able to convert a raw `new` lead.
- Converting **creates three linked records in one step**:
  - an **Account** (a Customer record) — `Company` becomes the account name, email/phone carry over
  - a **Contact** — the lead's name/email/phone, **linked to the new account**
  - an **Opportunity** — tied to the account, starting in the **`prospecting`** stage
- **Field mapping should just work** — don't make the rep re-enter anything the lead already has. Carry over the **assigned user / team** too, so ownership stays with the same rep.
- Like SugarCRM, let the rep **link to an existing account** instead of always creating a new one — if the company is already a customer, the new contact and opportunity should attach to that account rather than creating a **duplicate**.
- The **Opportunity should be optional** — default it on, but let the rep skip it (sometimes it's just a contact, not a deal yet). Account + Contact are always created.
- Once converted, mark the lead **`converted`** (we'll need a new terminal status — today leads only go up to `qualified`/`unqualified`). A converted lead should **drop out of the active funnel** and **can't be converted twice**.
- Leave a trail: log the conversion as an **activity** so we can see when a lead became an account and who did it. If the lead has existing notes/activities, **relate them to the new records** so history isn't lost.
- Respect **Studio custom fields** — if a field exists on both the lead and the account, carry the value across.

**No bulk/mass conversion for now** — one lead at a time is fine. And keep it inside our existing record-level visibility rules (reps only convert leads they can see).

This is the biggest paper-cut in the sales workflow right now. Can someone scope it? I'd love it this quarter.

Marcus
