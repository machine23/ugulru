package ugulru_test

import (
	"fmt"
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

func TestInMemoryCache_Load(t *testing.T) {
	t.Run("Test loading a new entry", func(t *testing.T) {
		cache := ugulru.NewInMemoryCache[string, int](2, 5*time.Minute)
		loader := func() (int, error) {
			return 1, nil
		}
		value, err := cache.Load("key1", loader)
		assert.NoError(t, err)
		assert.Equal(t, 1, value)
		value, ok := cache.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, 1, value)
	})

	t.Run("Test loading an existing entry", func(t *testing.T) {
		cache := ugulru.NewInMemoryCache[string, int](2, 5*time.Minute)
		loader := func() (int, error) {
			return 1, nil
		}
		cache.Load("key1", loader)
		loader = func() (int, error) {
			return 2, nil
		}
		value, err := cache.Load("key1", loader)
		assert.NoError(t, err)
		assert.Equal(t, 1, value) // should return the cached value
		value, ok := cache.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, 1, value)
	})

	t.Run("Test loading another new entry", func(t *testing.T) {
		cache := ugulru.NewInMemoryCache[string, int](2, 5*time.Minute)
		loader := func() (int, error) {
			return 3, nil
		}
		value, err := cache.Load("key2", loader)
		assert.NoError(t, err)
		assert.Equal(t, 3, value)
		value, ok := cache.Get("key2")
		assert.True(t, ok)
		assert.Equal(t, 3, value)
	})

	t.Run("Test cache capacity limit with Load", func(t *testing.T) {
		cache := ugulru.NewInMemoryCache[string, int](2, 5*time.Minute)

		// Load first entry
		loader1 := func() (int, error) {
			return 1, nil
		}
		cache.Load("key1", loader1)

		// Load second entry
		loader2 := func() (int, error) {
			return 3, nil
		}
		cache.Load("key2", loader2)

		// Load third entry, which should cause the first entry to be evicted
		loader3 := func() (int, error) {
			return 4, nil
		}
		value, err := cache.Load("key3", loader3)
		assert.NoError(t, err)
		assert.Equal(t, 4, value)

		// Verify that the first entry has been evicted
		_, ok := cache.Get("key1")
		assert.False(t, ok, "key1 should be evicted")

		// Verify that the second entry is still in the cache
		value, ok = cache.Get("key2")
		assert.True(t, ok, "key2 should still be in the cache")
		assert.Equal(t, 3, value)

		// Verify that the third entry is in the cache
		value, ok = cache.Get("key3")
		assert.True(t, ok, "key3 should be in the cache")
		assert.Equal(t, 4, value)
	})

	t.Run("Test loader error", func(t *testing.T) {
		cache := ugulru.NewInMemoryCache[string, int](2, 5*time.Minute)
		loader := func() (int, error) {
			return 0, fmt.Errorf("loader error")
		}
		_, err := cache.Load("key4", loader)
		assert.Error(t, err)
		_, ok := cache.Get("key4")
		assert.False(t, ok)
	})
}
