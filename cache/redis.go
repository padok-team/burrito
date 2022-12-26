package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	Client *redis.Client
}

func NewRedisCache(addr string, password string, db int) *RedisCache {
	return &RedisCache{
		Client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password, // no password set
			DB:       db,       // use default DB
		}),
	}
}

func (r *RedisCache) Get(key string) ([]byte, error) {
	val, err := r.Client.Get(context.TODO(), key).Result()
	if err == redis.Nil {
		return nil, &CacheError{
			Err: err,
			Nil: true,
		}
	}
	if err != nil {
		return nil, &CacheError{
			Err: err,
			Nil: false,
		}
	}
	return []byte(val), nil
}

func (r *RedisCache) Set(key string, value []byte, ttl int) error {
	err := r.Client.Set(context.TODO(), key, value, time.Second*time.Duration(ttl)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisCache) Delete(key string) error {
	err := r.Client.Del(context.TODO(), key).Err()
	if err != nil {
		return err
	}
	return nil
}
