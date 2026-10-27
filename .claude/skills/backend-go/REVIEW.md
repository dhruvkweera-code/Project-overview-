# Review + Ship

## 1. Two-axis review

Diff against the base: `git diff <base>...HEAD`. Keep the two axes **separate**, so clean
style can't hide wrong behaviour (or the reverse).

**Spec axis** (against the approved plan in the context file):
- Planned tests or behaviour that are missing or partial
- Behaviour nobody asked for (scope creep)
- Implemented, but wrong

**Standards axis** - repo standards first (they win), then [CHECKLIST.md](CHECKLIST.md), then these
smells. Smells are judgement calls, not hard rules:
- **Mysterious Name** → rename. If no honest name comes, the design is murky.
- **Duplicated Code** → extract once.
- **Speculative Generality** → delete unused params, interfaces and options.
- **Middle Man** (pass-through func/struct) → call the real target.
- **Feature Envy** → move the logic next to the data it uses.
- **Data Clumps / Primitive Obsession** → a small type (`type OrderID int64`, `Money`).
- **Repeated Switches** → one map or one interface.
- **Shotgun Surgery / Divergent Change** → group what changes together.
- **Long file / long func** → over 150 / 40 lines means split.

Skip anything `gofmt`, `go vet` or `golangci-lint` already enforce.

## 2. Optimisation pass (only while all tests are green)

- Remove dead code. Collapse needless branches. Inline single-use helpers.
- Replace query loops with batch queries.
- Re-run `go test -race ./...` after the pass.

## 3. Explain the final code

Before the PR, give the user a short walkthrough: each file → its job → what happens on a request.
Plain words, ≤ 10 lines.

## 4. PR body

```markdown
## Summary
<smallest visual that makes it clear: call tree, file-tree diff, or mermaid sequence>

## Evidence
- **Before:** <failing test output / curl / timing>
  **After:** <passing test output / curl / timing>

## Merge Danger
**Door:** one-way (migration, data change, public API) | two-way (easy revert)
**Blast Radius:** <one word: endpoint / service / all-tenants>
```

No preamble. Brief prose. Use the domain words from `GLOSSARY.md`.

## 5. Close out

- Commit with a message that states what changed and, for bugs, the root cause.
- Set the context file `Status: done`.
- Offer a retro line: one thing in the repo (lint rule, script, test helper) that would have
  made this session faster.
