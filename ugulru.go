package ugulru

import (
	"container/list"
	"sync"
	"time"
)

type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Put(key K, value V)
	Remove(key K)
	RemoveExpired()
	Load(key K, loader func() (V, error)) (V, error)
}

type InMemoryCache[K comparable, V any] struct {
	cache    map[K]*list.Element
	list     *list.List
	capacity int
	ttl      time.Duration
	mu       sync.Mutex
}

type entry[K comparable, V any] struct {
	key       K
	value     V
	timestamp time.Time
}

func NewInMemoryCache[K comparable, V any](capacity int, ttl time.Duration) *InMemoryCache[K, V] {
	return &InMemoryCache[K, V]{
		cache:    make(map[K]*list.Element),
		list:     list.New(),
		capacity: capacity,
		ttl:      ttl,
	}
}

func (c *InMemoryCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var zero V
	if elem, ok := c.cache[key]; ok {
		entry := elem.Value.(*entry[K, V])
		if time.Since(entry.timestamp) > c.ttl {
			c.list.Remove(elem)
			return zero, false
		}
		c.list.MoveToFront(elem)
		return entry.value, true
	}
	return zero, false
}

func (c *InMemoryCache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		entry := elem.Value.(*entry[K, V])
		entry.value = value
		entry.timestamp = time.Now()
		c.list.MoveToFront(elem)
		return
	}

	if c.list.Len() >= c.capacity {
		elem := c.list.Back()
		entry := elem.Value.(*entry[K, V])
		delete(c.cache, entry.key)
		c.list.Remove(elem)
	}

	entry := &entry[K, V]{key: key, value: value, timestamp: time.Now()}
	elem := c.list.PushFront(entry)
	c.cache[key] = elem
}

func (c *InMemoryCache[K, V]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		delete(c.cache, key)
		c.list.Remove(elem)
	}
}

func (c *InMemoryCache[K, V]) Load(key K, loader func() (V, error)) (V, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		entry := elem.Value.(*entry[K, V])
		if time.Since(entry.timestamp) > c.ttl {
			c.list.Remove(elem)
		} else {
			c.list.MoveToFront(elem)
			return entry.value, nil
		}
	}

	value, err := loader()
	if err != nil {
		return value, err
	}

	if c.list.Len() >= c.capacity {
		elem := c.list.Back()
		entry := elem.Value.(*entry[K, V])
		delete(c.cache, entry.key)
		c.list.Remove(elem)
	}

	entry := &entry[K, V]{key: key, value: value, timestamp: time.Now()}
	elem := c.list.PushFront(entry)
	c.cache[key] = elem

	return value, nil
}

func (c *InMemoryCache[K, V]) RemoveExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for elem := c.list.Back(); elem != nil; elem = elem.Prev() {
		entry := elem.Value.(*entry[K, V])
		if time.Since(entry.timestamp) > c.ttl {
			delete(c.cache, entry.key)
			c.list.Remove(elem)
		} else {
			break
		}
	}
}
