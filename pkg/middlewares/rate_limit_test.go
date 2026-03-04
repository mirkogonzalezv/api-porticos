package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimiter_ReaderLimitExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := newRoleAwareRateLimiter(2, 60)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, "user-1")
		c.Set(ContextUserRoleKey, "reader")
		c.Next()
	})
	router.Use(limiter.Middleware())
	router.GET("/resource", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for i := 1; i <= 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/resource", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		router.ServeHTTP(w, req)

		if i < 3 && w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, w.Code)
		}
		if i == 3 && w.Code != http.StatusTooManyRequests {
			t.Fatalf("request %d: expected 429, got %d", i, w.Code)
		}
	}
}

func TestRateLimiter_PartnerHasHigherQuotaThanReader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := newRoleAwareRateLimiter(2, 60) // partner -> 4

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, "partner-1")
		c.Set(ContextUserRoleKey, "partner")
		c.Next()
	})
	router.Use(limiter.Middleware())
	router.GET("/resource", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for i := 1; i <= 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/resource", nil)
		req.RemoteAddr = "10.0.0.2:12345"
		router.ServeHTTP(w, req)

		if i <= 4 && w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, w.Code)
		}
		if i == 5 && w.Code != http.StatusTooManyRequests {
			t.Fatalf("request %d: expected 429, got %d", i, w.Code)
		}
	}
}

func TestRateLimiter_PublicUsesIPFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := newRoleAwareRateLimiter(20, 60) // public -> 10

	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/public", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for i := 1; i <= 11; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/public", nil)
		req.RemoteAddr = "10.0.0.3:12345"
		router.ServeHTTP(w, req)

		if i <= 10 && w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, w.Code)
		}
		if i == 11 && w.Code != http.StatusTooManyRequests {
			t.Fatalf("request %d: expected 429, got %d", i, w.Code)
		}
	}
}
