package benchmarks

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"birdactyl-panel-backend/internal/services"
)

func BenchmarkCacheSet(b *testing.B) {
	c := services.NewMemoryCache()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(fmt.Sprintf("key_%d", i), "value", 5*time.Minute)
	}
}

func BenchmarkCacheGet_Hit(b *testing.B) {
	c := services.NewMemoryCache()
	for i := 0; i < 1000; i++ {
		c.Set(fmt.Sprintf("key_%d", i), "value", 5*time.Minute)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(fmt.Sprintf("key_%d", i%1000))
	}
}

func BenchmarkCacheGet_Miss(b *testing.B) {
	c := services.NewMemoryCache()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(fmt.Sprintf("miss_%d", i))
	}
}

func BenchmarkCacheGet_Expired(b *testing.B) {
	c := services.NewMemoryCache()
	for i := 0; i < 1000; i++ {
		c.Set(fmt.Sprintf("key_%d", i), "value", -1*time.Second)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(fmt.Sprintf("key_%d", i%1000))
	}
}

func BenchmarkCacheDelete(b *testing.B) {
	c := services.NewMemoryCache()
	for i := 0; i < b.N; i++ {
		c.Set(fmt.Sprintf("key_%d", i), "value", 5*time.Minute)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Delete(fmt.Sprintf("key_%d", i))
	}
}

func BenchmarkCacheDeletePrefix(b *testing.B) {
	c := services.NewMemoryCache()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		for j := 0; j < 100; j++ {
			c.Set(fmt.Sprintf("prefix_%d_%d", i, j), "value", 5*time.Minute)
		}
		for j := 0; j < 20; j++ {
			c.Set(fmt.Sprintf("other_%d_%d", i, j), "value", 5*time.Minute)
		}
		b.StartTimer()
		c.DeletePrefix(fmt.Sprintf("prefix_%d_", i))
	}
}

func BenchmarkCacheCleanup(b *testing.B) {
	for _, count := range []int{100, 1000, 5000} {
		b.Run(fmt.Sprintf("items_%d", count), func(b *testing.B) {
			c := services.NewMemoryCache()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				for j := 0; j < count; j++ {
					if j%2 == 0 {
						c.Set(fmt.Sprintf("key_%d", j), "value", -1*time.Second)
					} else {
						c.Set(fmt.Sprintf("key_%d", j), "value", 5*time.Minute)
					}
				}
				b.StartTimer()
				c.Cleanup()
			}
		})
	}
}

func BenchmarkCacheParallel_ReadHeavy(b *testing.B) {
	c := services.NewMemoryCache()
	for i := 0; i < 1000; i++ {
		c.Set(fmt.Sprintf("key_%d", i), "value", 5*time.Minute)
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Get(fmt.Sprintf("key_%d", i%1000))
			i++
		}
	})
}

func BenchmarkCacheParallel_WriteHeavy(b *testing.B) {
	c := services.NewMemoryCache()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Set(fmt.Sprintf("key_%d", i), "value", 5*time.Minute)
			i++
		}
	})
}

func BenchmarkCacheParallel_Mixed(b *testing.B) {
	c := services.NewMemoryCache()
	for i := 0; i < 1000; i++ {
		c.Set(fmt.Sprintf("key_%d", i), "value", 5*time.Minute)
	}
	b.ReportAllocs()
	b.ResetTimer()
	var counter int64
	var mu sync.Mutex
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			counter++
			n := counter
			mu.Unlock()
			if n%3 == 0 {
				c.Set(fmt.Sprintf("key_%d", n), "value", 5*time.Minute)
			} else if n%3 == 1 {
				c.Get(fmt.Sprintf("key_%d", n%1000))
			} else {
				c.Delete(fmt.Sprintf("key_%d", n%500))
			}
		}
	})
}
