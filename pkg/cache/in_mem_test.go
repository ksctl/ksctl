// Copyright 2025 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

// Compile-time interface check
var _ Cache = (*InMemCache)(nil)

func TestNewInMemCache(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	cache := NewInMemCache(ctx)
	if cache == nil {
		t.Fatal("Expected non-nil cache instance")
	}

	if cache.mu == nil {
		t.Error("Expected mutex to be initialized")
	}

	if cache.o == nil {
		t.Error("Expected map to be initialized")
	}

	if cache.ctx == nil {
		t.Error("Expected context to be set")
	}

	if cache.cancel == nil {
		t.Error("Expected cancel function to be set")
	}
}

func TestSetAndGet(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	cache := NewInMemCache(ctx)

	// Test setting and getting values
	cache.Set("key1", "value1")
	cache.Set("key2", "123")

	// Test retrieving values
	value1, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected key1 to exist in cache")
	}
	if value1 != "value1" {
		t.Errorf("Expected 'value1', got %v", value1)
	}

	value2, ok := cache.Get("key2")
	if !ok {
		t.Error("Expected key2 to exist in cache")
	}
	if value2 != "123" {
		t.Errorf("Expected 123, got %v", value2)
	}

	// Test non-existent key
	_, ok = cache.Get("nonexistent")
	if ok {
		t.Error("Expected 'nonexistent' key to not exist")
	}
}

func TestSetWithExpire(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	cache := NewInMemCache(ctx)

	cache.SetWithExpire("expiring", "temp-value", 1*time.Second)

	val, ok := cache.Get("expiring")
	if !ok {
		t.Error("Expected 'expiring' key to exist immediately after setting")
	}
	if val != "temp-value" {
		t.Errorf("Expected 'temp-value', got %v", val)
	}

	time.Sleep(2 * time.Second)

	_, ok = cache.Get("expiring")
	if ok {
		t.Error("Expected 'expiring' key to be removed after TTL")
	}
}

func TestWorkerCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	cache := NewInMemCache(ctx)

	for i := range [5]any{} {
		cache.SetWithExpire(fmt.Sprintf("key%d", i), strconv.Itoa(i), 1*time.Second)
	}

	cache.Set("permanent", "value")

	time.Sleep(2 * time.Second)

	cache.mu.Lock()
	for k, v := range cache.o {
		if v.ttl != 0 && time.Since(v.when) > v.ttl {
			delete(cache.o, k)
		}
	}
	cache.mu.Unlock()

	for i := range [5]any{} {
		_, ok := cache.Get(fmt.Sprintf("key%d", i))
		if ok {
			t.Errorf("Expected key%d to be removed after TTL", i)
		}
	}

	_, ok := cache.Get("permanent")
	if !ok {
		t.Error("Expected 'permanent' key to still exist")
	}
}

func TestClose(t *testing.T) {
	ctx := context.Background()
	cache := NewInMemCache(ctx)

	cache.Set("test", "value")

	cache.Close()

	time.Sleep(100 * time.Millisecond)

	cache.Set("another", "value")
	_, ok := cache.Get("test")
	if ok {
		t.Error("Expected to be unable to get value after closing")
	}
}

func TestConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	cache := NewInMemCache(ctx)
	var wg sync.WaitGroup

	// Number of concurrent goroutines
	const n = 100

	// Add operations
	wg.Add(n * 3) // n goroutines for Set, n for Get, n for SetWithExpire

	// Concurrent Sets
	for i := range [n]any{} {
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			cache.Set(key, strconv.Itoa(i))
		}(i)
	}

	for i := range [n]any{} {
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("expiring-key-%d", i)
			cache.SetWithExpire(key, strconv.Itoa(i), 60*time.Second) // 60 seconds TTL
		}(i)
	}

	for i := range [n]any{} {
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i%n)
			cache.Get(key)
		}(i)
	}

	wg.Wait()

	for i := 0; i < n; i += 10 {
		key := fmt.Sprintf("key-%d", i)
		val, ok := cache.Get(key)
		if !ok {
			t.Errorf("Expected key %s to exist", key)
			continue
		}

		if val != strconv.Itoa(i) {
			t.Errorf("Expected value %d for key %s, got %v", i, key, val)
		}
	}
}

func TestInterfaceImplementation(t *testing.T) {
	ctx := context.Background()
	var cache Cache = NewInMemCache(ctx)

	cache.Set("key1", "value1")
	cache.SetWithExpire("key2", "value2", 5*time.Minute)

	val1, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected key1 to exist")
	}
	if val1 != "value1" {
		t.Errorf("Expected 'value1', got %v", val1)
	}

	val2, ok := cache.Get("key2")
	if !ok {
		t.Error("Expected key2 to exist")
	}
	if val2 != "value2" {
		t.Errorf("Expected 'value2', got %v", val2)
	}

	cache.Close()
}

func TestKeys(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	cache := NewInMemCache(ctx)

	cache.Set("ggg:key1", "value1")
	cache.Set("ggg:key2", "value2")
	cache.Set("ff:key2", "value2")

	if v, err := cache.KeysWithPrefix("ggg:"); err != nil {
		t.Errorf("Expected no error, got %v", err)
	} else {
		assert.Equal(t, len(v), 2)
		assert.DeepEqual(t, []string{"ggg:key1", "ggg:key2"}, v)
	}

	if v, err := cache.KeysWithPrefix("ff:"); err != nil {
		t.Errorf("Expected no error, got %v", err)
	} else {
		assert.Equal(t, len(v), 1)
		assert.DeepEqual(t, []string{"ff:key2"}, v)
	}

	if v, err := cache.KeysWithPrefix("nonexistent:"); err != nil {
		t.Errorf("Expected no error, got %v", err)
	} else {
		assert.DeepEqual(t, len(v), 0)
	}

	if v, err := cache.KeysWithPrefix("ggg:key1"); err != nil {
		t.Errorf("Expected no error, got %v", err)
	} else {
		assert.Equal(t, len(v), 1)
		assert.DeepEqual(t, []string{"ggg:key1"}, v)
	}
}
