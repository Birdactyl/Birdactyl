package tests

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"birdactyl-panel-backend/internal/services"
)

func TestCacheSetAndGet(t *testing.T) {
	c := services.NewMemoryCache()
	key := "test_key"
	value := "test_value"

	c.Set(key, value, 5*time.Minute)

	t.Run("Get Hit", func(t *testing.T) {
		v, found := c.Get(key)
		if !found {
			t.Fatalf("Expected to find key %s in cache", key)
		}
		if v != value {
			t.Errorf("Expected value %v, got %v", value, v)
		}
	})

	t.Run("Get Miss", func(t *testing.T) {
		_, found := c.Get("non_existent_key")
		if found {
			t.Errorf("Expected not to find non_existent_key in cache")
		}
	})
}

func TestCacheExpiration(t *testing.T) {
	c := services.NewMemoryCache()
	key := "exp_key"

	c.Set(key, "value", -1*time.Second)

	_, found := c.Get(key)
	if found {
		t.Errorf("Expected key %s to be expired and not found", key)
	}
}

func TestCacheDelete(t *testing.T) {
	c := services.NewMemoryCache()
	key := "del_key"

	c.Set(key, "value", 5*time.Minute)
	c.Delete(key)

	_, found := c.Get(key)
	if found {
		t.Errorf("Expected key %s to be deleted", key)
	}
}

func TestCacheDeletePrefix(t *testing.T) {
	c := services.NewMemoryCache()
	c.Set("prefix_1", "v1", 5*time.Minute)
	c.Set("prefix_2", "v2", 5*time.Minute)
	c.Set("other_1", "v3", 5*time.Minute)

	c.DeletePrefix("prefix_")

	if _, found := c.Get("prefix_1"); found {
		t.Errorf("Expected prefix_1 to be deleted")
	}
	if _, found := c.Get("prefix_2"); found {
		t.Errorf("Expected prefix_2 to be deleted")
	}
	if _, found := c.Get("other_1"); !found {
		t.Errorf("Expected other_1 to remain in cache")
	}
}

func TestCacheConcurrency(t *testing.T) {
	c := services.NewMemoryCache()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.Set(fmt.Sprintf("key_%d", n), n, 5*time.Minute)
			c.Get(fmt.Sprintf("key_%d", n))
		}(i)
	}
	wg.Wait()
}
