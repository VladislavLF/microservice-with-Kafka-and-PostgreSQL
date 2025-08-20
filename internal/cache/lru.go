package cache

import (
	"container/list"
	"sync"
	"time"
)

type entry[K comparable, V any] struct {
	key       K
	value     V
	expiresAt time.Time
}

type LRU[K comparable, V any] struct {
	mu       sync.RWMutex
	capacity int
	ttl      time.Duration
	ll       *list.List
	table    map[K]*list.Element
}

func NewLRU[K comparable, V any](capacity int, ttl time.Duration) *LRU[K, V] {
	if capacity <= 0 {
		capacity = 1
	}
	return &LRU[K, V]{
		capacity: capacity,
		ttl:      ttl,
		ll:       list.New(),
		table:    make(map[K]*list.Element, capacity),
	}
}

func (c *LRU[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	elem, ok := c.table[key]
	if !ok {
		c.mu.RUnlock()
		var zero V
		return zero, false
	}
	ent := elem.Value.(entry[K, V])
	if time.Now().After(ent.expiresAt) {
		c.mu.RUnlock()
		c.Delete(key)
		var zero V
		return zero, false
	}
	c.mu.RUnlock()
	c.mu.Lock()
	c.ll.MoveToFront(elem)
	c.mu.Unlock()
	return ent.value, true
}

func (c *LRU[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.table[key]; ok {
		elem.Value = entry[K, V]{key: key, value: value, expiresAt: time.Now().Add(c.ttl)}
		c.ll.MoveToFront(elem)
		return
	}

	elem := c.ll.PushFront(entry[K, V]{key: key, value: value, expiresAt: time.Now().Add(c.ttl)})
	c.table[key] = elem

	if c.ll.Len() > c.capacity {
		oldest := c.ll.Back()
		if oldest != nil {
			ent := oldest.Value.(entry[K, V])
			delete(c.table, ent.key)
			c.ll.Remove(oldest)
		}
	}
}

func (c *LRU[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.table[key]; ok {
		delete(c.table, key)
		c.ll.Remove(elem)
	}
}

func (c *LRU[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ll.Len()
}
