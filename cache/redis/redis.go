package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	Client *redis.Client
}

func New(addr string, password string, db int) *Cache {
	return &Cache{
		Client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password, // no password set
			DB:       db,       // use default DB
		}),
	}
}

func (c *Cache) Get(key string) ([]byte, error) {
	val, err := c.Client.Get(context.TODO(), key).Result()
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

func (c *Cache) Set(key string, value []byte, ttl int) error {
	err := c.Client.Set(context.TODO(), key, value, time.Second*time.Duration(ttl)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) Delete(key string) error {
	// TODO: implement delete if needed
	return nil
}
