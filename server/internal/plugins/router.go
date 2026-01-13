package plugins

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"birdactyl-panel-backend/internal/models"
	pb "birdactyl-panel-backend/internal/plugins/proto"

	"github.com/gofiber/fiber/v2"
)

type Route struct {
	PluginID string
	Method   string
	Path     string
}

var routes []Route

type RateLimitPreset struct {
	RequestsPerMinute int
	BurstLimit        int
}

var presetConfigs = map[string]RateLimitPreset{
	"read": {
		RequestsPerMinute: 60,
		BurstLimit:        80,
	},
	"write": {
		RequestsPerMinute: 30,
		BurstLimit:        40,
	},
	"strict": {
		RequestsPerMinute: 10,
		BurstLimit:        15,
	},
}

const (
	pluginRateLimitShards    = 64
	pluginRateLimitMaxPerShard = 10000
)

type pluginRateBucket struct {
	tokens         float64
	maxTokens      float64
	refillRate     float64
	lastRefillTime time.Time
	key            string
	element        *list.Element
}

type pluginRateShard struct {
	buckets map[string]*pluginRateBucket
	lru     *list.List
	mu      sync.Mutex
}

var pluginRateLimiter struct {
	shards [pluginRateLimitShards]*pluginRateShard
	once   sync.Once
}

func initPluginRateLimiter() {
	pluginRateLimiter.once.Do(func() {
		for i := 0; i < pluginRateLimitShards; i++ {
			pluginRateLimiter.shards[i] = &pluginRateShard{
				buckets: make(map[string]*pluginRateBucket),
				lru:     list.New(),
			}
		}
	})
}

func getPluginRateShard(key string) *pluginRateShard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return pluginRateLimiter.shards[h.Sum32()%pluginRateLimitShards]
}

func checkPluginRateLimit(key string, rpm, burst int) (allowed bool, remaining int, resetIn int) {
	initPluginRateLimiter()

	s := getPluginRateShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	refillRate := float64(rpm) / 60.0
	bucket, exists := s.buckets[key]

	if !exists {
		if len(s.buckets) >= pluginRateLimitMaxPerShard {
			if oldest := s.lru.Back(); oldest != nil {
				ob := oldest.Value.(*pluginRateBucket)
				s.lru.Remove(oldest)
				delete(s.buckets, ob.key)
			}
		}

		bucket = &pluginRateBucket{
			tokens:         float64(burst),
			maxTokens:      float64(burst),
			refillRate:     refillRate,
			lastRefillTime: now,
			key:            key,
		}
		bucket.element = s.lru.PushFront(bucket)
		s.buckets[key] = bucket
	} else {
		s.lru.MoveToFront(bucket.element)
	}

	elapsed := now.Sub(bucket.lastRefillTime).Seconds()
	if elapsed > 0 && bucket.refillRate > 0 {
		newTokens := bucket.tokens + elapsed*bucket.refillRate
		if newTokens > bucket.maxTokens {
			newTokens = bucket.maxTokens
		}
		bucket.tokens = newTokens
		bucket.lastRefillTime = now
	}

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

func generateRateLimitKey(c *fiber.Ctx, pluginID string) string {
	ip := c.IP()
	if cfIP := c.Get("CF-Connecting-IP"); cfIP != "" {
		ip = cfIP
	} else if realIP := c.Get("X-Real-IP"); realIP != "" {
		ip = realIP
	}

	data := fmt.Sprintf("%s:%s:%s:%s", ip, c.Method(), pluginID, c.Path())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

func RegisterPluginRoutes(app *fiber.App) {
	app.All("/api/v1/plugins/:pluginId/*", handlePluginRoute)
}

func getRateLimitForRoute(pluginID, method, path string) *pb.RateLimitConfig {
	ps := GetStreamRegistry().Get(pluginID)
	if ps != nil && ps.Info != nil {
		for _, r := range ps.Info.Routes {
			if (r.Method == "*" || r.Method == method) && matchRoutePath(r.Path, path) {
				return r.RateLimit
			}
		}
	}

	plugin := GetRegistry().Get(pluginID)
	if plugin != nil && plugin.Info != nil {
		for _, r := range plugin.Info.Routes {
			if (r.Method == "*" || r.Method == method) && matchRoutePath(r.Path, path) {
				return r.RateLimit
			}
		}
	}

	return nil
}

func handlePluginRoute(c *fiber.Ctx) error {
	pluginID := c.Params("pluginId")

	path := c.Params("*")
	if path == "" {
		path = "/"
	} else {
		path = "/" + path
	}

	if pluginID == "ui" && path == "/manifests" {
		return handleUIManifests(c)
	}

	if path == "/ui/bundle.js" {
		return servePluginBundle(c, pluginID)
	}

	rateLimitCfg := getRateLimitForRoute(pluginID, c.Method(), path)
	if rateLimitCfg != nil {
		var rpm, burst int

		if rateLimitCfg.Preset != "" {
			if preset, ok := presetConfigs[rateLimitCfg.Preset]; ok {
				rpm = preset.RequestsPerMinute
				burst = preset.BurstLimit
			}
		} else if rateLimitCfg.RequestsPerMinute > 0 {
			rpm = int(rateLimitCfg.RequestsPerMinute)
			burst = int(rateLimitCfg.BurstLimit)
			if burst <= 0 {
				burst = rpm
			}
		}

		if rpm > 0 {
			key := generateRateLimitKey(c, pluginID)
			allowed, remaining, resetIn := checkPluginRateLimit(key, rpm, burst)

			c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", rpm))
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
		}
	}

	headers := make(map[string]string)
	c.Request().Header.VisitAll(func(k, v []byte) {
		headers[string(k)] = string(v)
	})

	query := make(map[string]string)
	c.Request().URI().QueryArgs().VisitAll(func(k, v []byte) {
		query[string(k)] = string(v)
	})

	userID := ""
	if user, ok := c.Locals("user").(*models.User); ok && user != nil {
		userID = user.ID.String()
	}

	req := &pb.HTTPRequest{
		Method:  c.Method(),
		Path:    path,
		Headers: headers,
		Query:   query,
		Body:    c.Body(),
		UserId:  userID,
	}

	if ps := GetStreamRegistry().Get(pluginID); ps != nil {
		resp, err := ps.SendHTTP(req)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "error": "plugin error: " + err.Error()})
		}
		if resp == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "route not found"})
		}
		for k, v := range resp.Headers {
			c.Set(k, v)
		}
		return c.Status(int(resp.Status)).Send(resp.Body)
	}

	plugin := GetRegistry().Get(pluginID)
	if plugin == nil || !plugin.Online {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "plugin not found"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := plugin.Client.OnHTTP(ctx, req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "error": "plugin error: " + err.Error()})
	}

	for k, v := range resp.Headers {
		c.Set(k, v)
	}

	return c.Status(int(resp.Status)).Send(resp.Body)
}


func servePluginBundle(c *fiber.Ctx, pluginID string) error {
	manifest := GetUIRegistry().Get(pluginID)
	if manifest == nil || !manifest.HasBundle {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "plugin bundle not found",
		})
	}

	if len(manifest.BundleData) > 0 {
		c.Set("Content-Type", "application/javascript")
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		return c.Send(manifest.BundleData)
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"success": false,
		"error":   "bundle not embedded",
	})
}

func handleUIManifests(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")
	manifests := GetUIRegistry().All()
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"plugins": manifests,
		},
	})
}
