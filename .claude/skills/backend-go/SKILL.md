---
name: backend-go
description: Daily Go + Postgres backend workflow. Use for ANY coding request (new API, endpoint, feature, schema/migration, job, bug fix, refactor, performance) and when the user says continue/resume. Never writes code first - grills the user, shares the best approach and explains what the code will do, presents a TDD plan, waits for approval, then builds test-first in small files (max 150 lines) while keeping a resumable context file in .claude/context/.
---

# Backend Go

Stack: **Go** · stdlib `net/http` (1.22+ routing) · **Postgres** via `pgx` + `sqlc` · `testing` + `httptest` + testcontainers-go.
If the repo already uses something else (chi, gin, GORM, migrate tool...), follow the repo.

**Prime directive:** correct, fast, readable, small. Line count is never a goal. Every line earns its place.

## Hard rules (always on)

1. **No code before approval.** Approach + explanation + plan first, code second.
2. **150-line cap per file** (test files: 250). Functions ≤ 40 lines. Lines ≤ 120 chars.
   About to exceed? Split by responsibility (handler / service / store / types), never mid-logic.
3. **Test first.** One failing test → minimum code → green. One slice at a time.
4. **Context file is the memory.** `.claude/context/<feature>.md` updated after every slice and decision.
5. **Code is truth.** On resume, run the build and tests before trusting the context file. Never rewrite passing code.

## Step 0: Resume check (every session start)

If the user says continue/resume, or `.claude/context/` has a file whose status is not `done`:
1. Read it (template: [CONTEXT.md](CONTEXT.md)).
2. Run `go build ./... && go test ./...` and compare with the checkboxes.
3. Tell the user in ≤3 lines: feature, status, next step.
4. Continue from **Next step**. Skip Steps 1–3 if the plan is already approved.

## Step 1: Size the request

| Request | Flow |
|---|---|
| New endpoint / feature / schema change / refactor / bug with unclear cause | Full: Steps 2 → 3 → 4 → 5 |
| Small edit (< ~20 lines, rename, typo, obvious one-liner) | **Mini-approach** below, then build |
| User says "just do it" | Mini-approach, no waiting |
| Bug / panic / wrong result / slow | Debug loop in [CHECKLIST.md](CHECKLIST.md) §Debug, then Step 3 for the fix |

**Mini-approach** (even for small edits, before any code):
```
Approach: <what I'll change and where>
What it does: <behaviour in plain words>
Why this way: <one line, alternative rejected>
OK?
```

## Step 2: Grill

Follow [GRILL.md](GRILL.md). Max 2 rounds, ≤ 6 questions each, every question with a recommended answer.
Create the context file as soon as the first decision is settled.

## Step 3: Approach + explanation + TDD plan → wait

Present the plan using [TDD-PLAN.md](TDD-PLAN.md). It must include:
- **Best approach** and why, with alternatives rejected.
- **What the code does**: request → handler → service → store → DB, in plain words, plus error cases.
- Seams, ordered test list, file map with line budgets, risks.

Then **stop and wait** for "ok / yes / go". Edits to the plan → update it and the context file, then ask again.

## Step 4: Build, slice by slice

For each test in the plan, in order:
1. Write the failing test. Run it: `go test ./<pkg> -run <TestName>`. Show it is **red** for the right reason.
2. Write the minimum code to pass. No speculative params, flags or interfaces.
3. Run it again → **green**. Then `go vet ./... && go test -race ./<pkg>`.
4. Check the line counts of touched files: `wc -l <files>`. Over the cap → split now.
5. Tick the test in the context file and rewrite **Next step**.
6. After each slice, explain it in 2–3 lines: what was added, and what that code does.

Run the full suite + `golangci-lint run` once at the end.

## Step 5: Review + ship

Follow [REVIEW.md](REVIEW.md): two-axis review (Spec / Standards + [CHECKLIST.md](CHECKLIST.md)),
optimisation pass, PR body. Set the context status to `done`.

## Design rules (short)

- **Deep modules**: a small interface hiding real behaviour. Delete pass-through layers.
- Handler = decode + validate + authZ + map errors to HTTP. **No business logic.**
- Service = the logic. Takes deps as struct fields or args, returns `(T, error)`.
- Store = sqlc queries. **No business rules.**
- **Accept interfaces, return structs.** Define interfaces at the *consumer*, and only when 2 implementations exist (prod + test fake).
- `context.Context` is the first param of everything doing I/O.
- Wrap errors with context: `fmt.Errorf("create order: %w", err)`. Sentinel/typed errors for domain cases.

## Done means

- [ ] All plan tests green, `go test -race ./...`, `go vet`, `golangci-lint` clean
- [ ] Every file ≤ 150 lines (tests ≤ 250)
- [ ] CHECKLIST passes for touched code
- [ ] Context file status `done`, PR body written

_Built from mattpocock/skills (engineering): grilling, to-spec, tdd, codebase-design,
diagnosing-bugs, code-review, pr, handoff, domain-modeling. Adapted for Go + Postgres._
