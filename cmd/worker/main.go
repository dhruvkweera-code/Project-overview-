package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"high-throughput-api/internal/config"
	"high-throughput-api/internal/queue"
	"high-throughput-api/internal/repository/db"
	"high-throughput-api/internal/worker"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	consumer := queue.NewKafkaConsumer(cfg.KafkaBrokers, queue.TopicMessages, "message-workers")
	defer consumer.Close()

	processor := worker.NewProcessor(db.New(pool))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Println("worker started, consuming topic", queue.TopicMessages)
	for {
		msg, err := consumer.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			log.Printf("fetch message: %v", err)
			continue
		}

		log.Printf("received message: partition=%d offset=%d value=%s", msg.Partition, msg.Offset, msg.Value)

		if err := processor.Process(ctx, msg.Value); err != nil {
			log.Printf("process message: %v (not committing, will retry on next poll)", err)
			continue
		}

		log.Printf("processed message: offset=%d", msg.Offset)
		if err := consumer.CommitMessages(ctx, msg); err != nil {
			log.Printf("commit message: %v", err)
		}
	}

	log.Println("worker shutting down")
}
