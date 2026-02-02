package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/database"
	"github.com/user/web3-insight/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database for worker dependencies
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connected for worker")

	// Initialize worker dependencies (RSS collector, web crawler, embedding service, etc.)
	worker.InitWorkerDependencies(db, &cfg.LLM)
	log.Println("Worker dependencies initialized")

	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: cfg.Worker.Concurrency,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			log.Printf("Task %s failed: %v", task.Type(), err)
		}),
	})

	mux := worker.NewTaskMux()

	log.Printf("Worker starting with concurrency %d", cfg.Worker.Concurrency)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("Worker failed: %v", err)
	}
}
