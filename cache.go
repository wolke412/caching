package caching

import (
	"errors"
	"sync"
)

type Cache[T any] struct {
	data map[int]T
	mu   sync.RWMutex
}

// TODO: add preallocation
func NewCache[T any]() *Cache[T] {
    
	return &Cache[T] {
		data: make(map[int]T),
        mu: sync.RWMutex{},
	}
}

func (c *Cache[T]) Add(key int, value T) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.data[key] = value
}

func (c *Cache[T]) AddMultiple( entries map[int]T ) {
	c.mu.Lock()

	defer c.mu.Unlock()

	for key, value := range entries {
		c.data[key] = value
	}
}

func (c *Cache[T]) Update( key int, value T ) error {
	c.mu.Lock()

	defer c.mu.Unlock()

	if _, exists := c.data[key]; !exists {
		return errors.New("key not found")
	}

	c.data[key] = value

	return nil
}

func (c *Cache[T]) Upsert( key int, value T ) error {

    if c.Exists(key) {
        return c.Update( key, value )            
    }

    c.Add( key, value )

    return nil
}

func (c *Cache[T]) Delete( keys ...int ) error {
	c.mu.Lock()

	defer c.mu.Unlock()
    
    for key := range keys {

        if c.Exists(key) {
            delete(c.data, key)
        }
    }

	return nil
}

func (c *Cache[T]) Exists( key int ) bool {

	c.mu.RLock()

	defer c.mu.RUnlock()

	_, exists := c.data[key];

    return exists 
}

func (c *Cache[T]) Find( key int ) (T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if value, exists := c.data[key]; exists {
		return value, nil
	}

	return *new(T), errors.New("key not found")
}

func (c *Cache[T]) Get() *map[int]T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &c.data
}

func (c *Cache[T]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

    clear( c.data )
}
