package redisclient

import (
	"context"
	"fmt"
	"log"

	"config-center/internal/config"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init(cfg *config.Config) error {
	Client = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
	})

	ctx := context.Background()
	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	log.Println("Redis connected successfully")
	return nil
}

func Publish(ctx context.Context, channel string, message string) error {
	return Client.Publish(ctx, channel, message).Err()
}

func Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return Client.Subscribe(ctx, channels...)
}
