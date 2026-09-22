package worker

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ramesh3214/taskflow/repository"
	"github.com/redis/go-redis/v9"
)

type TaskWorker struct {
	redis    *redis.Client
	taskRepo *repository.Taskrepo
	workerID string
}

func NewTaskWorker(
	redisClient *redis.Client,
	taskRepo *repository.Taskrepo,
	workerID string,
) *TaskWorker {
	return &TaskWorker{
		redis:    redisClient,
		taskRepo: taskRepo,
		workerID: workerID,
	}
}

func (w *TaskWorker) Start(ctx context.Context) {

	fmt.Printf("%s started...\n", w.workerID)

	// Channel used to send task IDs
	// from Start() to worker goroutines.
	jobs := make(chan uint)

	// Number of goroutines inside one worker process.
	workerCount := 3

	// Start worker goroutines.
	for i := 0; i < workerCount; i++ {
		go w.processJobs(ctx, jobs)
	}

	// Continuously receive tasks from Redis.
	for {

		result, err := w.redis.BLPop(
			ctx,
			0,
			"task_queue",
		).Result()

		if err != nil {

			// Context cancelled means worker is shutting down.
			if ctx.Err() != nil {
				fmt.Printf(
					"%s stopping...\n",
					w.workerID,
				)
				return
			}

			fmt.Println("worker error:", err)
			continue
		}

		// Redis BLPOP result:
		// result[0] = queue name
		// result[1] = task ID
		taskID, err := strconv.ParseUint(
			result[1],
			10,
			64,
		)

		if err != nil {
			fmt.Println(
				"invalid task ID:",
				err,
			)
			continue
		}

		fmt.Printf(
			"%s received task: %d\n",
			w.workerID,
			taskID,
		)

		// Send task ID to one of the
		// available goroutines.
		select {

		case jobs <- uint(taskID):

		case <-ctx.Done():
			return
		}
	}
}

func (w *TaskWorker) processJobs(
	ctx context.Context,
	jobs <-chan uint,
) {

	for {

		select {

		// Receive task ID from jobs channel.
		case taskID := <-jobs:

			fmt.Printf(
				"%s processing task: %d\n",
				w.workerID,
				taskID,
			)

			// Get task from PostgreSQL.
			task, err := w.taskRepo.GetTaskByID(
				ctx,
				taskID,
			)

			if err != nil {
				fmt.Println(
					"failed to get task:",
					err,
				)
				continue
			}

			// ==================================================
			// TEMPORARY FAILURE SIMULATION
			// ==================================================
			//
			// Tasks divisible by 3 will fail.
			//
			// Example:
			// 3  -> fail
			// 6  -> fail
			// 9  -> fail
			//
			// Other tasks will succeed.
			//
			if task.ID%3 == 0 {

				fmt.Printf(
					"%s task %d failed\n",
					w.workerID,
					task.ID,
				)

				// Increase retry count.
				task, err = w.taskRepo.IncrementRetryCount(
					ctx,
					task.ID,
				)

				if err != nil {
					fmt.Println(
						"failed to increment retry count:",
						err,
					)
					continue
				}

				fmt.Printf(
					"%s task %d retry count: %d\n",
					w.workerID,
					task.ID,
					task.RetryCount,
				)

				// Maximum 3 retries.
				if task.RetryCount <= 3 {

					// Put task back into Redis queue.
					err = w.redis.RPush(
						ctx,
						"task_queue",
						task.ID,
					).Err()

					if err != nil {
						fmt.Println(
							"failed to push task back to queue:",
							err,
						)
						continue
					}

					fmt.Printf(
						"%s task %d added back to queue\n",
						w.workerID,
						task.ID,
					)

				} else {

					// Maximum retries exceeded.
					err = w.taskRepo.UpdateTaskByID(
						ctx,
						task.ID,
						"FAILED",
					)

					if err != nil {
						fmt.Println(
							"failed to mark task as failed:",
							err,
						)
						continue
					}

					err = w.redis.RPush(ctx, "failed_task_queue", task.ID).Err()

					if err != nil {
						fmt.Println("failed to push task to failed queue:", err)
						continue
					}

					fmt.Printf(
						"%s task %d permanently failed\n",
						w.workerID,
						task.ID,
					)
				}
				
				continue
			}

			
			err = w.taskRepo.UpdateTaskByID(
				ctx,
				task.ID,
				"COMPLETED",
			)

			if err != nil {
				fmt.Println(
					"failed to update task:",
					err,
				)
				continue
			}

			fmt.Printf(
				"%s completed task: %d\n",
				w.workerID,
				task.ID,
			)

		// Worker shutdown.
		case <-ctx.Done():

			fmt.Printf(
				"%s goroutine stopped\n",
				w.workerID,
			)

			return
		}
	}
}
