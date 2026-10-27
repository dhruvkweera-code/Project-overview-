# Go + Postgres Checklist

Use while building and in review. Every item is a common production failure.

## HTTP & input
- Decode with `json.NewDecoder(r.Body)`, `DisallowUnknownFields()`, and a size cap via `http.MaxBytesReader`.
- Validate at the handler; inside the service, trust the types.
- Map domain errors → HTTP status in **one** function (`writeError`). Never leak SQL or stack traces.
- Server timeouts set: `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`.
- **AuthZ on every resource** (owner/tenant check), not just authN.

## Errors
- Wrap: `fmt.Errorf("op: %w", err)`. Check with `errors.Is` / `errors.As`.
- `pgx.ErrNoRows` → domain `ErrNotFound`. Unique violation (`23505`) → `ErrConflict`.
- No `_ =` on errors that matter. No `panic` in request paths. Log once, where handled.

## Database (pgx + sqlc)
- **No N+1**: never run a query inside a loop. Use `= ANY($1)`, joins, or `pgx.Batch` / `CopyFrom`.
- Select only the needed columns. Paginate every list, cursor style:
  `WHERE (created_at, id) < ($1, $2) ORDER BY created_at DESC, id DESC LIMIT $3`.
- Add an index for every new `WHERE` / `JOIN` / `ORDER BY` shape. Verify with `EXPLAIN ANALYZE`.
- Multi-write work runs in **one transaction**: `pgx.BeginFunc` or `defer tx.Rollback(ctx)` + `Commit`.
  Keep it short, with no HTTP calls inside.
- Constraints in the schema (unique, FK, not null, check): the DB is the last line of defence.
- Use `timestamptz` for times, `numeric` or int cents for money (never float), `uuid`/`bigint` for IDs.
- Migrations: backward compatible (expand → backfill → contract). Index big tables with
  `CREATE INDEX CONCURRENTLY`.
- One shared `pgxpool.Pool`, created at startup. Never open a connection per request.

## Concurrency & reliability
- Pass `ctx` through everything. Respect cancellation. Set `context.WithTimeout` on outbound calls.
- Mutating endpoints are **idempotent** (idempotency key + unique constraint, or `ON CONFLICT`).
- Races: use DB constraints, `SELECT ... FOR UPDATE`, or a `version` column, not in-memory checks.
- Independent I/O runs concurrently with `errgroup.WithContext`, bounded by `g.SetLimit(n)`.
- Every goroutine has an exit path (no leaks). `go test -race` stays clean.
- Retries only for idempotent operations, with backoff + jitter.
- `http.Client` has a `Timeout` and is reused, never `http.DefaultClient` for external calls.
- Graceful shutdown: `signal.NotifyContext` + `srv.Shutdown(ctx)`.

## Security
- Parameterised queries only (sqlc gives this). Never build SQL strings with `fmt.Sprintf`.
- Secrets come from env/secret manager. Never log them or put them in URLs.
- Rate-limit public and auth endpoints.

## Observability
- `log/slog` with structured fields + a request ID. No PII.
- Measure external calls and slow queries.

## Performance rule
Measure first: `go test -bench . -benchmem`, `pprof`, `EXPLAIN ANALYZE`. Change one thing and
re-measure. No guessing, no premature caching. If caching: key, TTL and invalidation are decided up front.

## Debug loop (bugs, panics, wrong results, slowness)
1. **Feedback loop first**: one command that goes **red on this exact bug** - a failing test,
   a `curl` script against the dev server, a replayed payload, or `git bisect run`.
   Fast, deterministic, unattended. Flaky? Loop it (`go test -count=100 -race`) until the rate is high.
   **No red command = no theories.** If you can't build one, say so and ask for logs or access.
2. **Minimise**: cut inputs and steps one at a time until everything left is load-bearing.
3. **Hypothesise**: 3–5 ranked, falsifiable guesses ("if X, then changing Y fixes it").
   Show them to the user before testing.
4. **Instrument**: one variable at a time. Use `dlv` or targeted logs tagged `[DEBUG-xxxx]`.
   Perf bugs: `pprof` + bisect, not logs.
5. **Fix**: the minimal repro becomes a failing test → fix → green → re-run the original loop.
6. **Cleanup**: `grep -r "DEBUG-"` is empty. Root cause goes in the commit message.
Redact secrets as `<REDACTED>` in anything shown.
