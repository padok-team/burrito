package mock

import (
	"fmt"
	"strings"

	errors "github.com/padok-team/burrito/internal/datastore/storage/error"
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
		return nil, &errors.StorageError{
			Err: fmt.Errorf("%s", "Not found"),
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

func (a *Mock) List(prefix string) ([]string, error) {
	keySet := map[string]bool{}
	for k := range a.data {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		pathIndexs := strings.Split(k, "/")
		keySet[pathIndexs[len(pathIndexs)-2]] = true
	}
	return mapKeys(keySet), nil
}

func mapKeys(m map[string]bool) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
