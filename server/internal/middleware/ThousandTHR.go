package middleware

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	numShards        = 256
	maxBucketsTotal  = 1000000
	maxBucketsPerShard = maxBucketsTotal / numShards
	cleanupBatchSize = 1000
	cleanupInterval  = 30 * time.Second
	bucketExpiry     = 5 * time.Minute
)

type RateLimitBucket struct {
	tokens         float64
	maxTokens      float64
	refillRate     float64
	lastRefillTime time.Time
	lastAccessTime time.Time
	key            string
	element        *list.Element
}

type shard struct {
	buckets map[string]*RateLimitBucket
	lru     *list.List
	mu      sync.Mutex
}

type ProxyTrust int

const (
	TrustNone       ProxyTrust = iota
	TrustCloudflare
	TrustProxy
	TrustAll
)

type ThousandTHRConfig struct {
	RequestsPerMinute int
	BurstLimit        int
	SkipFailedRequest bool
	KeyGenerator      func(*fiber.Ctx) string
	ProxyTrust        ProxyTrust
}

type rateLimitManager struct {
	shards    [numShards]*shard
	stopCh    chan struct{}
	running   bool
	lifecycle sync.Mutex
}

var rateLimiter *rateLimitManager

func init() {
	rateLimiter = &rateLimitManager{}
	for i := 0; i < numShards; i++ {
		rateLimiter.shards[i] = &shard{
			buckets: make(map[string]*RateLimitBucket),
			lru:     list.New(),
		}
	}
}

func ThousandTHR(config ThousandTHRConfig) fiber.Handler {
	if config.RequestsPerMinute <= 0 {
		config.RequestsPerMinute = 60
	}
	if config.BurstLimit <= 0 {
		config.BurstLimit = config.RequestsPerMinute
	}
	if config.KeyGenerator == nil {
		config.KeyGenerator = createKeyGenerator(config.ProxyTrust)
	}

	configHash := generateConfigHash(config)
	refillRate := float64(config.RequestsPerMinute) / 60.0

	return func(c *fiber.Ctx) error {
		key := config.KeyGenerator(c) + ":" + configHash

		allowed, remaining, resetIn := checkRateLimit(key, config, refillRate)

		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetIn))

		if !allowed {
			c.Set("Retry-After", fmt.Sprintf("%d", resetIn))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":        429,
					"message":     "Rate limit exceeded",
					"retry_after": resetIn,
				},
			})
		}

		err := c.Next()

		if config.SkipFailedRequest && c.Response().StatusCode() >= 400 {
			refundToken(key)
		}

		return err
	}
}

func createKeyGenerator(trust ProxyTrust) func(*fiber.Ctx) string {
	return func(c *fiber.Ctx) string {
		ip := c.IP()

		switch trust {
		case TrustCloudflare:
			if cfIP := c.Get("CF-Connecting-IP"); cfIP != "" {
				ip = cfIP
			}
		case TrustProxy:
			if realIP := c.Get("X-Real-IP"); realIP != "" {
				ip = realIP
			}
		case TrustAll:
			if cfIP := c.Get("CF-Connecting-IP"); cfIP != "" {
				ip = cfIP
			} else if realIP := c.Get("X-Real-IP"); realIP != "" {
				ip = realIP
			}
		}

		return ip + ":" + c.Method() + ":" + c.Path()
	}
}

func generateConfigHash(config ThousandTHRConfig) string {
	data := fmt.Sprintf("%d:%d", config.RequestsPerMinute, config.BurstLimit)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

func getShard(key string) *shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return rateLimiter.shards[h.Sum32()%numShards]
}

func checkRateLimit(key string, config ThousandTHRConfig, refillRate float64) (allowed bool, remaining int, resetIn int) {
	s := getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	bucket, exists := s.buckets[key]

	if !exists {
		if len(s.buckets) >= maxBucketsPerShard {
			s.evictOldest()
		}

		bucket = &RateLimitBucket{
			tokens:         float64(config.BurstLimit),
			maxTokens:      float64(config.BurstLimit),
			refillRate:     refillRate,
			lastRefillTime: now,
			lastAccessTime: now,
			key:            key,
		}
		bucket.element = s.lru.PushFront(bucket)
		s.buckets[key] = bucket
	} else {
		s.lru.MoveToFront(bucket.element)
	}

	elapsed := now.Sub(bucket.lastRefillTime).Seconds()
	if elapsed > 0 && bucket.refillRate > 0 {
		bucket.tokens = minFloat(bucket.tokens+elapsed*bucket.refillRate, bucket.maxTokens)
		bucket.lastRefillTime = now
	}
	bucket.lastAccessTime = now

	remaining = int(bucket.tokens)
	if bucket.refillRate > 0 && bucket.tokens < bucket.maxTokens {
		resetIn = int((bucket.maxTokens - bucket.tokens) / bucket.refillRate)
	}

	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true, int(bucket.tokens), resetIn
	}

	if bucket.refillRate > 0 {
		resetIn = int(1.0 / bucket.refillRate)
	} else {
		resetIn = 60
	}
	return false, 0, resetIn
}

func (s *shard) evictOldest() {
	if s.lru.Len() == 0 {
		return
	}
	oldest := s.lru.Back()
	if oldest != nil {
		bucket := oldest.Value.(*RateLimitBucket)
		s.lru.Remove(oldest)
		delete(s.buckets, bucket.key)
	}
}

func refundToken(key string) {
	s := getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	if bucket, exists := s.buckets[key]; exists {
		bucket.tokens = minFloat(bucket.tokens+1.0, bucket.maxTokens)
	}
}

func CleanupRateLimitStore() {
	rateLimiter.lifecycle.Lock()
	defer rateLimiter.lifecycle.Unlock()

	if rateLimiter.running {
		return
	}

	rateLimiter.stopCh = make(chan struct{})
	rateLimiter.running = true

	go cleanupWorker()
}

func cleanupWorker() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for i := 0; i < numShards; i++ {
				select {
				case <-rateLimiter.stopCh:
					return
				default:
					cleanupShard(rateLimiter.shards[i])
				}
			}
		case <-rateLimiter.stopCh:
			return
		}
	}
}

func cleanupShard(s *shard) {
	now := time.Now()
	var toDelete []string

	s.mu.Lock()
	count := 0
	for e := s.lru.Back(); e != nil && count < cleanupBatchSize; {
		bucket := e.Value.(*RateLimitBucket)
		if now.Sub(bucket.lastAccessTime) > bucketExpiry {
			toDelete = append(toDelete, bucket.key)
			prev := e.Prev()
			s.lru.Remove(e)
			e = prev
		} else {
			break
		}
		count++
	}

	for _, key := range toDelete {
		delete(s.buckets, key)
	}
	s.mu.Unlock()
}

func StopCleanup() {
	rateLimiter.lifecycle.Lock()
	defer rateLimiter.lifecycle.Unlock()

	if !rateLimiter.running {
		return
	}

	close(rateLimiter.stopCh)
	rateLimiter.running = false
}

func GetBucketCount() int {
	total := 0
	for i := 0; i < numShards; i++ {
		s := rateLimiter.shards[i]
		s.mu.Lock()
		total += len(s.buckets)
		s.mu.Unlock()
	}
	return total
}

func GetShardStats() []int {
	stats := make([]int, numShards)
	for i := 0; i < numShards; i++ {
		s := rateLimiter.shards[i]
		s.mu.Lock()
		stats[i] = len(s.buckets)
		s.mu.Unlock()
	}
	return stats
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
