# Phase 1 - Basic Go API

**Status:** done
**Updated:** 2026-10-27
**Branch:** main

## Goal
Scaffold the Go API with a working, hardened HTTP server and a `/health` check, per README Phase 1.

## Decisions
- Q1 health shape: single `GET /health` -> `{"status":"ok"}`, no separate liveness/readiness yet
- Q2 config: `PORT` env var, default `8080`
- Q3 routing: stdlib `net/http` (Go 1.22+ method patterns), no router library
- Q4 hardening: server timeouts + graceful shutdown added now
- Q5 module path: `high-throughput-api` (local placeholder, rename later with `go mod edit -module`)
- Q6 validation: deferred - no real request body exists yet

## Approach
Minimal stdlib HTTP server (`cmd/api/main.go`) wired to `internal/config` (env-based port) and
`internal/handler` (health check). No framework, no DB/Kafka yet - those arrive in later phases.

## Plan (tests in build order)
- [x] TestLoad_DefaultPort
- [x] TestLoad_PortFromEnv
- [x] TestHealth_ReturnsOK

## File map
| File | Job | Lines |
|---|---|---|
| internal/config/config.go | read PORT env var | 15 |
| internal/handler/health.go | GET /health -> JSON ok | 9 |
| cmd/api/main.go | wire server, timeouts, graceful shutdown | 49 |

## Next step
Manual check: run `go run ./cmd/api` and hit `GET /health` from Postman/curl. Then decide
next endpoint (README Phase 2: clean architecture split, or start on `/messages`).

## Open questions / blockers
None.
