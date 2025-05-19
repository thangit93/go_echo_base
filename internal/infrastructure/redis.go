package infrastructure

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/thangit93/echo-base/config"
)

var (
	client *redis.Client
	once   sync.Once
)

// Init initializes the Redis client and tests the connection
func InitRedis() error {
	var err error

	once.Do(func() {
		client = redis.NewClient(&redis.Options{
			Addr:     config.REDIS_ADDR,
			Password: "", // hoặc config.REDIS_PASSWORD nếu có
			DB:       0,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = client.Ping(ctx).Result()
		if err != nil {
			log.Printf("❌ Redis ping failed: %v", err)
			client = nil
		} else {
			log.Println("✅ Connected to Redis!")
		}
	})

	return err
}

// GetClient returns the Redis client instance
func GetClient() *redis.Client {
	if client == nil {
		log.Fatal("Redis client is not initialized. Did you forget to call redisclient.Init()?")
	}
	return client
}
