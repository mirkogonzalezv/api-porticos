package requests

import (
	"strings"

	"rea/porticos/internal/modules/concesionarias/domain/entities"
	domainErrors "rea/porticos/pkg/errors"
)

type ConcesionariaUpsertRequest struct {
	Codigo string `json:"codigo,omitempty"`
	Nombre string `json:"nombre"`
	Estado string `json:"estado,omitempty"`
}

func (r *ConcesionariaUpsertRequest) ToEntity() (*entities.Concesionaria, error) {
	out := &entities.Concesionaria{
		Codigo: strings.TrimSpace(r.Codigo),
		Nombre: strings.TrimSpace(r.Nombre),
		Estado: strings.TrimSpace(r.Estado),
	}
	if err := out.Validate(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ConcesionariaUpsertRequest) ToEntityWithID(id string) (*entities.Concesionaria, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, domainErrors.NewValidationError("CONCESIONARIA_ID_REQUIRED", "id es obligatorio")
	}
	out, err := r.ToEntity()
	if err != nil {
		return nil, err
	}
	out.ID = id
	return out, nil
}
