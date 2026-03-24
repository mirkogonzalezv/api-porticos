package requests

import (
	"strings"
	"time"

	domainErrors "rea/porticos/pkg/errors"
)

type GeoBatchRequest struct {
	VehiculoID string        `json:"vehiculoId"`
	DeviceID   string        `json:"deviceId"`
	Positions  []GeoPosition `json:"positions"`
}

type GeoPosition struct {
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Speed     float64 `json:"speed"`
	Heading   float64 `json:"heading"`
	Timestamp string  `json:"timestamp"`
}

func (r *GeoBatchRequest) Validate() error {
	if strings.TrimSpace(r.VehiculoID) == "" {
		return domainErrors.NewValidationError("GEO_VEHICULO_REQUIRED", "vehiculoId es obligatorio")
	}
	if len(r.Positions) == 0 {
		return domainErrors.NewValidationError("GEO_POSITIONS_REQUIRED", "positions es obligatorio")
	}
	if len(r.Positions) > 500 {
		return domainErrors.NewValidationError("GEO_POSITIONS_LIMIT", "positions excede el máximo permitido (500)")
	}
	for i := range r.Positions {
		p := r.Positions[i]
		if p.Lat < -90 || p.Lat > 90 {
			return domainErrors.NewValidationError("GEO_LAT_INVALID", "lat fuera de rango")
		}
		if p.Lng < -180 || p.Lng > 180 {
			return domainErrors.NewValidationError("GEO_LNG_INVALID", "lng fuera de rango")
		}
		if p.Speed < 0 {
			return domainErrors.NewValidationError("GEO_SPEED_INVALID", "speed inválida")
		}
		if p.Heading < 0 || p.Heading >= 360 {
			return domainErrors.NewValidationError("GEO_HEADING_INVALID", "heading inválido")
		}
		if _, err := time.Parse(time.RFC3339, strings.TrimSpace(p.Timestamp)); err != nil {
			return domainErrors.NewValidationError("GEO_TIMESTAMP_INVALID", "timestamp debe usar RFC3339")
		}
	}
	return nil
}
