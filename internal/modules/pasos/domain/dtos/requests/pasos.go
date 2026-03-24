package requests

import (
	"strings"
	"time"

	"rea/porticos/internal/modules/pasos/domain/entities"
	domainErrors "rea/porticos/pkg/errors"
)

type CreatePasoPorticoRequest struct {
	VehiculoID    string   `json:"vehiculoId"`
	PorticoID     string   `json:"porticoId"`
	FechaHoraPaso string   `json:"fechaHoraPaso"`
	Latitud       *float64 `json:"latitud,omitempty"`
	Longitud      *float64 `json:"longitud,omitempty"`
	DireccionPaso string   `json:"direccionPaso,omitempty"`
	MontoCobrado  int      `json:"montoCobrado"`
	Moneda        string   `json:"moneda,omitempty"`
	Fuente        string   `json:"fuente,omitempty"`
}

func (r *CreatePasoPorticoRequest) ToEntity(ownerID string) (*entities.PasoPortico, error) {
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return nil, domainErrors.NewValidationError("PASO_OWNER_REQUIRED", "usuario no autenticado")
	}

	fechaHoraPaso, err := time.Parse(time.RFC3339, strings.TrimSpace(r.FechaHoraPaso))
	if err != nil {
		return nil, domainErrors.NewValidationError("PASO_FECHA_INVALID", "fechaHoraPaso debe usar formato RFC3339")
	}

	out := &entities.PasoPortico{
		OwnerSupabaseUserID: ownerID,
		VehiculoID:          strings.TrimSpace(r.VehiculoID),
		PorticoID:           strings.TrimSpace(r.PorticoID),
		FechaHoraPaso:       fechaHoraPaso,
		DireccionPaso:       strings.TrimSpace(r.DireccionPaso),
		Latitud:             r.Latitud,
		Longitud:            r.Longitud,
		MontoCobrado:        r.MontoCobrado,
		Moneda:              strings.TrimSpace(r.Moneda),
		Fuente:              strings.TrimSpace(r.Fuente),
	}
	if err := out.ValidateForCreate(); err != nil {
		return nil, err
	}

	return out, nil
}
