package middlewares

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateEntry struct {
	count       int
	windowStart time.Time
}

type roleAwareRateLimiter struct {
	mu         sync.Mutex
	store      map[string]*rateEntry
	window     time.Duration
	baseLimit  int
	pruneEvery int
	reqCount   int
}

func newRoleAwareRateLimiter(baseLimit, windowSec int) *roleAwareRateLimiter {
	if baseLimit <= 0 {
		baseLimit = 100
	}
	if windowSec <= 0 {
		windowSec = 60
	}
	return &roleAwareRateLimiter{
		store:      make(map[string]*rateEntry),
		window:     time.Duration(windowSec) * time.Second,
		baseLimit:  baseLimit,
		pruneEvery: 200,
	}
}

func (rl *roleAwareRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := c.ClientIP()
		authenticated := false
		role := "public"

		if rawUserID, ok := c.Get(ContextUserIDKey); ok {
			if userID, okCast := rawUserID.(string); okCast && strings.TrimSpace(userID) != "" {
				identifier = "user:" + strings.TrimSpace(userID)
				authenticated = true
			}
		}
		if rawRole, ok := c.Get(ContextUserRoleKey); ok {
			if roleValue, okCast := rawRole.(string); okCast && strings.TrimSpace(roleValue) != "" {
				role = strings.ToLower(strings.TrimSpace(roleValue))
			}
		}
		if !authenticated {
			identifier = "ip:" + identifier
		}

		limit := rl.limitByRole(role, authenticated)
		now := time.Now()

		rl.mu.Lock()
		entry, ok := rl.store[identifier]
		if !ok {
			entry = &rateEntry{windowStart: now}
			rl.store[identifier] = entry
		}

		if now.Sub(entry.windowStart) >= rl.window {
			entry.count = 0
			entry.windowStart = now
		}

		if entry.count >= limit {
			resetAt := entry.windowStart.Add(rl.window)
			retryAfter := int(time.Until(resetAt).Seconds())
			if retryAfter < 0 {
				retryAfter = 0
			}
			remaining := 0

			rl.reqCount++
			rl.pruneLocked(now)
			rl.mu.Unlock()

			setRateHeaders(c, limit, remaining, resetAt.Unix(), retryAfter)

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"type":    "RATE_LIMIT",
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "demasiadas solicitudes, intenta nuevamente más tarde",
				},
				"timestamp": now.UTC().Format(time.RFC3339),
			})
			return
		}

		entry.count++
		remaining := limit - entry.count
		resetAt := entry.windowStart.Add(rl.window)

		rl.reqCount++
		rl.pruneLocked(now)
		rl.mu.Unlock()

		setRateHeaders(c, limit, remaining, resetAt.Unix(), 0)
		c.Next()
	}
}

func (rl *roleAwareRateLimiter) limitByRole(role string, authenticated bool) int {
	// Perfil por defecto:
	// - public: 50% del base
	// - reader: base
	// - partner: 2x base
	// - admin: 3x base
	if !authenticated {
		limit := rl.baseLimit / 2
		if limit < 10 {
			limit = 10
		}
		return limit
	}

	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin":
		return rl.baseLimit * 3
	case "partner":
		return rl.baseLimit * 2
	case "reader":
		return rl.baseLimit
	default:
		return rl.baseLimit
	}
}

func (rl *roleAwareRateLimiter) pruneLocked(now time.Time) {
	if rl.pruneEvery <= 0 || rl.reqCount%rl.pruneEvery != 0 {
		return
	}

	cutoff := now.Add(-2 * rl.window)
	for key, entry := range rl.store {
		if entry.windowStart.Before(cutoff) {
			delete(rl.store, key)
		}
	}
}

func setRateHeaders(c *gin.Context, limit, remaining int, resetUnix int64, retryAfter int) {
	c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(resetUnix, 10))
	if retryAfter > 0 {
		c.Header("Retry-After", strconv.Itoa(retryAfter))
	}
}
