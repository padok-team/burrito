package mock

import (
	"errors"

	"github.com/padok-team/burrito/internal/storage"
)

type Storage struct {
	data map[string][]byte
}

func New() *Storage {
	return &Storage{
		data: make(map[string][]byte),
	}
}

func (s *Storage) Get(key string) ([]byte, error) {
	val, ok := s.data[key]
	if !ok {
		return nil, &storage.StorageError{
			Err: errors.New("key not found"),
			Nil: true,
		}
	}
	return val, nil
}

func (s *Storage) Set(key string, value []byte, ttl int) error {
	s.data[key] = value
	return nil
}

func (s *Storage) Delete(key string) error {
	delete(s.data, key)
	return nil
}
