package controllers

import "errors"

type Cache interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttl int) error
	Delete(key string) error
}

type MemoryCache struct {
	data map[string][]byte
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
