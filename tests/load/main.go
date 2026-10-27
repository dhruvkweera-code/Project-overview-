package main

import (
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func messageTargeter(baseURL string) vegeta.Targeter {
	var counter int64
	return func(tgt *vegeta.Target) error {
		if tgt == nil {
			return vegeta.ErrNilTarget
		}
		n := atomic.AddInt64(&counter, 1)
		body := fmt.Sprintf(`{"userId":"load-user-%d","messageId":"load-msg-%d","content":"load test payload"}`, n, n)

		tgt.Method = http.MethodPost
		tgt.URL = baseURL + "/messages"
		tgt.Body = []byte(body)
		tgt.Header = http.Header{"Content-Type": []string{"application/json"}}
		return nil
	}
}

func runStage(baseURL string, rps int, duration time.Duration) {
	targeter := messageTargeter(baseURL)
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	rate := vegeta.Rate{Freq: rps, Per: time.Second}
	for res := range attacker.Attack(targeter, rate, duration, fmt.Sprintf("%d rps", rps)) {
		metrics.Add(res)
	}
	metrics.Close()

	fmt.Printf("\n=== target %d RPS for %s ===\n", rps, duration)
	fmt.Printf("requests: %d, achieved rate: %.1f/s, success: %.1f%%\n",
		metrics.Requests, metrics.Rate, metrics.Success*100)
	fmt.Printf("latency  p50=%s p95=%s p99=%s max=%s\n",
		metrics.Latencies.P50, metrics.Latencies.P95, metrics.Latencies.P99, metrics.Latencies.Max)
	fmt.Printf("status codes: %v\n", metrics.StatusCodes)
	if len(metrics.Errors) > 0 {
		fmt.Printf("errors: %v\n", metrics.Errors)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: loadtest <baseURL> <rps> [durationSeconds]")
		os.Exit(1)
	}
	baseURL := os.Args[1]

	var rps int
	fmt.Sscanf(os.Args[2], "%d", &rps)

	duration := 10 * time.Second
	if len(os.Args) > 3 {
		var secs int
		fmt.Sscanf(os.Args[3], "%d", &secs)
		duration = time.Duration(secs) * time.Second
	}

	runStage(baseURL, rps, duration)
}
