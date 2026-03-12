package entities

import (
	domainErrors "rea/porticos/pkg/errors"
	"strings"
)

type Portico struct {
	ID                    string   `json:"id"`
	Codigo                string   `json:"codigo"`
	Nombre                string   `json:"nombre"`
	ConcesionariaID       string   `json:"concesionariaId,omitempty"`
	Concesionaria         string   `json:"concesionaria,omitempty"`
	Latitude              float64  `json:"latitude"`
	Longitude             float64  `json:"longitude"`
	Bearing               *float64 `json:"bearing,omitempty"`
	DetectionRadiusMeters *float64 `json:"detectionRadiusMeters,omitempty"`
	Tarifas               []Tarifa `json:"tarifas,omitempty"`
}

func (p *Portico) Validate() error {
	if strings.TrimSpace(p.Codigo) == "" {
		return domainErrors.NewValidationError("PORTICO_CODIGO_REQUIRED", "codigo es obligatorio")
	}

	if strings.TrimSpace(p.Nombre) == "" {
		return domainErrors.NewValidationError("PORTICO_NOMBRE_REQUIRED", "nombre es obligatorio")
	}

	p.ConcesionariaID = strings.TrimSpace(p.ConcesionariaID)
	if p.ConcesionariaID == "" {
		return domainErrors.NewValidationError("PORTICO_CONCESIONARIA_REQUIRED", "concesionariaId es obligatorio")
	}

	if p.Latitude < -90 || p.Latitude > 90 {
		return domainErrors.NewValidationError("PORTICO_LATITUDE_INVALID", "latitude debe estar entre -90 y 90")
	}

	if p.Longitude < -180 || p.Longitude > 180 {
		return domainErrors.NewValidationError("PORTICO_LONGITUDE_INVALID", "longitude debe estar entre -180 y 180")
	}

	if p.Bearing != nil && (*p.Bearing < 0 || *p.Bearing > 360) {
		return domainErrors.NewValidationError("PORTICO_BEARING_INVALID", "bearing debe estar entre 0 y 360")
	}
	if p.DetectionRadiusMeters != nil && *p.DetectionRadiusMeters <= 0 {
		return domainErrors.NewValidationError("PORTICO_RADIUS_INVALID", "detectionRadiusMeters debe ser mayor que 0")
	}

	for i := range p.Tarifas {
		if err := p.Tarifas[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}
