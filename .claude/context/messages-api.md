# Messages API (POST /messages, GET /messages/{messageId})

**Status:** done
**Updated:** 2026-10-27
**Branch:** main

## Goal
README Phase 2+3 combined: idempotent message submission + status lookup, backed by Postgres.

## Decisions
- Q1 auth placeholder: `X-User-ID` header stands in for real auth until built later
- Q2 no Kafka yet: insert as `PENDING` and stop; workers/processing arrive in Phase 4
- Q3 duplicate submit: `202` with the original row, not an error
- Q4 validation: userId/messageId/content required non-empty; content capped at 10,000 chars
- Q5 sqlc for DB access
- Q6 golang-migrate for schema migrations
- Mid-build correction: GET lookups scope by (caller's X-User-ID, messageId) together, not
  messageId alone - message_id is only unique per-user (composite key), so a global lookup
  risked returning another user's row nondeterministically. Wrong-owner and truly-missing
  both now return 404 (never 403), to avoid leaking whether an ID exists under another user.

## Approach
sqlc-generated repository (`internal/repository/db`) + a small service layer
(`internal/service`) with idempotent create and owner-scoped lookup, fronted by stdlib
HTTP handlers (`internal/handler`) using a locally-defined interface so handler tests use
a fake instead of the real DB. Service tests run against the real dev Postgres container
(no testcontainers - Docker API access has a permission issue in this shell session, so
tests connect directly via DATABASE_URL to the already-running `deployments-postgres-1`
container instead).

## Plan (tests in build order)
- [x] TestLoad_DatabaseURLFromEnv (config)
- [x] TestCreate_Valid (service)
- [x] TestCreate_Duplicate (service)
- [x] TestGetByID_Owner (service)
- [x] TestGetByID_WrongUser_ReturnsNotFound (service)
- [x] TestGetByID_NotFound (service)
- [x] TestCreate_Valid / InvalidJSON / MissingFields / ContentTooLong (handler)
- [x] TestGetByID_Valid / MissingUserHeader / NotFound (handler)
- [x] End-to-end smoke test via a built binary against the live container

## File map
| File | Job | Lines |
|---|---|---|
| migrations/000001_create_messages.{up,down}.sql | schema + unique/check constraints | ~15 |
| internal/repository/queries/message.sql | sqlc source queries | 8 |
| internal/repository/db/* | sqlc-generated (not hand-written, no line cap) | - |
| internal/service/message_service.go | idempotent create, owner-scoped get | 56 |
| internal/service/errors.go | ErrNotFound | 5 |
| internal/handler/message.go | decode/validate/map, HTTP <-> service | 109 |
| internal/handler/response.go | writeJSON/writeError/writeServiceErr | 31 |
| cmd/api/main.go | wires pgxpool, service, handler, routes | 64 |

## Next step
Manual Postman check (user), then decide: Phase 2 cleanup pass (not much left - the split
is already handler/service/repository), or move to Phase 4 (Kafka) or Phase 6 (rate limiting).

## Open questions / blockers
None.
