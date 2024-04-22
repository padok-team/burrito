package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/padok-team/burrito/internal/burrito/config"
	storageerrors "github.com/padok-team/burrito/internal/datastore/storage/error"
)

type Storage struct {
	Client *redis.Client
}

func New(config config.Redis) *Storage {
	return &Storage{
		Client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", config.Hostname, config.ServerPort),
			Password: config.Password, // no password set
			DB:       config.Database, // use default DB
		}),
	}
}

func (s *Storage) Get(key string) ([]byte, error) {
	val, err := s.Client.Get(context.TODO(), key).Result()
	if err == redis.Nil {
		return nil, &storageerrors.StorageError{
			Err: err,
			Nil: true,
		}
	}
	if err != nil {
		return nil, &storageerrors.StorageError{
			Err: err,
			Nil: false,
		}
	}
	return []byte(val), nil
}

func (s *Storage) Set(key string, value []byte, ttl int) error {
	err := s.Client.Set(context.TODO(), key, value, time.Second*time.Duration(ttl)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) Delete(key string) error {
	err := s.Client.Del(context.TODO(), key).Err()
	if err != nil {
		return err
	}
	return nil
}
