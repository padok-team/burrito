package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/padok-team/burrito/internal/storage"
)

type Storage struct {
	Client *redis.Client
}

func New(addr string, password string, db int) *Storage {
	return &Storage{
		Client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password, // no password set
			DB:       db,       // use default DB
		}),
	}
}

func (s *Storage) Get(key string) ([]byte, error) {
	val, err := s.Client.Get(context.TODO(), key).Result()
	if err == redis.Nil {
		return nil, &storage.StorageError{
			Err: err,
			Nil: true,
		}
	}
	if err != nil {
		return nil, &storage.StorageError{
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
