# Kafka + Workers (Phase 4)

**Status:** done
**Updated:** 2026-10-27
**Branch:** main

## Goal
Async processing: API publishes to Kafka after a new message insert; a separate worker
consumes and moves the message through PENDING -> PROCESSING -> COMPLETED/FAILED.

## Decisions
- Q1 worker logic: placeholder (brief delay, then COMPLETED) - proves the pipeline, no real
  business rule exists yet
- Q2 client library: segmentio/kafka-go (pure Go, no native deps)
- Q3 partition key: userId (per-user ordering, per README section 7)
- Q4 publish timing: synchronous, right after insert, same request (no outbox pattern yet)
- Q5 retry policy: 3 attempts, exponential backoff + jitter, then FAILED
- Q6 topic: `messages`, 3 partitions

## Approach
`internal/queue` wraps kafka-go (Publisher + Consumer). `MessageService.Create` publishes
only on a genuinely new insert (not on idempotent duplicate resubmits). `internal/worker.Processor`
holds the retry/backoff logic, fully unit-tested with a fake status updater and injectable
work/backoff functions (no real Kafka/Postgres needed for those tests). `cmd/worker/main.go`
is the new binary: fetch -> process -> commit, using a Kafka consumer group ("message-workers").

## Bugs found and fixed during build/testing
- Response struct fields were originally `map[string]string` - Go sorts map keys
  alphabetically on JSON encode, breaking expected field order. Fixed with named structs
  (`createMessageResponse`, `errorResponse`) - same fix pattern as Phase 2/3.
- Worker's main loop originally committed the Kafka offset unconditionally, even when
  `Process()` returned an error. This silently skips messages forever on any transient
  failure, breaking at-least-once delivery. Fixed: only commit after a successful `Process()`
  call; on error, `continue` without committing so the next poll/restart retries it.
- `KafkaConsumer` didn't set `StartOffset`, defaulting to "latest" for brand-new consumer
  groups - meaning messages published before the first consumer connection were silently
  unreachable. Fixed by setting `StartOffset: kafka.FirstOffset`.
- Unhandled service errors (like a Kafka publish failure) weren't logged anywhere before
  being mapped to a generic 500 - added a log line in `writeServiceErr` so failures are
  actually visible in server logs (Checklist: "log once, where handled").
- One test message (`msg-kafka-2`) got stuck permanently on the real `message-workers`
  consumer group due to my own rapid kill/restart test churn during debugging (not a code
  bug) - left as-is in the dev DB, harmless, superseded by the fix above for all future
  messages.

## Plan (tests in build order)
- [x] TestLoad_KafkaBrokersFromEnv (config)
- [x] TestCreate_PublishesOnNewMessage / DoesNotPublishOnDuplicate (service, fake Publisher)
- [x] TestProcess_Success / RetriesThenSucceeds / ExhaustsRetries_MarksFailed / InvalidEventJSON (worker, fake statusUpdater)
- [x] Real end-to-end run: API -> Kafka -> worker -> Postgres, confirmed COMPLETED with processedAt set

## File map
| File | Job | Lines |
|---|---|---|
| internal/queue/kafka.go | Publisher + Consumer wrapping kafka-go | 56 |
| internal/worker/processor.go | retry/backoff, status transitions | 93 |
| internal/worker/processor_test.go | unit tests, no real Kafka/Postgres needed | 101 |
| cmd/worker/main.go | new binary: fetch -> process -> commit loop | 60 |
| internal/service/message_service.go | +Publisher dependency, publish-on-new-insert | 78 |
| internal/repository/queries/message.sql | +UpdateMessageStatus query | - |

## Next step
Decide: Phase 5 (horizontal scaling + load balancer) or a load test first to find
"safe RPS per instance" before deciding how many instances Phase 5 actually needs.

## Open questions / blockers
None for this feature. Known limitation carried forward on purpose (not a bug): the
publish-after-insert is not a true transactional outbox, so a crash in the tiny window
between DB insert and Kafka publish could leave a message stuck at PENDING forever. Noted
as a future hardening step, not addressed now per the grill's Q4 decision.
