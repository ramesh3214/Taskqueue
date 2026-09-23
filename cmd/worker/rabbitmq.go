package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/ramesh3214/taskflow/config"
	"github.com/ramesh3214/taskflow/queue"
	"github.com/ramesh3214/taskflow/repository"
	"github.com/ramesh3214/taskflow/service"
	"github.com/ramesh3214/taskflow/worker"
)

func main() {

	// -----------------------------
	// Load Environment Variables
	// -----------------------------
	if err := godotenv.Load(); err != nil {
    log.Println(".env file not found, using environment variables")
}
	// -----------------------------
	// Connect to Database
	// -----------------------------
	db, err := config.DBConnection()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	log.Println("Database connected successfully!")

	// -----------------------------
	// Create Task Repository
	// -----------------------------
	taskRepo := repository.CreateNewTask(db)

	// -----------------------------
	// Connect to RabbitMQ
	// -----------------------------
	rabbitMQ, err := queue.NewRabbitMQ()
	if err != nil {
		log.Fatalf("RabbitMQ connection failed: %v", err)
	}
	defer rabbitMQ.Close()

	log.Println("RabbitMQ connected successfully!")

	// -----------------------------
	// Setup RabbitMQ
	// -----------------------------
	err = rabbitMQ.Setup()
	if err != nil {
		log.Fatalf("RabbitMQ setup failed: %v", err)
	}

	log.Println("RabbitMQ setup successful!")

	// -----------------------------
	// Create Email Service
	// -----------------------------
	emailService := service.NewEmailService()

	// -----------------------------
	// Create Email Worker
	// -----------------------------
	emailWorker := worker.NewEmailWorker(emailService)

	// -----------------------------
	// Create Task Worker
	// -----------------------------
	taskWorker := worker.NewTaskworker(
		taskRepo,
		rabbitMQ,
		emailWorker,
	)

	// -----------------------------
	// Start Worker
	// -----------------------------
	log.Println("Task worker started...")

	// Run worker in separate goroutine
	go func() {
		err := taskWorker.Start()
		if err != nil {
			log.Printf("worker stopped with error: %v", err)
		}
	}()

	// -----------------------------
	// Wait for Shutdown Signal
	// -----------------------------
	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	log.Println("Shutdown signal received...")
	log.Println("Stopping worker...")

	// -----------------------------
	// Graceful Shutdown
	// -----------------------------
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	_ = ctx

	log.Println("Closing RabbitMQ connection...")

	if err := rabbitMQ.Close(); err != nil {
		log.Printf("RabbitMQ close error: %v", err)
	}

	log.Println("Worker shutdown completed!")
}