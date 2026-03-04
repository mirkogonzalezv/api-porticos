package middlewares

import (
	"strings"

	"rea/porticos/pkg/auth"

	"github.com/danielkov/gin-helmet/ginhelmet"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine, opts Options) {
	router.Use(gin.Logger())

	// Usa solo este recovery centralizado; evita duplicar con gin.Recovery()
	router.Use(ErrorHandlerMiddleware())

	router.Use(ginhelmet.Default())

	verifier := auth.NewSupabaseVerifier(
		opts.SupabaseJWKSURL,
		opts.SupabaseJWTIssuer,
		opts.SupabaseJWTAudience,
	)
	router.Use(AuthJWTMiddleware(verifier, opts.RoleResolver))

	rateLimiter := newRoleAwareRateLimiter(opts.RateLimit, opts.RateLimitWindowSec)
	router.Use(rateLimiter.Middleware())

	corsCfg := cors.Config{
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}
	if opts.AllowedOrigins != "" {
		corsCfg.AllowOrigins = splitCSV(opts.AllowedOrigins)
	}
	router.Use(cors.New(corsCfg))
}

func splitCSV(v string) []string {
	raw := strings.Split(v, ",")
	out := make([]string, 0, len(raw))
	for _, s := range raw {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
