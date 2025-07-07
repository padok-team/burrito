package mock

import (
	"fmt"
	"strings"

	errors "github.com/padok-team/burrito/internal/datastore/storage/error"
	"github.com/padok-team/burrito/internal/datastore/storage/utils"
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
	key = "/" + utils.SanitizePrefix(key)
	key = strings.TrimSuffix(key, "/")
	val, ok := s.data[key]
	if !ok {
		return nil, &errors.StorageError{
			Err: fmt.Errorf("object %s not found", key),
			Nil: true,
		}
	}
	return val, nil
}

func (s *Mock) Set(key string, value []byte, ttl int) error {
	key = "/" + utils.SanitizePrefix(key)
	key = strings.TrimSuffix(key, "/")
	s.data[key] = value
	return nil
}

func (s *Mock) Check(key string) ([]byte, error) {
	key = "/" + utils.SanitizePrefix(key)
	key = strings.TrimSuffix(key, "/")
	val, ok := s.data[key]
	if !ok {
		return nil, &errors.StorageError{
			Err: fmt.Errorf("object %s not found", key),
			Nil: true,
		}
	}
	return val, nil
}

func (s *Mock) Delete(key string) error {
	key = "/" + utils.SanitizePrefix(key)
	key = strings.TrimSuffix(key, "/")
	_, ok := s.data[key]
	if !ok {
		return &errors.StorageError{
			Err: fmt.Errorf("%s not found", key),
			Nil: true,
		}
	}
	delete(s.data, key)
	return nil
}

func (a *Mock) List(prefix string) ([]string, error) {
	listPrefix := fmt.Sprintf("/%s", utils.SanitizePrefix(prefix))
	keySet := map[string]bool{}
	found := false

	for k := range a.data {
		if !strings.HasPrefix(k, listPrefix) {
			continue
		}
		found = true

		// Extract the folder part from the path that's one level deeper than the prefix
		parts := strings.Split(strings.TrimPrefix(k, listPrefix), "/")
		if len(parts) > 0 {
			folderPath := listPrefix + parts[0]
			keySet[folderPath] = true
		}
	}

	// Return an error if no keys match the prefix
	if !found {
		return nil, &errors.StorageError{
			Err: fmt.Errorf("prefix %s not found", listPrefix),
			Nil: true,
		}
	}

	return mapKeys(keySet), nil
}

// ListRecursive recursively lists all files under a prefix
func (a *Mock) ListRecursive(prefix string) ([]string, error) {
	listPrefix := fmt.Sprintf("/%s", utils.SanitizePrefix(prefix))
	var keys []string
	found := false

	for k := range a.data {
		if strings.HasPrefix(k, listPrefix) {
			keys = append(keys, k)
			found = true
		}
	}

	// Return an error if no keys match the prefix
	if !found {
		return nil, &errors.StorageError{
			Err: fmt.Errorf("prefix %s not found", listPrefix),
			Nil: true,
		}
	}

	return keys, nil
}

func mapKeys(m map[string]bool) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
