package service

import (
	"context"
	"errors"
	"fmt"

	"high-throughput-api/internal/repository/db"

	"github.com/jackc/pgx/v5"
)

// Update replaces the content of one message and sends it to Kafka to be processed again.
// The message is found by its unique key (userID, messageID), so users can only change their own.
func (s *MessageService) Update(ctx context.Context, userID, messageID, newContent string) (db.Message, error) {
	updatedMessage, err := s.q.UpdateMessageContent(ctx, db.UpdateMessageContentParams{
		UserID:    userID,
		MessageID: messageID,
		Content:   newContent,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.Message{}, s.whyNotUpdated(ctx, userID, messageID)
	}
	if err != nil {
		return db.Message{}, fmt.Errorf("update message: %w", err)
	}

	if err := s.publishMessageEvent(ctx, updatedMessage); err != nil {
		return db.Message{}, fmt.Errorf("update message: %w: %v", ErrQueueUnavailable, err)
	}
	return updatedMessage, nil
}

// whyNotUpdated runs only when the UPDATE changed no row: the message is missing or being processed.
func (s *MessageService) whyNotUpdated(ctx context.Context, userID, messageID string) error {
	_, err := s.q.GetMessageByUserAndMessageID(ctx, db.GetMessageByUserAndMessageIDParams{
		UserID:    userID,
		MessageID: messageID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update message: check existing: %w", err)
	}
	return ErrConflict
}
