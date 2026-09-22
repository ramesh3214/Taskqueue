package queue

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type TaskQueue struct {
	redis *redis.Client
}

func NewTaskQueue(redisClient *redis.Client) *TaskQueue {
	return &TaskQueue{
		redis: redisClient,
	}
}

func (q *TaskQueue) Push(
	ctx context.Context,
	taskID uint,
) error {
	err := q.redis.RPush(
		ctx,
		"task_queue",
		taskID,
	).Err()

	if err != nil {
		return fmt.Errorf("failed to push task to queue: %w", err)
	}

	return nil
}
