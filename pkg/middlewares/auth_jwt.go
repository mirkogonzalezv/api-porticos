package middlewares

import (
	"context"
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

type UserRoleResolver interface {
	ResolveRole(ctx context.Context, supabaseUserID string) (string, error)
}

func AuthJWTMiddleware(verifier *auth.SupabaseVerifier, roleResolver UserRoleResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		// CORS preflight siempre público.
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Health endpoint público
		if c.Request.Method == http.MethodGet && c.Request.URL.Path == "/api/v1/health" {
			c.Next()
			return
		}
		// Endpoints públicos de cuentas
		if c.Request.Method == http.MethodPost && (c.Request.URL.Path == "/api/v1/accounts/signup" ||
			c.Request.URL.Path == "/api/v1/accounts") {
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

		resolvedRole := strings.ToLower(strings.TrimSpace(claims.Role))
		if roleResolver != nil {
			roleFromDB, err := roleResolver.ResolveRole(c.Request.Context(), claims.Subject)
			if err != nil {
				respondAuthError(c, err)
				return
			}
			resolvedRole = strings.ToLower(strings.TrimSpace(roleFromDB))
		}
		if resolvedRole == "" {
			respondAuthError(c, domainErrors.NewForbiddenError("ROLE_REQUIRED", "perfil no autorizado"))
			return
		}

		c.Set(ContextUserIDKey, claims.Subject)
		c.Set(ContextUserRoleKey, resolvedRole)
		c.Set(ContextUserEmailKey, claims.Email)
		c.Next()
	}
}

func respondAuthError(c *gin.Context, err error) {
	status, payload := httpMapper.MapErrorToHttp(err)
	c.AbortWithStatusJSON(status, payload)
}
