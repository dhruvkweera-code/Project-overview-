# Load Test Results (single instance, this dev machine)

**Status:** done
**Updated:** 2026-10-27
**Branch:** main
**Machine:** 8 cores, 14GB RAM, shared with VSCode/Docker/this session - not a dedicated server

## Goal
Find the real "safe RPS per instance" number for the capacity formula in README section 8,
instead of guessing, before deciding how many instances Phase 5 needs for 10,000 RPS.

## Tool
`tests/load/main.go` - a small Go program using the `vegeta` library. Generates a unique
`userId`/`messageId` per request (otherwise duplicates would hit the fast idempotent-skip
path instead of a real insert, giving a fake number). Usage:
`loadtest <baseURL> <rps> <durationSeconds>`. Targets `POST /messages` specifically, not
`/health` - the real bottleneck is the DB insert + Kafka publish, not routing.

## Bugs found and fixed via load testing (not guessed, measured)
1. **kafka-go's default `BatchTimeout` is 1 second.** Every synchronous publish was waiting
   up to that long to flush. Fixed: explicit `BatchTimeout: 10ms` in `internal/queue/kafka.go`.
   Effect at 100 RPS: p50 524ms -> 1.5ms, p99 1.0s -> 4.5ms.
2. **pgxpool's default pool size is `max(4, NumCPU)`** - 8 connections on this machine.
   Way too small for thousands of concurrent requests; connections became the queueing
   bottleneck. Fixed: explicit, configurable pool size (`DB_MAX_CONNS` env var, default 50)
   via `pgxpool.ParseConfig` + `MaxConns` in `cmd/api/main.go`. Postgres itself allows up to
   100 connections by default, confirmed via `SHOW max_connections`, so 50 leaves headroom.
   Effect at 5000 RPS: p50 4.3s -> 11ms, achieved rate 3728/s -> full 5000/s.

## Results (after both fixes, single API instance)
| Target RPS | Achieved | p50 | p95 | p99 | Max | Success |
|---|---|---|---|---|---|---|
| 100 | 100/s | 1.5ms | 2.1ms | 3.9ms | 14ms | 100% |
| 1,000 | 1000/s | 0.9ms | 1.7ms | 7.5ms | 69ms | 100% |
| 3,000 | 3000/s | 1.0ms | 4.5ms | 22ms | 139ms | 100% |
| 5,000 | 5000/s | 11ms | 415ms | 496ms | 632ms | 100% (tail latency rising) |
| 10,000 | only 4675/s | 2.0s | 4.5s | 4.8s | 5.3s | 100% (but broken - severe queueing) |

## Conclusion
**Safe RPS per instance on this machine: ~3,000.** 5,000 still technically "succeeds" (no
HTTP errors) but tail latency (p95/p99 in the hundreds of ms) signals it's past the comfortable
ceiling. 10,000 is a clear breakdown - achieved rate falls well short of target with
multi-second latency.

Using README section 8's formula for the 10,000 RPS target:
```
Required instances = 10,000 / 3,000 = 4 instances (round up), + spare capacity for failures/spikes
```
This number is specific to this dev machine's hardware and default Postgres/Kafka config
(unoptimized, single-node, shared with other processes). A production deployment on real
servers, with tuned Postgres and a multi-broker Kafka cluster, would very likely sustain a
much higher per-instance number - this result tells us the *methodology* works and gives a
conservative planning baseline, not a hard production ceiling.

## Follow-up: tried raising the pool further - made things worse, not better
Raised `DB_MAX_CONNS` to 150 and Postgres's own `max_connections` to 300 (via
`docker-compose.yml` command override) to see if 10,000 RPS was a tuning problem or a
hardware ceiling. Result: **worse**, not better - 5,000 RPS latency went from p50=11ms to
p50=1.7s, and CPU monitoring (`mpstat`) during the 10,000 RPS run showed ~99% CPU
utilization with rising `%sys` (context-switch/kernel overhead), confirming a genuine
8-core CPU ceiling, not a connection-starvation problem. More concurrent connections than
the CPU can actually run in parallel just adds scheduling overhead. Reverted `DB_MAX_CONNS`
default back to 50 (the measured sweet spot); left Postgres's `max_connections=300` in
`docker-compose.yml` since it's harmless headroom, just unused by the app itself.

**Conclusion stands: ~3,000 RPS/instance is the real ceiling on this hardware**, and it is
a CPU limit, not a configuration mistake - confirms Phase 5 (more instances, i.e. more
CPU cores total) is the correct next step, not further single-instance tuning.

## What wasn't measured yet (scope for later)
- CPU/memory saturation point specifically (didn't profile with pprof/top during the run)
- Worker/Kafka consumer throughput and lag under sustained load (worker was not running
  during these tests - deliberately isolated to measure API accept-path capacity only,
  per README's own distinction between scaling the API vs. scaling workers)
- Behavior under sustained load over minutes (each stage was only 10 seconds)
- Failure-injection scenarios (DB down, Kafka down, instance crash) - README section 9

## Next step
Phase 5 (horizontal scaling + load balancer), using 3,000 RPS/instance as the planning
number, or deepen the load test (longer duration, worker running, failure injection) first.
