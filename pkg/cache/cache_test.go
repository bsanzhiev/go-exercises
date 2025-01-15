package cache

import (
	"sync"
	"testing"
	"time"
)

func TestCache_SetGet(t *testing.T) {
	cache := NewCache(time.Second)

	// Test basic set/get
	cache.Set("key1", "value1", 2*time.Second)
	if val, ok := cache.Get("key1"); !ok || val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	// Test non-existent key
	if _, ok := cache.Get("non-existent"); ok {
		t.Error("Expected not found for non-existent key")
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := NewCache(time.Second)

	cache.Set("key1", "value1", 100*time.Millisecond)

	// Check before expiration
	if _, ok := cache.Get("key1"); !ok {
		t.Error("Value should exist before expiration")
	}

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Check after expiration
	if _, ok := cache.Get("key1"); ok {
		t.Error("Value should not exist after expiration")
	}
}

func TestCache_Concurrent(t *testing.T) {
	cache := NewCache(time.Second)
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := string(rune(i))
			cache.Set(key, i, time.Second)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := string(rune(i))
			cache.Get(key)
		}(i)
	}

	wg.Wait()
}

func TestCache_Cleanup(t *testing.T) {
	cache := NewCache(100 * time.Millisecond)

	cache.Set("key1", "value1", 50*time.Millisecond)
	cache.Set("key2", "value2", 300*time.Millisecond)

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	// key1 should be cleaned up, key2 should still exist
	if _, ok := cache.Get("key1"); ok {
		t.Error("key1 should have been cleaned up")
	}
	if _, ok := cache.Get("key2"); !ok {
		t.Error("key2 should still exist")
	}
}
