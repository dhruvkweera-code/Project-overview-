package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"high-throughput-api/internal/repository/db"
)

type statusCall struct {
	userID, messageID, status string
}

type fakeStatusUpdater struct {
	calls []statusCall
}

func (f *fakeStatusUpdater) UpdateMessageStatus(ctx context.Context, arg db.UpdateMessageStatusParams) (db.Message, error) {
	f.calls = append(f.calls, statusCall{arg.UserID, arg.MessageID, arg.Status})
	return db.Message{UserID: arg.UserID, MessageID: arg.MessageID, Status: arg.Status}, nil
}

func zeroBackoff(attempt int) time.Duration { return 0 }

func newTestProcessor(upd *fakeStatusUpdater, work func(ctx context.Context) error) *Processor {
	return &Processor{q: upd, maxAttempts: 3, backoff: zeroBackoff, work: work}
}

func TestProcess_Success(t *testing.T) {
	upd := &fakeStatusUpdater{}
	p := newTestProcessor(upd, func(ctx context.Context) error { return nil })

	err := p.Process(context.Background(), []byte(`{"userId":"u1","messageId":"m1"}`))
	if err != nil {
		t.Fatalf("Process: %v", err)
	}
	if len(upd.calls) != 2 {
		t.Fatalf("calls = %d, want 2 (PROCESSING, COMPLETED), got %+v", len(upd.calls), upd.calls)
	}
	if upd.calls[0].status != "PROCESSING" || upd.calls[1].status != "COMPLETED" {
		t.Errorf("statuses = %v, want [PROCESSING COMPLETED]", upd.calls)
	}
}

func TestProcess_RetriesThenSucceeds(t *testing.T) {
	attempts := 0
	work := func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("transient failure")
		}
		return nil
	}
	upd := &fakeStatusUpdater{}
	p := newTestProcessor(upd, work)

	if err := p.Process(context.Background(), []byte(`{"userId":"u1","messageId":"m1"}`)); err != nil {
		t.Fatalf("Process: %v", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
	if upd.calls[len(upd.calls)-1].status != "COMPLETED" {
		t.Errorf("final status = %q, want COMPLETED", upd.calls[len(upd.calls)-1].status)
	}
}

func TestProcess_ExhaustsRetries_MarksFailed(t *testing.T) {
	attempts := 0
	work := func(ctx context.Context) error {
		attempts++
		return errors.New("permanent failure")
	}
	upd := &fakeStatusUpdater{}
	p := newTestProcessor(upd, work)

	if err := p.Process(context.Background(), []byte(`{"userId":"u1","messageId":"m1"}`)); err != nil {
		t.Fatalf("Process: %v", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3 (maxAttempts)", attempts)
	}
	if upd.calls[len(upd.calls)-1].status != "FAILED" {
		t.Errorf("final status = %q, want FAILED", upd.calls[len(upd.calls)-1].status)
	}
}

func TestProcess_InvalidEventJSON(t *testing.T) {
	upd := &fakeStatusUpdater{}
	p := newTestProcessor(upd, func(ctx context.Context) error { return nil })

	err := p.Process(context.Background(), []byte(`not json`))
	if err == nil {
		t.Fatal("Process: want error for invalid JSON, got nil")
	}
	if len(upd.calls) != 0 {
		t.Errorf("calls = %d, want 0 (should fail before touching DB)", len(upd.calls))
	}
}
