// nolint
package api_test

import "errors"

// errBackend is a test-local storage.StorageBackend implementation that
// always returns a plain (non-StorageError) error, to exercise the generic
// "issue with the storage backend" branches that the mock backend
// (internal/datastore/storage/mock) can never trigger, since it only ever
// returns a Nil StorageError (mapped to 404 Not Found).
type errBackend struct{}

func (e *errBackend) Get(key string) ([]byte, error) {
	return nil, errors.New("generic backend error")
}

func (e *errBackend) Check(key string) ([]byte, error) {
	return nil, errors.New("generic backend error")
}

func (e *errBackend) Set(key string, value []byte, ttl int) error {
	return errors.New("generic backend error")
}

func (e *errBackend) Delete(key string) error {
	return errors.New("generic backend error")
}

func (e *errBackend) List(prefix string) ([]string, error) {
	return nil, errors.New("generic backend error")
}

func (e *errBackend) ListRecursive(prefix string) ([]string, error) {
	return nil, errors.New("generic backend error")
}
