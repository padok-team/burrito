package controllers

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttl int) error
	Delete(key string) error
}

type MemoryCache struct {
	data map[string][]byte
}

type RedisCache struct {
	Client *redis.Client
}

func newRedisCache(addr string, password string, db int) *RedisCache {
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
	if err != nil {
		return nil, err
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
	return nil
}

func newMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: map[string][]byte{},
	}
}

func (m *MemoryCache) Get(key string) ([]byte, error) {
	if _, ok := m.data[key]; !ok {
		return nil, errors.New("key not found")
	}
	return m.data[key], nil
}

func (m *MemoryCache) Set(key string, value []byte, ttl int) error {
	m.data[key] = value
	return nil
}

func (m *MemoryCache) Delete(key string) error {
	delete(m.data, key)
	return nil
}
