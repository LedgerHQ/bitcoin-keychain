package keystore

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
)

func set(c *redis.Client, key string, value interface{}) error {
	p, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.Set(context.Background(), key, string(p), 0).Err()
}

func get(c *redis.Client, key string, dest interface{}) error {
	p, err := c.Get(context.Background(), key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(p), dest)
}
