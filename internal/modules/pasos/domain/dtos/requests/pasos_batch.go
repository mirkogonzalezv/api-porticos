package requests

import domainErrors "rea/porticos/pkg/errors"

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
	return domainErrors.NewValidationError("PASO_BATCH_DISABLED", "pasos/batch no está habilitado; usa geo/batch")
}
