package middlewares

type Options struct {
	Environment         string
	AllowedOrigins      string
	SupabaseJWKSURL     string
	SupabaseJWTIssuer   string
	SupabaseJWTAudience string
}
