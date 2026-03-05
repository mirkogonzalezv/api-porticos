package entities

import (
	"strings"
	"time"

	domainErrors "rea/porticos/pkg/errors"
)

type PasoPortico struct {
	ID                  string    `json:"id"`
	OwnerSupabaseUserID string    `json:"ownerSupabaseUserId"`
	VehiculoID          string    `json:"vehiculoId"`
	PorticoID           string    `json:"porticoId"`
	FechaHoraPaso       time.Time `json:"fechaHoraPaso"`
	Latitud             *float64  `json:"latitud,omitempty"`
	Longitud            *float64  `json:"longitud,omitempty"`
	MontoCobrado        int       `json:"montoCobrado"`
	Moneda              string    `json:"moneda"`
	Fuente              string    `json:"fuente"`
	CreatedAt           time.Time `json:"createdAt"`
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
