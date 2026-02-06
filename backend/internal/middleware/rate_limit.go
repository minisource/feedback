package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
	// Max requests per window
	Max int
	// Window duration
	Window time.Duration
	// Key function to identify requesters
	KeyFunc func(c *fiber.Ctx) string
}

// DefaultRateLimitKeyFunc returns the default key function for rate limiting
func DefaultRateLimitKeyFunc(c *fiber.Ctx) string {
	// Use combination of tenant and IP for rate limiting
	tenantID, _ := c.Locals("tenantId").(string)
	return tenantID + ":" + c.IP()
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(cfg RateLimitConfig) fiber.Handler {
	type visitor struct {
		count    int
		lastSeen time.Time
	}

	var (
		visitors = make(map[string]*visitor)
		mu       sync.Mutex
	)

	// Cleanup goroutine
	go func() {
		for {
			time.Sleep(cfg.Window)
			mu.Lock()
			for key, v := range visitors {
				if time.Since(v.lastSeen) > cfg.Window {
					delete(visitors, key)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *fiber.Ctx) error {
		key := cfg.KeyFunc(c)

		mu.Lock()
		v, exists := visitors[key]
		if !exists {
			visitors[key] = &visitor{count: 1, lastSeen: time.Now()}
			mu.Unlock()
			return c.Next()
		}

		// Reset if window expired
		if time.Since(v.lastSeen) > cfg.Window {
			v.count = 1
			v.lastSeen = time.Now()
			mu.Unlock()
			return c.Next()
		}

		// Increment count
		v.count++
		v.lastSeen = time.Now()

		if v.count > cfg.Max {
			mu.Unlock()
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "too_many_requests",
				"message": "Rate limit exceeded",
			})
		}

		mu.Unlock()
		return c.Next()
	}
}
