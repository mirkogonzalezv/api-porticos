package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireRoles_AllowsAuthorizedRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextUserRoleKey, "admin")
		c.Next()
	})
	router.GET("/protected", RequireRoles("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestRequireRoles_DeniesUnauthorizedRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextUserRoleKey, "reader")
		c.Next()
	})
	router.POST("/protected", RequireRoles("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON body, got error: %v", err)
	}
	errObj, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object in response body")
	}
	if errObj["code"] != "ROLE_FORBIDDEN" {
		t.Fatalf("expected ROLE_FORBIDDEN code, got %v", errObj["code"])
	}
}

func TestRequireRoles_DeniesMissingRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.DELETE("/protected", RequireRoles("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/protected", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}
