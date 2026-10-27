# Grill

Interview the user until the feature is clear enough to plan. Think of it as a **design tree**:
each decision unlocks the decisions that hang off it.

## Rules

- **Max 2 rounds, ≤ 6 questions per round.** Leftovers become **Assumptions** in the plan,
  which the user can veto.
- Each round asks the **frontier**: only questions whose prerequisites are already settled.
  A question that depends on another open question waits for the next round.
- Every question has **my recommended answer**. Phrase it so "yes" accepts the recommendation.
- **Facts are my job, decisions are the user's.** Never ask what I can find myself:
  read the repo (go.mod, existing handlers, `sqlc.yaml`, migrations, `GLOSSARY.md`, ADRs) first.
- After each round, write settled decisions into the context file immediately.
- Skip questions the repo already answers. Say "I see you use X, so I'll follow it."

## Format

```
❓ **Q1 - <title>**: <question, with options a/b/c if useful>

➡️ <recommended answer + one-line why>

---
```

End with: `Reply like "Q2 b, Q4 no". Anything you skip = yes.`

## Question bank (pick what applies, don't ask all)

**Contract**
- Method + path, request body, response shape, status codes?
- Who can call it? Auth method, and which **authorization** rule (owner / tenant / role)?
- Pagination for lists: cursor (recommended) or offset? Default/max page size?

**Data**
- New tables/columns? Constraints (unique, FK, not null, check)?
- Which queries run on it, and so which indexes are needed?
- Soft delete or hard delete? Timestamps (`created_at`, `updated_at`)?
- Money/time types: `numeric`/int cents, `timestamptz` (recommended).

**Behaviour & edge cases**
- What happens on duplicate submit? (recommend: idempotency key or unique constraint)
- Concurrent updates to the same row? (recommend: optimistic `version` column or `SELECT ... FOR UPDATE`)
- What must happen atomically, i.e. in one transaction?
- Validation rules and their error messages?

**Integrations**
- External calls (payment, email, S3...)? Timeout and retry policy? Sync or async job?

**Non-functional**
- Expected load / latency target? Any rate limiting?
- What must be logged or measured? Any PII that must never be logged?

## Stress-test with scenarios

Invent 1–2 concrete edge cases per round to force precision, e.g.
"User submits the same order twice within 100 ms on a flaky network. One order or two?"

## Domain words

If a term is fuzzy ("account": user or customer?), ask which one is meant and add it to
`GLOSSARY.md` (no implementation details there). Offer an ADR in `docs/adr/` only if the
decision is hard to reverse, surprising, **and** a real trade-off.

## Exit

The grill is done when the frontier is empty or 2 rounds are used. Then go to Step 3 (plan).
