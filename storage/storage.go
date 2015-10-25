package storage

import "sync"

type Store interface {
	Set(key string, val interface{}) error
	Get(key string) (interface{}, error)
	Del(key string) error
}

func NewStore() Store {
	return newMemory()
}

func newMemory() *memory {
	return &memory{
		m: make(map[string]interface{}),
	}
}

type memory struct {
	m map[string]interface{}
	sync.RWMutex
}

func (m *memory) Set(key string, val interface{}) error {
	m.Lock()
	defer m.Unlock()
	m.m[key] = val
	return nil
}

func (m *memory) Get(key string) (interface{}, error) {
	m.RLock()
	defer m.RUnlock()
	return m.m[key], nil
}

func (m *memory) Del(key string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.m, key)
	return nil
}
