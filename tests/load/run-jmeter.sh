#!/bin/bash
# Usage: ./run-jmeter.sh <host> <port> <rps> <durationSeconds>
set -euo pipefail

HOST="${1:-localhost}"
PORT="${2:-8080}"
RPS="${3:-1000}"
DURATION="${4:-10}"
# Threads needed ~= RPS x expected_latency_seconds (Little's Law). We assume up to
# 1 full second of latency under load so we don't under-provision threads ourselves.
THROUGHPUT=$((RPS * 60))
THREADS=$RPS
if [ "$THREADS" -lt 50 ]; then THREADS=50; fi
if [ "$THREADS" -gt 3000 ]; then THREADS=3000; fi

JMETER_BIN="$HOME/tools/apache-jmeter-5.6.3/bin/jmeter"
RESULTS_FILE="/tmp/jmeter-results-$(date +%s).jtl"

echo "Running JMeter: ${RPS} RPS target against ${HOST}:${PORT} for ${DURATION}s (threads=${THREADS})"

"$JMETER_BIN" -n -t "$(dirname "$0")/jmeter-test-plan.jmx" \
  -Jhost="$HOST" -Jport="$PORT" -Jthroughput="$THROUGHPUT" -Jduration="$DURATION" -Jthreads="$THREADS" \
  -l "$RESULTS_FILE"

echo ""
echo "=== Summary ==="
echo "Results saved to: $RESULTS_FILE"
