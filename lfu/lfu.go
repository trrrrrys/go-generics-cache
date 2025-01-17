package lfu

import (
	"container/heap"
)

// Cache is a thread safe LRU cache
type Cache[K comparable, V any] struct {
	cap   int
	queue *priorityQueue[K, V]
	items map[K]*entry[K, V]
}

// Option is an option for LFU cache.
type Option func(*options)

type options struct {
	capacity int
}

func newOptions() *options {
	return &options{
		capacity: 128,
	}
}

// WithCapacity is an option to set cache capacity.
func WithCapacity(cap int) Option {
	return func(o *options) {
		o.capacity = cap
	}
}

// NewCache creates a new LFU cache whose capacity is the default size (128).
func NewCache[K comparable, V any](opts ...Option) *Cache[K, V] {
	o := newOptions()
	for _, optFunc := range opts {
		optFunc(o)
	}
	return &Cache[K, V]{
		cap:   o.capacity,
		queue: newPriorityQueue[K, V](o.capacity),
		items: make(map[K]*entry[K, V], o.capacity),
	}
}

// Get looks up a key's value from the cache.
func (c *Cache[K, V]) Get(key K) (zero V, _ bool) {
	e, ok := c.items[key]
	if !ok {
		return
	}
	e.referenced()
	heap.Fix(c.queue, e.index)
	return e.val, true
}

// Set sets a value to the cache with key. replacing any existing value.
func (c *Cache[K, V]) Set(key K, val V) {
	if e, ok := c.items[key]; ok {
		c.queue.update(e, val)
		return
	}

	if len(c.items) == c.cap {
		evictedEntry := heap.Pop(c.queue).(*entry[K, V])
		delete(c.items, evictedEntry.key)
	}

	e := newEntry(key, val)
	heap.Push(c.queue, e)
	c.items[key] = e
}

// Keys returns the keys of the cache. the order is from oldest to newest.
func (c *Cache[K, V]) Keys() []K {
	keys := make([]K, 0, len(c.items))
	for _, entry := range *c.queue {
		keys = append(keys, entry.key)
	}
	return keys
}

// Delete deletes the item with provided key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	if e, ok := c.items[key]; ok {
		heap.Remove(c.queue, e.index)
		delete(c.items, key)
	}
}

// Len returns the number of items in the cache.
func (c *Cache[K, V]) Len() int {
	return c.queue.Len()
}
