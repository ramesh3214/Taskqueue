package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ramesh3214/taskflow/dto"
	"github.com/ramesh3214/taskflow/model"
	"github.com/ramesh3214/taskflow/queue"
	"github.com/ramesh3214/taskflow/repository"
)

type Taskworker struct {
	taskRepo    *repository.Taskrepo
	rabbitMQ    *queue.RabbitMQ
	emailWorker *EmailWorker
}

func NewTaskworker(
	taskRepo *repository.Taskrepo,
	rabbitMQ *queue.RabbitMQ,
	emailWorker *EmailWorker,
) *Taskworker {

	return &Taskworker{
		taskRepo:    taskRepo,
		rabbitMQ:    rabbitMQ,
		emailWorker: emailWorker,
	}
}

func (w *Taskworker) Start() error {

	messages, err := w.rabbitMQ.Consume()
	if err != nil {
		return fmt.Errorf(
			"failed to start consumer: %w",
			err,
		)
	}

	log.Println("Task worker is listening for messages...")

	for msg := range messages {

		log.Println(
			"Received message:",
			string(msg.Body),
		)

		// -----------------------------
		// Decode message
		// -----------------------------
		var taskMessage dto.TaskMessage

		err := json.Unmarshal(
			msg.Body,
			&taskMessage,
		)
		if err != nil {
			log.Printf(
				"failed to decode message: %v",
				err,
			)

			// Invalid message ko retry nahi karna.
			// Direct DLQ mein bhejenge.
			err = w.rabbitMQ.Publish(
				"dlq_exchange",
				"task.failed",
				string(msg.Body),
			)
			if err != nil {
				log.Printf(
					"failed to publish message to DLQ: %v",
					err,
				)
				continue
			}

			msg.Ack(false)
			continue
		}

		// -----------------------------
		// Get Task from DB
		// -----------------------------
		task, err := w.taskRepo.GetTaskByID(
			context.Background(),
			taskMessage.TaskID,
		)
		if err != nil {
			log.Printf(
				"failed to get task %d: %v",
				taskMessage.TaskID,
				err,
			)

			continue
		}

		log.Printf(
			"Task fetched: %+v",
			task,
		)

		// -----------------------------
		// Process Task
		// -----------------------------
		err = w.processTask(task)

		if err == nil {

			// -----------------------------
			// SUCCESS
			// -----------------------------
			err = w.taskRepo.UpdateTaskByID(
				context.Background(),
				task.ID,
				"COMPLETED",
			)
			if err != nil {
				log.Printf(
					"failed to mark task %d as COMPLETED: %v",
					task.ID,
					err,
				)
				continue
			}

			log.Printf(
				"Task %d marked COMPLETED",
				task.ID,
			)

			// ACK only after successful processing
			// and successful DB update.
			err = msg.Ack(false)
			if err != nil {
				log.Printf(
					"failed to ACK task %d: %v",
					task.ID,
					err,
				)
				continue
			}

			log.Printf(
				"Task %d ACKed",
				task.ID,
			)

			continue
		}

		// -----------------------------
		// FAILED
		// -----------------------------
		log.Printf(
			"Task %d processing failed: %v",
			task.ID,
			err,
		)

		// Increment retry count
		task, err = w.taskRepo.IncrementRetryCount(
			context.Background(),
			task.ID,
		)
		if err != nil {
			log.Printf(
				"failed to increment retry count: %v",
				err,
			)
			continue
		}

		log.Printf(
			"Task %d retry count: %d",
			task.ID,
			task.RetryCount,
		)

		// -----------------------------
		// Retry
		// -----------------------------
		if task.RetryCount <= 3 {

			err = w.rabbitMQ.Publish(
				"retry_exchange",
				"task.retry",
				string(msg.Body),
			)
			if err != nil {
				log.Printf(
					"failed to publish task %d to retry queue: %v",
					task.ID,
					err,
				)
				continue
			}

			log.Printf(
				"Task %d sent to retry queue",
				task.ID,
			)

			err = msg.Ack(false)
			if err != nil {
				log.Printf(
					"failed to ACK retry message: %v",
					err,
				)
			}

			continue
		}

		// -----------------------------
		// Maximum retries reached
		// -----------------------------

		err = w.taskRepo.UpdateTaskByID(
			context.Background(),
			task.ID,
			"FAILED",
		)
		if err != nil {
			log.Printf(
				"failed to mark task %d as FAILED: %v",
				task.ID,
				err,
			)
			continue
		}

		// Send task to DLQ
		err = w.rabbitMQ.Publish(
			"dlq_exchange",
			"task.failed",
			string(msg.Body),
		)
		if err != nil {
			log.Printf(
				"failed to publish task %d to DLQ: %v",
				task.ID,
				err,
			)
			continue
		}

		log.Printf(
			"Task %d moved to DLQ",
			task.ID,
		)

		// ACK original message after DLQ publish
		err = msg.Ack(false)
		if err != nil {
			log.Printf(
				"failed to ACK DLQ task %d: %v",
				task.ID,
				err,
			)
		}
	}

	return nil
}

func (w *Taskworker) processTask(
	task *model.Task,
) error {

	log.Printf(
		"Processing task %d: %s",
		task.ID,
		task.TaskType,
	)

	switch task.TaskType {

	case "WELCOME_EMAIL":

		log.Printf(
			"Sending welcome email to %s",
			task.Email,
		)

		return w.emailWorker.SendWelcomeEmail(task)

	default:

		return fmt.Errorf(
			"unknown task type: %s",
			task.TaskType,
		)
	}
}
