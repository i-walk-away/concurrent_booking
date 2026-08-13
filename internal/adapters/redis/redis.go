package redis

import (
	"context"
	"fmt"
	"log"

	goredis "github.com/redis/go-redis/v9"
)

// NewClient creates a Redis client connected to the given address.
//
// The returned client is verified with a PING before being returned.
func NewClient(addr string) (*goredis.Client, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr: addr,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		_ = rdb.Close()

		return nil, fmt.Errorf("redis ping: %w", err)
	}

	log.Printf("connected to redis at %s", addr)

	return rdb, nil
}
