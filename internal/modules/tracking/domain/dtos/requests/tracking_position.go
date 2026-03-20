package requests

import (
	"strings"
	"time"

	domainErrors "rea/porticos/pkg/errors"
)

type TrackingPositionRequest struct {
	VehiculoID string  `json:"vehiculoId"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	Speed      float64 `json:"speed"`
	Heading    float64 `json:"heading"`
	Timestamp  string  `json:"timestamp"`
}

func (r *TrackingPositionRequest) Validate() (time.Time, error) {
	if strings.TrimSpace(r.VehiculoID) == "" {
		return time.Time{}, domainErrors.NewValidationError("TRACKING_VEHICULO_REQUIRED", "vehiculoId es obligatorio")
	}
	if r.Lat < -90 || r.Lat > 90 {
		return time.Time{}, domainErrors.NewValidationError("TRACKING_LAT_INVALID", "lat inválida")
	}
	if r.Lng < -180 || r.Lng > 180 {
		return time.Time{}, domainErrors.NewValidationError("TRACKING_LNG_INVALID", "lng inválida")
	}
	if r.Speed < 0 {
		return time.Time{}, domainErrors.NewValidationError("TRACKING_SPEED_INVALID", "speed inválida")
	}
	if r.Heading < 0 || r.Heading >= 360 {
		return time.Time{}, domainErrors.NewValidationError("TRACKING_HEADING_INVALID", "heading inválido")
	}
	ts, err := time.Parse(time.RFC3339, strings.TrimSpace(r.Timestamp))
	if err != nil {
		return time.Time{}, domainErrors.NewValidationError("TRACKING_TIMESTAMP_INVALID", "timestamp debe usar RFC3339")
	}
	return ts, nil
}
