package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func RedisConnect() (*redis.Client, error) {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,

		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// Check Redis connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()

		return nil, fmt.Errorf(
			"redis connection failed: %w",
			err,
		)
	}

	fmt.Println("Redis connected successfully")

	return rdb, nil
}