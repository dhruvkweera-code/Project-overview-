# TDD Plan (show this, then wait for approval)

Keep it scannable. No code beyond signatures and SQL shapes.

````markdown
## Plan: <feature>

### 1. Goal
<2 lines: what the user gets>

### 2. Best approach - and why
<the design in 3–6 lines>
- **Why:** <reason tied to the user's answers: speed, safety, simplicity>
- **Rejected:** <alternative> - <why not>

### 3. What the code does
Request flow, in plain words:
```
POST /orders
 └─ handler: decode JSON → validate → check caller owns account → call service
     └─ service.CreateOrder: begin tx → check idempotency key → insert order + items → commit
         └─ store (sqlc): InsertOrder, InsertOrderItems (one batch query, no loop)
 ← 201 {order} | 400 validation | 403 not owner | 409 duplicate key
```
Error cases: <each one and what the caller sees>

### 4. Seams (where tests live)
| Seam | Catches | Misses |
|---|---|---|
| HTTP via `httptest` | status codes, JSON shape, auth | nothing deep in SQL |
| Service + real Postgres (testcontainers) | business rules, tx, constraints | HTTP mapping |

### 5. Test list (build order = red → green order)
1. `TestCreateOrder_Valid` - happy path returns the order with an ID
2. `TestCreateOrder_InvalidBody` - 400 with the field error
3. `TestCreateOrder_NotOwner` - 403, nothing written
4. `TestCreateOrder_DuplicateKey` - second call returns the first order, no double insert
...each line: name - why it matters

### 6. File map (budget: 150 lines/file, tests 250)
| File | Job | Est. lines |
|---|---|---|
| `internal/order/handler.go` | HTTP decode/validate/map errors | ~80 |
| `internal/order/service.go` | business logic + tx | ~100 |
| `internal/order/errors.go` | domain errors | ~20 |
| `db/queries/order.sql` | sqlc queries | ~30 |
| `db/migrations/00X_orders.sql` | schema + indexes | ~25 |
| `internal/order/*_test.go` | tests per seam | ~200 |

### 7. Risks (from CHECKLIST)
- <e.g. duplicate submits → unique (account_id, idempotency_key)>
- <e.g. list endpoint → index on (account_id, created_at DESC), cursor pagination>

### 8. Assumptions (veto any)
- <anything not settled in the grill>

**Reply "go" to start, or tell me what to change.**
````

## Rules for the plan

- **Approach section is mandatory**: always name at least one rejected alternative.
- **"What the code does" is written for a human reader**: plain words, the request path,
  every error outcome. This is the explanation the user reads before any code exists.
- Prefer the **highest seam** (HTTP or service). Fewer seams = fewer, better tests.
  No tests on private functions.
- Use **real Postgres** in tests. Only external third parties get fakes.
- Order tests so each one builds on the last (tracer bullets). Happy path first,
  then validation, authZ, edge cases, concurrency.
- Any file estimated > 120 lines: plan its split up front.
- Copy the approved plan into the context file with checkboxes.
