package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"high-throughput-api/internal/repository/db"

	"github.com/jackc/pgx/v5"
)

type Publisher interface {
	Publish(ctx context.Context, key, value []byte) error
}

type MessageService struct {
	q   db.Querier
	pub Publisher
}

func NewMessageService(q db.Querier, pub Publisher) *MessageService {
	return &MessageService{q: q, pub: pub}
}

type messageEvent struct {
	UserID    string `json:"userId"`
	MessageID string `json:"messageId"`
}

func (s *MessageService) Create(ctx context.Context, userID, messageID, content string) (db.Message, error) {
	msg, err := s.q.CreateMessage(ctx, db.CreateMessageParams{
		UserID:    userID,
		MessageID: messageID,
		Content:   content,
	})
	if err == nil {
		if pubErr := s.publishCreated(ctx, msg); pubErr != nil {
			return db.Message{}, fmt.Errorf("create message: publish: %w", pubErr)
		}
		return msg, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.Message{}, fmt.Errorf("create message: %w", err)
	}

	existing, err := s.q.GetMessageByUserAndMessageID(ctx, db.GetMessageByUserAndMessageIDParams{
		UserID:    userID,
		MessageID: messageID,
	})
	if err != nil {
		return db.Message{}, fmt.Errorf("create message: fetch existing after conflict: %w", err)
	}
	return existing, nil
}

func (s *MessageService) publishCreated(ctx context.Context, msg db.Message) error {
	value, err := json.Marshal(messageEvent{UserID: msg.UserID, MessageID: msg.MessageID})
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return s.pub.Publish(ctx, []byte(msg.UserID), value)
}

func (s *MessageService) GetByID(ctx context.Context, callerUserID, messageID string) (db.Message, error) {
	msg, err := s.q.GetMessageByUserAndMessageID(ctx, db.GetMessageByUserAndMessageIDParams{
		UserID:    callerUserID,
		MessageID: messageID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.Message{}, ErrNotFound
	}
	if err != nil {
		return db.Message{}, fmt.Errorf("get message: %w", err)
	}
	return msg, nil
}
