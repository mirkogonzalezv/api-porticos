package middlewares

import (
	domainErrors "rea/porticos/pkg/errors"
	httpMapper "rea/porticos/pkg/http"
	"rea/porticos/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandlerMiddleware maneja errores globalmente
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		var err error

		switch e := recovered.(type) {
		case *domainErrors.DomainError:
			err = e
		case error:
			err = e
		default:
			err = domainErrors.NewInternalError("PANIC_001", "Unexpected panic occurred")
		}

		// Mapear error a HTTP
		statusCode, errorResponse := httpMapper.MapErrorToHttp(err)

		// Log según severidad
		if statusCode >= 500 {
			logger.L().Error("Server error",
				zap.String("error", err.Error()),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.Int("status", statusCode),
			)
		} else {
			logger.L().Warn("Client error",
				zap.String("error", err.Error()),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.Int("status", statusCode),
			)
		}

		c.JSON(statusCode, errorResponse)
	})
}

// ErrorLoggerMiddleware registra errores 5xx incluso si no hubo panic.
func ErrorLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		if status < 500 {
			return
		}

		errMsg := ""
		if len(c.Errors) > 0 {
			errMsg = c.Errors.String()
		} else {
			errMsg = "no error details"
		}

		reqID, _ := c.Get("request_id")
		userID, _ := c.Get(ContextUserIDKey)

		logger.L().Error("Request failed",
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("error", errMsg),
			zap.Any("request_id", reqID),
			zap.Any("user_id", userID),
			zap.Duration("latency_ms", time.Since(start)),
		)
	}
}
