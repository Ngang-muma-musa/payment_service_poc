package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"paymentservice/internal/infrastructure/beanstalk"
	"paymentservice/internal/infrastructure/orm"
	"paymentservice/internal/infrastructure/redis"
	"paymentservice/internal/infrastructure/worker"
	"strconv"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()
	beanstalkAddr := os.Getenv("BEANSTALK_ADR")
	redisDsn := os.Getenv("REDIS_DSN")
	tubeName := os.Getenv("BEANSTALK_TUBE_NAME")
	maxWorkersStr := os.Getenv("MAX_WORKERS")
	if maxWorkersStr == "" {
		maxWorkersStr = "1"
	}
	maxWorkers, err := strconv.Atoi(maxWorkersStr)
	if err != nil {
		log.Fatalf("invalid MAX_WORKERS: %v", err)
	}

	conn, err := beanstalk.BuildBeanstalkQueue(beanstalkAddr)
	if err != nil {
		log.Fatalf("failed to connect to beanstalk: %v", err)
	}

	cache, err := redis.BuildRedisCache(redisDsn)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	repo := orm.NewRedisPaymentRepository(cache, 0)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 1; i <= maxWorkers; i++ {
		w := worker.NewWorker(
			ctx,
			conn,
			repo,
			tubeName,
		)
		go func(id int) {
			log.Printf("Worker #%d started", id)
			w.Start(ctx)
			log.Printf("Worker #%d stopped", id)
		}(i)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()

	time.Sleep(2 * time.Second)
	log.Println("Worker stopped.")
}
