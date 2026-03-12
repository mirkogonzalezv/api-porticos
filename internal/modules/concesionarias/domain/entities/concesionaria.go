package entities

import (
	"strings"

	domainErrors "rea/porticos/pkg/errors"
)

type Concesionaria struct {
	ID     string `json:"id"`
	Codigo string `json:"codigo,omitempty"`
	Nombre string `json:"nombre"`
	Estado string `json:"estado"`
}

func (c *Concesionaria) Validate() error {
	c.Codigo = strings.TrimSpace(c.Codigo)
	c.Nombre = strings.TrimSpace(c.Nombre)
	c.Estado = strings.ToLower(strings.TrimSpace(c.Estado))

	if c.Nombre == "" {
		return domainErrors.NewValidationError("CONCESIONARIA_NOMBRE_REQUIRED", "nombre es obligatorio")
	}
	if len(c.Nombre) > 120 {
		return domainErrors.NewValidationError("CONCESIONARIA_NOMBRE_INVALID", "nombre no puede superar 120 caracteres")
	}
	if c.Codigo != "" && len(c.Codigo) > 50 {
		return domainErrors.NewValidationError("CONCESIONARIA_CODIGO_INVALID", "codigo no puede superar 50 caracteres")
	}
	if c.Estado == "" {
		c.Estado = "active"
	}
	switch c.Estado {
	case "active", "inactive":
	default:
		return domainErrors.NewValidationError("CONCESIONARIA_ESTADO_INVALID", "estado inválido")
	}

	return nil
}
