package queue

import (
	"context"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

const TopicMessages = "messages"

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(brokers, topic string) *KafkaPublisher {
	return &KafkaPublisher{writer: &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(brokers, ",")...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
		BatchTimeout: 10 * time.Millisecond,
	}}
}

func (p *KafkaPublisher) Publish(ctx context.Context, key, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{Key: key, Value: value})
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}

type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaConsumer(brokers, topic, groupID string) *KafkaConsumer {
	return &KafkaConsumer{reader: kafka.NewReader(kafka.ReaderConfig{
		Brokers:     strings.Split(brokers, ","),
		Topic:       topic,
		GroupID:     groupID,
		StartOffset: kafka.FirstOffset,
	})}
}

func (c *KafkaConsumer) FetchMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.FetchMessage(ctx)
}

func (c *KafkaConsumer) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return c.reader.CommitMessages(ctx, msgs...)
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
