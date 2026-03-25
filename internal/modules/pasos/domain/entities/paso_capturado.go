package entities

import (
	"strings"
	"time"

	domainErrors "rea/porticos/pkg/errors"
)

type PasoCapturado struct {
	ID                  string     `json:"id"`
	OwnerSupabaseUserID string     `json:"ownerSupabaseUserId"`
	VehiculoID          string     `json:"vehiculoId"`
	VehiculoPatente     string     `json:"vehiculoPatente,omitempty"`
	PorticoID           string     `json:"porticoId"`
	PorticoCodigo       string     `json:"porticoCodigo,omitempty"`
	ConcesionariaNombre string     `json:"concesionariaNombre,omitempty"`
	ViaID               string     `json:"viaId,omitempty"`
	FechaHoraInicio     time.Time  `json:"fechaHoraInicio"`
	FechaHoraFin        time.Time  `json:"fechaHoraFin"`
	EntryTimestamp      *time.Time `json:"entryTimestamp,omitempty"`
	ExitTimestamp       *time.Time `json:"exitTimestamp,omitempty"`
	EntryHit            bool       `json:"entryHit"`
	ExitHit             bool       `json:"exitHit"`
	HeadingAvg          *float64   `json:"headingAvg,omitempty"`
	SpeedAvg            *float64   `json:"speedAvg,omitempty"`
	DireccionPaso       string     `json:"direccionPaso,omitempty"`
	Status              string     `json:"status"`
	SourcePosition      any        `json:"sourcePosition,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
}

func (p *PasoCapturado) ValidateForCreate() error {
	if strings.TrimSpace(p.OwnerSupabaseUserID) == "" {
		return domainErrors.NewValidationError("CAPTURA_OWNER_REQUIRED", "usuario no autenticado")
	}
	if strings.TrimSpace(p.VehiculoID) == "" {
		return domainErrors.NewValidationError("CAPTURA_VEHICULO_REQUIRED", "vehiculoId es obligatorio")
	}
	if strings.TrimSpace(p.PorticoID) == "" {
		return domainErrors.NewValidationError("CAPTURA_PORTICO_REQUIRED", "porticoId es obligatorio")
	}
	if p.FechaHoraInicio.IsZero() || p.FechaHoraFin.IsZero() {
		return domainErrors.NewValidationError("CAPTURA_FECHA_REQUIRED", "fechaHoraInicio/Fin son obligatorias")
	}
	if p.FechaHoraFin.Before(p.FechaHoraInicio) {
		return domainErrors.NewValidationError("CAPTURA_FECHA_RANGE_INVALID", "fechaHoraFin debe ser posterior a fechaHoraInicio")
	}
	if p.Status == "" {
		p.Status = "CAPTURED"
	}
	p.Status = strings.ToUpper(strings.TrimSpace(p.Status))
	switch p.Status {
	case "CAPTURED", "CONFIRMED", "DISCARDED":
	default:
		return domainErrors.NewValidationError("CAPTURA_STATUS_INVALID", "status no permitido")
	}
	if p.DireccionPaso != "" {
		d := strings.ToUpper(strings.TrimSpace(p.DireccionPaso))
		switch d {
		case "N", "S", "E", "W", "NE", "NW", "SE", "SW":
		default:
			return domainErrors.NewValidationError("CAPTURA_DIRECCION_INVALID", "direccionPaso no permitida")
		}
		p.DireccionPaso = d
	}
	return nil
}
