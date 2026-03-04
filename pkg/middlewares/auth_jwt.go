package middlewares

import (
	"net/http"
	"strings"

	"rea/porticos/pkg/auth"
	domainErrors "rea/porticos/pkg/errors"
	httpMapper "rea/porticos/pkg/http"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey    = "auth.user_id"
	ContextUserRoleKey  = "auth.user_role"
	ContextUserEmailKey = "auth.user_email"
)

func AuthJWTMiddleware(verifier *auth.SupabaseVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Health endpoint público
		if c.Request.Method == http.MethodGet && c.Request.URL.Path == "/api/v1/health" {
			c.Next()
			return
		}

		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			respondAuthError(c, domainErrors.NewUnauthorizedError("AUTH_HEADER_REQUIRED", "Authorization header es obligatorio"))
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			respondAuthError(c, domainErrors.NewUnauthorizedError("AUTH_SCHEME_INVALID", "Authorization debe usar esquema Bearer"))
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
		if token == "" {
			respondAuthError(c, domainErrors.NewUnauthorizedError("AUTH_TOKEN_REQUIRED", "token JWT es obligatorio"))
			return
		}

		claims, err := verifier.Verify(c.Request.Context(), token)
		if err != nil {
			respondAuthError(c, domainErrors.NewUnauthorizedError("AUTH_TOKEN_INVALID", "token JWT inválido o expirado"))
			return
		}

		c.Set(ContextUserIDKey, claims.Subject)
		c.Set(ContextUserRoleKey, claims.Role)
		c.Set(ContextUserEmailKey, claims.Email)
		c.Next()
	}
}

func respondAuthError(c *gin.Context, err error) {
	status, payload := httpMapper.MapErrorToHttp(err)
	c.AbortWithStatusJSON(status, payload)
}
