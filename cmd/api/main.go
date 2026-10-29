package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"
	"time"

	"high-throughput-api/internal/config"
	"high-throughput-api/internal/handler"
	"high-throughput-api/internal/queue"
	"high-throughput-api/internal/repository/db"
	"high-throughput-api/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	go func() {
		log.Println("pprof debug server on localhost:6060")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("parse database config: %v", err)
	}
	poolConfig.MaxConns = cfg.DBMaxConns

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	publisher := queue.NewKafkaPublisher(cfg.KafkaBrokers, queue.TopicMessages)
	defer publisher.Close()

	messageSvc := service.NewMessageService(db.New(pool), publisher)
	messageHandler := handler.NewMessageHandler(messageSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("POST /messages", messageHandler.Create)
	mux.HandleFunc("GET /messages/{messageId}", messageHandler.GetByID)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}
