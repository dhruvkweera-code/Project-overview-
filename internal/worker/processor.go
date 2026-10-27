package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"high-throughput-api/internal/repository/db"

	"github.com/jackc/pgx/v5/pgtype"
)

type statusUpdater interface {
	UpdateMessageStatus(ctx context.Context, arg db.UpdateMessageStatusParams) (db.Message, error)
}

type Processor struct {
	q           statusUpdater
	maxAttempts int
	backoff     func(attempt int) time.Duration
	work        func(ctx context.Context) error
}

func NewProcessor(q statusUpdater) *Processor {
	return &Processor{
		q:           q,
		maxAttempts: 3,
		backoff:     defaultBackoff,
		work:        placeholderWork,
	}
}

func defaultBackoff(attempt int) time.Duration {
	base := time.Duration(1<<attempt) * 100 * time.Millisecond
	return base + time.Duration(rand.Int63n(int64(base)))
}

func placeholderWork(ctx context.Context) error {
	time.Sleep(50 * time.Millisecond)
	return nil
}

type messageEvent struct {
	UserID    string `json:"userId"`
	MessageID string `json:"messageId"`
}

func (p *Processor) Process(ctx context.Context, value []byte) error {
	var evt messageEvent
	if err := json.Unmarshal(value, &evt); err != nil {
		return fmt.Errorf("decode event: %w", err)
	}

	if err := p.setStatus(ctx, evt, "PROCESSING", false); err != nil {
		return fmt.Errorf("mark processing: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= p.maxAttempts; attempt++ {
		lastErr = p.work(ctx)
		if lastErr == nil {
			break
		}
		if attempt < p.maxAttempts {
			time.Sleep(p.backoff(attempt))
		}
	}

	finalStatus := "COMPLETED"
	if lastErr != nil {
		finalStatus = "FAILED"
	}
	if err := p.setStatus(ctx, evt, finalStatus, true); err != nil {
		return fmt.Errorf("mark %s: %w", finalStatus, err)
	}
	return nil
}

func (p *Processor) setStatus(ctx context.Context, evt messageEvent, status string, setProcessedAt bool) error {
	var processedAt pgtype.Timestamptz
	if setProcessedAt {
		processedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	}
	_, err := p.q.UpdateMessageStatus(ctx, db.UpdateMessageStatusParams{
		UserID:      evt.UserID,
		MessageID:   evt.MessageID,
		Status:      status,
		ProcessedAt: processedAt,
	})
	return err
}
