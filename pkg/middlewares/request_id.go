package middlewares

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"rea/porticos/pkg/logger"

	"github.com/gin-gonic/gin"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := strings.TrimSpace(c.GetHeader("X-Request-Id"))
		if reqID == "" {
			reqID = generateRequestID()
		}

		c.Set("request_id", reqID)
		c.Header("X-Request-Id", reqID)

		ctx := logger.WithRequestID(c.Request.Context(), reqID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func generateRequestID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
