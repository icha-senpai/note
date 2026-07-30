package gws

import (
	"sync"
)

type SessionStorage interface {
	// Returns the number of key-value pairs in the storage
	Len() int

	// retrieves the value for a given key. If the key exists, it returns the value and true; otherwise, it returns nil and false
	Load(key string) (value any, exist bool)

	// removes the key-value pair from the storage for a given key
	Delete(key string)

	// saves the key-value pair in the storage
	Store(key string, value any)

	// If the function returns false, the iteration stops early.
	Range(f func(key string, value any) bool)
}

// creates and returns a new smap instance
func newSmap() *smap {
	return &smap{data: make(map[string]any)}
}

// map-based implementation of the session storage
type smap struct {
	sync.Mutex
	data map[string]any
}

// returns the number of key-value pairs in the storage
func (c *smap) Len() int {
	c.Lock()
	defer c.Unlock()
	return len(c.data)
}

// retrieves the value for a given key. If the key exists, it returns the value and true; otherwise, it returns nil and false
func (c *smap) Load(key string) (value any, exist bool) {
	c.Lock()
	defer c.Unlock()
	value, exist = c.data[key]
	return
}

// removes the key-value pair from the storage for a given key
func (c *smap) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.data, key)
}

// saves the key-value pair in the storage
func (c *smap) Store(key string, value any) {
	c.Lock()
	defer c.Unlock()
	c.data[key] = value
}

func (c *smap) Range(f func(key string, value any) bool) {
	c.Lock()
	defer c.Unlock()

	for k, v := range c.data {
		if !f(k, v) {
			return
		}
	}
}

// concurrency-safe map structure
type ConcurrentMap[K comparable, V any] struct {
	m sync.Map
}

// creates a new concurrency-safe map
func NewConcurrentMap[K comparable, V any]() *ConcurrentMap[K, V] {
	return &ConcurrentMap[K, V]{}
}

// Len returns the number of elements in the map
func (c *ConcurrentMap[K, V]) Len() int {
	var length int
	c.m.Range(func(_, _ any) bool {
		length++
		return true
	})
	return length
}

// returns the value stored in the map for a key, or nil if no value is present
// The ok result indicates whether the value was found in the map
func (c *ConcurrentMap[K, V]) Load(key K) (value V, ok bool) {
	v, ok := c.m.Load(key)
	if !ok {
		return value, false
	}
	return v.(V), true
}

// Delete deletes the value for a key
func (c *ConcurrentMap[K, V]) Delete(key K) {
	c.m.Delete(key)
}

// sets the value for a key
func (c *ConcurrentMap[K, V]) Store(key K, value V) {
	c.m.Store(key, value)
}

// If f returns false, range stops the iteration
func (c *ConcurrentMap[K, V]) Range(f func(key K, value V) bool) {
	c.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}
