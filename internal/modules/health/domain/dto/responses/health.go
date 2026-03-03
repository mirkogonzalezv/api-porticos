package responses

import "time"

// El status se maneja via HTTP Status code (200 = healthy y 503 = unhealthy)
type HealthResponse struct {
	Timestamp time.Time `json:"timestamp" example:"2024-01-15T10.30:00Z"`
	Service   string    `json:"service" example:"api-porticos"`
}

// Constructor
// NewHealthResponse()
func NewHealthResponse() *HealthResponse {
	return &HealthResponse{
		Timestamp: time.Now(),
		Service:   "api-porticos",
	}
}
