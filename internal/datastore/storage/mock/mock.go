package mock

import (
	"errors"

	"github.com/padok-team/burrito/internal/datastore/storage"
)

type Mock struct {
	data map[string][]byte
}

func New() *Mock {
	return &Mock{
		data: make(map[string][]byte),
	}
}

func (s *Mock) Get(key string) ([]byte, error) {
	val, ok := s.data[key]
	if !ok {
		return nil, &storage.StorageError{
			Err: errors.New("key not found"),
			Nil: true,
		}
	}
	return val, nil
}

func (s *Mock) Set(key string, value []byte, ttl int) error {
	s.data[key] = value
	return nil
}

func (s *Mock) Delete(key string) error {
	delete(s.data, key)
	return nil
}

func (a *Mock) List(string) ([]string, error) {
	keys := make([]string, len(a.data))
	i := 0
	for k := range a.data {
		keys[i] = k
		i++
	}
	return keys, nil
}
