package lru

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestLRUCache(t *testing.T) {
	t.Run("Basic operations", func(t *testing.T) {
		cache := New[string, []byte](100, func(k string, v []byte) int { return len(k) + len(v) })

		cache.Set("key1", []byte("value1"))
		if val, ok := cache.Get("key1"); !ok || !bytes.Equal(val, []byte("value1")) {
			t.Errorf("Expected value1, got %s", val)
		}

		cache.Set("key1", []byte("newvalue1"))
		if val, ok := cache.Get("key1"); !ok || !bytes.Equal(val, []byte("newvalue1")) {
			t.Errorf("Expected newvalue1, got %s", val)
		}

		if _, ok := cache.Get("nonexistent"); ok {
			t.Error("Expected no value for nonexistent key")
		}

		cache.Remove("key1")
		if _, ok := cache.Get("key1"); ok {
			t.Error("Expected key1 to be removed")
		}
	})

	t.Run("Capacity and eviction", func(t *testing.T) {
		cache := New[string, []byte](30, func(k string, v []byte) int { return len(k) + len(v) })

		cache.Set("key1", []byte("12345"))
		cache.Set("key2", []byte("67890"))
		cache.Set("key3", []byte("abcde"))

		if cache.Len() != 3 {
			t.Errorf("Expected 3 items, got %d", cache.Len())
		}

		cache.Set("key4", []byte("fghij"))

		if _, ok := cache.Get("key1"); ok {
			t.Error("Expected key1 to be evicted")
		}

		if cache.Len() != 3 {
			t.Errorf("Expected 3 items after eviction, got %d", cache.Len())
		}

		if size := cache.ByteSize(); size != 27 {
			t.Errorf("Expected byte size of 27, got %d", size)
		}
	})

	t.Run("LRU behavior", func(t *testing.T) {
		cache := New[string, []byte](35, func(k string, v []byte) int { return len(k) + len(v) })

		cache.Set("key1", []byte("12345"))
		cache.Set("key2", []byte("67890"))
		cache.Set("key3", []byte("abcde"))

		cache.Get("key1")

		cache.Set("key4", []byte("fghij"))

		if _, ok := cache.Get("key2"); ok {
			t.Error("Expected key2 to be evicted")
		}

		if _, ok := cache.Get("key1"); !ok {
			t.Error("Expected key1 to still be in cache")
		}
	})

	t.Run("Large values", func(t *testing.T) {
		cache := New[string, []byte](100, func(k string, v []byte) int { return len(k) + len(v) })

		largeValue := make([]byte, 200)
		rand.Read(largeValue)

		cache.Set("large", largeValue)

		if cache.Len() != 0 {
			t.Error("Expected cache to be empty after inserting too large value")
		}

		if size := cache.ByteSize(); size != 0 {
			t.Errorf("Expected byte size of 0, got %d", size)
		}
	})

	t.Run("Concurrent access", func(t *testing.T) {
		cache := New[string, []byte](1000, func(k string, v []byte) int { return len(k) + len(v) })
		done := make(chan bool)
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					key := fmt.Sprintf("key%d-%d", id, j)
					value := []byte(fmt.Sprintf("value%d-%d", id, j))
					cache.Set(key, value)
					_, ok := cache.Get(key)
					if !ok {
						t.Logf("Failed to get key %s", key)
					}
				}
			}(i)
		}

		go func() {
			wg.Wait()
			done <- true
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}

		if cache.Len() > 100 {
			t.Errorf("Expected at most 100 items due to eviction, got %d", cache.Len())
		}
	})

	t.Run("Edge cases", func(t *testing.T) {
		cache := New[string, []byte](25, func(k string, v []byte) int { return len(k) + len(v) })

		cache.Set("key1", []byte("1234567890"))
		if cache.Len() != 1 {
			t.Errorf("Expected 1 item, got %d", cache.Len())
		}
		if size := cache.ByteSize(); size != 14 {
			t.Errorf("Expected byte size of 14, got %d", size)
		}

		cache.Set("key2", []byte("12345678901"))
		if cache.Len() != 1 || cache.ByteSize() != 15 {
			t.Errorf("Expected 1 item and 15 bytes, got %d items and %d bytes", cache.Len(), cache.ByteSize())
		}

		cache.Set("a", []byte("1"))
		cache.Set("b", []byte("2"))
		cache.Set("c", []byte("3"))
		cache.Set("d", []byte("4"))
		if cache.Len() != 5 || cache.ByteSize() != 23 {
			t.Errorf("Expected 5 items and 23 bytes, got %d items and %d bytes", cache.Len(), cache.ByteSize())
		}
	})
}

func BenchmarkLRUCache(b *testing.B) {
	cache := New[string, []byte](10*1024*1024, func(k string, v []byte) int { return len(k) + len(v) })

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key%d", i)
			value := []byte(fmt.Sprintf("value%d", i))
			cache.Set(key, value)
		}
	})

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key%d", i%1000)
			cache.Get(key)
		}
	})

	b.Run("Set and Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key%d", i%1000)
			value := []byte(fmt.Sprintf("value%d", i))
			cache.Set(key, value)
			cache.Get(key)
		}
	})
}
