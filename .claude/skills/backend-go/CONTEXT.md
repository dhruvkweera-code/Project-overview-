# Context File (session memory)

Lets a new session pick up a feature exactly where the last one stopped, without rewriting code.

## Where

`.claude/context/<feature-slug>.md` in the repo, e.g. `.claude/context/create-order-api.md`.
Commit it with the code. One file per feature.

## When to write

- **Create** it as soon as the first grill decision is settled.
- **Rewrite** it after every decision, after plan approval, and after every red → green slice.
- It is a **snapshot of now**, not a log. Overwrite old lines and keep it under 150 lines.
- Never put secrets or real customer data in it.

## Template

```markdown
# <Feature name>

**Status:** grilling | plan-approved | building (slice 3/6) | review | done
**Updated:** <YYYY-MM-DD HH:MM>
**Branch:** <git branch>

## Goal
<2 lines>

## Decisions
- Q1 contract: POST /orders → 201 {order}; 409 on duplicate idempotency key
- Q2 auth: caller must own account_id (tenant check in service)
- ...one line per settled answer

## Approach
<3–5 lines, the approved design + the rejected alternative>

## Plan (tests in build order)
- [x] TestCreateOrder_Valid
- [x] TestCreateOrder_InvalidBody
- [ ] TestCreateOrder_NotOwner
- [ ] TestCreateOrder_DuplicateKey

## File map
| File | Job | Lines |
|---|---|---|
| internal/order/handler.go | decode/validate/map errors | 74 |
| internal/order/service.go | logic + tx | 96 |

## Next step
<one exact action, e.g. "Write failing TestCreateOrder_NotOwner in internal/order/service_test.go">

## Open questions / blockers
- <anything waiting on the user>
```

## Resume procedure

1. Read the file.
2. Run `go build ./... && go test ./...`. If the code disagrees with the checkboxes, **the code wins**:
   fix the checkboxes and say so.
3. Tell the user: feature · status · next step (≤ 3 lines).
4. Continue from **Next step**. Do not re-grill settled decisions or rewrite passing code.
