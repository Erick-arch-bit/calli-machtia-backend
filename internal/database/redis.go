package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis(redisURL string) *redis.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("WARNING: Unable to parse Redis URL: %v", err)
		return nil
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("WARNING: Unable to ping Redis: %v", err)
		return nil
	}

	fmt.Println("Redis connected successfully")
	return client
}
