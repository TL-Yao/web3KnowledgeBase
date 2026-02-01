package main

import (
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/user/web3-insight/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		},
		asynq.Config{
			Concurrency: cfg.Worker.Concurrency,
			Queues:      cfg.Worker.Queues,
		},
	)

	mux := asynq.NewServeMux()
	// TODO: Register task handlers here
	// mux.HandleFunc("task:process", handleProcessTask)

	log.Printf("Worker starting with concurrency %d", cfg.Worker.Concurrency)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("Worker failed: %v", err)
	}
}
