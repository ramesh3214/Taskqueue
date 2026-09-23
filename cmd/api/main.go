package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/ramesh3214/taskflow/config"
	"github.com/ramesh3214/taskflow/controller"
	"github.com/ramesh3214/taskflow/model"
	"github.com/ramesh3214/taskflow/queue"
	"github.com/ramesh3214/taskflow/repository"
	"github.com/ramesh3214/taskflow/route"
	"github.com/ramesh3214/taskflow/service"
)

func main() {

	// Load environment variables
	if err := godotenv.Load(); err != nil {
    log.Println(".env file not found, using environment variables")
}

	// -------------------------
	// PostgreSQL
	// -------------------------

	db, err := config.DBConnection()
	if err != nil {
		log.Fatalf(
			"database connection failed: %v",
			err,
		)
	}

	// Database migration
	if err := db.AutoMigrate(
		&model.Auth{},
		&model.Task{},
	); err != nil {
		log.Fatalf(
			"database migration failed: %v",
			err,
		)
	}

	// -------------------------
	// Redis
	// -------------------------
	// Redis is still available
	// for caching.
	// It is NOT being used as a queue.

	redisClient, err := config.RedisConnect()
	if err != nil {
		log.Fatalf(
			"redis connection failed: %v",
			err,
		)
	}

	defer redisClient.Close()

	// -------------------------
	// RabbitMQ
	// -------------------------

	rabbitMQ, err := queue.NewRabbitMQ()
	if err != nil {
		log.Fatalf(
			"RabbitMQ connection failed: %v",
			err,
		)
	}

	log.Println(
		"RabbitMQ connected successfully!",
	)

	// Create exchange + queue + binding
	if err := rabbitMQ.Setup(); err != nil {
		log.Fatalf(
			"RabbitMQ setup failed: %v",
			err,
		)
	}

	log.Println(
		"RabbitMQ setup successful!",
	)

	defer rabbitMQ.Close()

	// -------------------------
	// Repositories
	// -------------------------

	authRepo := repository.NewAuthRepo(db)

	taskRepo := repository.CreateNewTask(db)

	// -------------------------
	// Services
	// -------------------------

	/*
		IMPORTANT:

		Old:

		taskQueue := queue.NewTaskQueue(redisClient)

		New:

		RabbitMQ is the task queue.
	*/

	authService := service.NewAuthService(
		redisClient,
		authRepo,
		taskRepo,
		rabbitMQ,
	)

	// -------------------------
	// Controllers
	// -------------------------

	authController := controller.Newcontroller(
		authService,
	)

	// -------------------------
	// Router
	// -------------------------

	router := mux.NewRouter()

	router.HandleFunc(
		"/health",
		health,
	).Methods(http.MethodGet)

	route.AuthRoute(
		router,
		authController,
	)

	// -------------------------
	// HTTP Server
	// -------------------------

	server := &http.Server{
		Addr:    ":5000",
		Handler: router,
	}

	go func() {

		log.Println(
			"TaskFlow API running on :5000",
		)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			log.Fatalf(
				"server error: %v",
				err,
			)
		}
	}()

	// -------------------------
	// Graceful Shutdown
	// -------------------------

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	log.Println(
		"Shutdown signal received...",
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {

		log.Printf(
			"graceful shutdown failed: %v",
			err,
		)

	} else {

		log.Println(
			"server shutdown completed",
		)
	}
}

func health(
	w http.ResponseWriter,
	r *http.Request,
) {

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	message := map[string]string{
		"status":  "ok",
		"message": "server is working",
	}

	w.WriteHeader(
		http.StatusOK,
	)

	if err := json.NewEncoder(w).Encode(message); err != nil {

		log.Printf(
			"failed to encode health response: %v",
			err,
		)
	}
}