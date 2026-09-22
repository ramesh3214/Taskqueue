package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/ramesh3214/taskflow/config"
	"github.com/ramesh3214/taskflow/repository"
	"github.com/ramesh3214/taskflow/worker"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatalf(
			"error loading .env file: %v",
			err,
		)
	}

	db, err := config.DBConnection()
	if err != nil {
		log.Fatalf(
			"database connection failed: %v",
			err,
		)
	}

	redisClient, err := config.RedisConnect()
	if err != nil {
		log.Fatalf(
			"redis connection failed: %v",
			err,
		)
	}
	defer redisClient.Close()

	workerID := os.Getenv("WORKER_ID")

	if workerID == "" {
		workerID = "worker-default"
	}

	taskRepo := repository.CreateNewTask(db)

	taskWorker := worker.NewTaskWorker(
		redisClient,
		taskRepo,
		workerID,
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	go taskWorker.Start(ctx)

	log.Println("TaskFlow Worker started")

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	log.Println("Worker shutdown signal received...")

	cancel()

	log.Println("Worker stopped")
}
