package ugulru_test

import (
	"testing"
	"time"

	"github.com/machine23/ugulru"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryCache_Put(t *testing.T) {
	cache := ugulru.NewInMemoryCache[string, int](2, 5*time.Minute)

	// Test adding a new entry
	cache.Put("key1", 1)
	value, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, 1, value)

	// Test updating an existing entry
	cache.Put("key1", 2)
	value, ok = cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, 2, value)

	// Test adding another entry
	cache.Put("key2", 3)
	value, ok = cache.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, 3, value)

	// Test cache capacity limit
	cache.Put("key3", 4)
	_, ok = cache.Get("key1")
	assert.False(t, ok) // key1 should be evicted
	value, ok = cache.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, 3, value)
	value, ok = cache.Get("key3")
	assert.True(t, ok)
	assert.Equal(t, 4, value)
}

func TestInMemoryCache_Remove(t *testing.T) {
	cache := ugulru.NewInMemoryCache[string, int](2, 5*time.Minute)

	// Test removing an existing entry
	cache.Put("key1", 1)
	cache.Remove("key1")
	_, ok := cache.Get("key1")
	assert.False(t, ok)

	// Test removing a non-existing entry
	cache.Remove("key2") // should not cause any error
	_, ok = cache.Get("key2")
	assert.False(t, ok)

	// Test removing an entry from a full cache
	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Remove("key1")
	_, ok = cache.Get("key1")
	assert.False(t, ok)
	value, ok := cache.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, 2, value)

	// Test removing an entry and adding a new one
	cache.Put("key3", 3)
	value, ok = cache.Get("key3")
	assert.True(t, ok)
	assert.Equal(t, 3, value)
}
