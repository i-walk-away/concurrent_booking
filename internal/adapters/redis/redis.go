package redis

import (
	"context"
	"log"

	goredis "github.com/redis/go-redis/v9"
)

// NewClient creates a Redis client connected to the given address.
//
// It terminates the process if the Redis server is unreachable.
func NewClient(addr string) *goredis.Client {
	rdb := goredis.NewClient(&goredis.Options{
		Addr: addr,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis ping: %v", err)
	}

	log.Printf("connected to redis at %s", addr)

	return rdb
}
