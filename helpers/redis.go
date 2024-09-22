package helpers

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
)

func initializeRedis(logger *slog.Logger) *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	password := os.Getenv("REDIS_PASSWORD")
	db := os.Getenv("REDIS_DB")
	dbNumber, _ := strconv.Atoi(db)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbNumber,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("Could not connect to Redis", "Error", err)
		return nil
	}

	logger.Info("Connected to Redis!")
	return rdb
}
