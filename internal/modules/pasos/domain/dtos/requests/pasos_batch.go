package requests

import (
	"strings"

	domainErrors "rea/porticos/pkg/errors"
)

type CreatePasoBatchItem struct {
	VehiculoID    string   `json:"vehiculoId"`
	PorticoID     string   `json:"porticoId"`
	FechaHoraPaso string   `json:"fechaHoraPaso"`
	Latitud       *float64 `json:"latitud,omitempty"`
	Longitud      *float64 `json:"longitud,omitempty"`
	DireccionPaso string   `json:"direccionPaso,omitempty"`
	Fuente        string   `json:"fuente,omitempty"`
}

type CreatePasoBatchRequest struct {
	Items []CreatePasoBatchItem `json:"items"`
}

func (r *CreatePasoBatchRequest) Validate() error {
	if r == nil || len(r.Items) == 0 {
		return domainErrors.NewValidationError("PASO_BATCH_EMPTY", "items es obligatorio")
	}
	if len(r.Items) > 200 {
		return domainErrors.NewValidationError("PASO_BATCH_LIMIT", "items excede el máximo permitido (200)")
	}
	for i := range r.Items {
		if strings.TrimSpace(r.Items[i].VehiculoID) == "" || strings.TrimSpace(r.Items[i].PorticoID) == "" {
			return domainErrors.NewValidationError("PASO_BATCH_ITEM_INVALID", "vehiculoId y porticoId son obligatorios")
		}
		if strings.TrimSpace(r.Items[i].FechaHoraPaso) == "" {
			return domainErrors.NewValidationError("PASO_BATCH_ITEM_INVALID", "fechaHoraPaso es obligatoria")
		}
	}
	return nil
}
