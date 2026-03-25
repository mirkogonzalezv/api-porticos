package entities

import (
	"strings"
	"time"

	domainErrors "rea/porticos/pkg/errors"
)

type PasoPortico struct {
	ID                  string     `json:"id"`
	OwnerSupabaseUserID string     `json:"ownerSupabaseUserId"`
	VehiculoID          string     `json:"vehiculoId"`
	VehiculoPatente     string     `json:"vehiculoPatente,omitempty"`
	PorticoID           string     `json:"porticoId"`
	PorticoCodigo       string     `json:"porticoCodigo,omitempty"`
	ConcesionariaNombre string     `json:"concesionariaNombre,omitempty"`
	FechaHoraPaso       time.Time  `json:"fechaHoraPaso"`
	EntryTimestamp      *time.Time `json:"entryTimestamp,omitempty"`
	ExitTimestamp       *time.Time `json:"exitTimestamp,omitempty"`
	DireccionPaso       string     `json:"direccionPaso,omitempty"`
	Latitud             *float64   `json:"latitud,omitempty"`
	Longitud            *float64   `json:"longitud,omitempty"`
	Heading             *float64   `json:"heading,omitempty"`
	Speed               *float64   `json:"speed,omitempty"`
	MontoCobrado        int        `json:"montoCobrado"`
	Moneda              string     `json:"moneda"`
	Fuente              string     `json:"fuente"`
	SourcePosition      any        `json:"sourcePosition,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
}

type ResumenPeriodo struct {
	Periodo    time.Time `json:"periodo"`
	TotalPasos int64     `json:"totalPasos"`
	TotalMonto int64     `json:"totalMonto"`
	Moneda     string    `json:"moneda"`
}

func (p *PasoPortico) ValidateForCreate() error {
	if strings.TrimSpace(p.OwnerSupabaseUserID) == "" {
		return domainErrors.NewValidationError("PASO_OWNER_REQUIRED", "usuario no autenticado")
	}
	if strings.TrimSpace(p.VehiculoID) == "" {
		return domainErrors.NewValidationError("PASO_VEHICULO_REQUIRED", "vehiculoId es obligatorio")
	}
	if strings.TrimSpace(p.PorticoID) == "" {
		return domainErrors.NewValidationError("PASO_PORTICO_REQUIRED", "porticoId es obligatorio")
	}
	if p.FechaHoraPaso.IsZero() {
		return domainErrors.NewValidationError("PASO_FECHA_REQUIRED", "fechaHoraPaso es obligatoria")
	}
	if p.MontoCobrado < 0 {
		return domainErrors.NewValidationError("PASO_MONTO_INVALID", "montoCobrado no puede ser negativo")
	}
	if p.EntryTimestamp != nil && p.ExitTimestamp != nil && p.ExitTimestamp.Before(*p.EntryTimestamp) {
		return domainErrors.NewValidationError("PASO_TIMESTAMP_RANGE_INVALID", "exitTimestamp debe ser posterior a entryTimestamp")
	}
	if p.DireccionPaso != "" {
		d := strings.ToUpper(strings.TrimSpace(p.DireccionPaso))
		switch d {
		case "N", "S", "E", "W", "NE", "NW", "SE", "SW":
		default:
			return domainErrors.NewValidationError("PASO_DIRECCION_INVALID", "direccionPaso no permitida")
		}
		p.DireccionPaso = d
	}

	if p.Moneda == "" {
		p.Moneda = "CLP"
	}
	p.Moneda = strings.ToUpper(strings.TrimSpace(p.Moneda))
	if len(p.Moneda) != 3 {
		return domainErrors.NewValidationError("PASO_MONEDA_INVALID", "moneda debe tener 3 caracteres")
	}

	if p.Fuente == "" {
		p.Fuente = "mobile"
	}
	p.Fuente = strings.ToLower(strings.TrimSpace(p.Fuente))
	switch p.Fuente {
	case "mobile", "backend", "batch":
	default:
		return domainErrors.NewValidationError("PASO_FUENTE_INVALID", "fuente no permitida")
	}

	if p.Latitud != nil && (*p.Latitud < -90 || *p.Latitud > 90) {
		return domainErrors.NewValidationError("PASO_LATITUD_INVALID", "latitud fuera de rango")
	}
	if p.Longitud != nil && (*p.Longitud < -180 || *p.Longitud > 180) {
		return domainErrors.NewValidationError("PASO_LONGITUD_INVALID", "longitud fuera de rango")
	}

	return nil
}
