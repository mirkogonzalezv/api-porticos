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
	BearingToleranceDeg   *int     `json:"bearingToleranceDeg,omitempty"`
	DetectionRadiusMeters *float64 `json:"detectionRadiusMeters,omitempty"`
	EntryRadiusMeters     *float64 `json:"entryRadiusMeters,omitempty"`
	ExitRadiusMeters      *float64 `json:"exitRadiusMeters,omitempty"`
	EntryLatitude         *float64 `json:"entryLatitude,omitempty"`
	EntryLongitude        *float64 `json:"entryLongitude,omitempty"`
	ExitLatitude          *float64 `json:"exitLatitude,omitempty"`
	ExitLongitude         *float64 `json:"exitLongitude,omitempty"`
	MaxCrossingSeconds    *int     `json:"maxCrossingSeconds,omitempty"`
	Tipo                  string   `json:"tipo"`
	Direccion             string   `json:"direccion"`
	VelocidadMaxima       int      `json:"velocidadMaxima"`
	ZonaDeteccionWKT      string   `json:"zonaDeteccionWkt,omitempty"`
	VehicleTypes          []string `json:"vehicleTypes,omitempty"`
	IsActive              bool     `json:"isActive"`
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
	if p.BearingToleranceDeg != nil {
		if *p.BearingToleranceDeg < 0 || *p.BearingToleranceDeg > 180 {
			return domainErrors.NewValidationError("PORTICO_BEARING_TOLERANCE_INVALID", "bearingToleranceDeg debe estar entre 0 y 180")
		}
	}
	if p.DetectionRadiusMeters != nil && *p.DetectionRadiusMeters <= 0 {
		return domainErrors.NewValidationError("PORTICO_RADIUS_INVALID", "detectionRadiusMeters debe ser mayor que 0")
	}
	if p.EntryRadiusMeters != nil && *p.EntryRadiusMeters <= 0 {
		return domainErrors.NewValidationError("PORTICO_ENTRY_RADIUS_INVALID", "entryRadiusMeters debe ser mayor que 0")
	}
	if p.ExitRadiusMeters != nil && *p.ExitRadiusMeters <= 0 {
		return domainErrors.NewValidationError("PORTICO_EXIT_RADIUS_INVALID", "exitRadiusMeters debe ser mayor que 0")
	}
	if (p.EntryLatitude != nil && p.EntryLongitude == nil) || (p.EntryLatitude == nil && p.EntryLongitude != nil) {
		return domainErrors.NewValidationError("PORTICO_ENTRY_COORDS_REQUIRED", "entryLatitude y entryLongitude deben estar juntas")
	}
	if p.EntryLatitude != nil && (*p.EntryLatitude < -90 || *p.EntryLatitude > 90) {
		return domainErrors.NewValidationError("PORTICO_ENTRY_LAT_INVALID", "entryLatitude debe estar entre -90 y 90")
	}
	if p.EntryLongitude != nil && (*p.EntryLongitude < -180 || *p.EntryLongitude > 180) {
		return domainErrors.NewValidationError("PORTICO_ENTRY_LNG_INVALID", "entryLongitude debe estar entre -180 y 180")
	}
	if (p.ExitLatitude != nil && p.ExitLongitude == nil) || (p.ExitLatitude == nil && p.ExitLongitude != nil) {
		return domainErrors.NewValidationError("PORTICO_EXIT_COORDS_REQUIRED", "exitLatitude y exitLongitude deben estar juntas")
	}
	if p.ExitLatitude != nil && (*p.ExitLatitude < -90 || *p.ExitLatitude > 90) {
		return domainErrors.NewValidationError("PORTICO_EXIT_LAT_INVALID", "exitLatitude debe estar entre -90 y 90")
	}
	if p.ExitLongitude != nil && (*p.ExitLongitude < -180 || *p.ExitLongitude > 180) {
		return domainErrors.NewValidationError("PORTICO_EXIT_LNG_INVALID", "exitLongitude debe estar entre -180 y 180")
	}
	if p.MaxCrossingSeconds != nil {
		if *p.MaxCrossingSeconds <= 0 || *p.MaxCrossingSeconds > 3600 {
			return domainErrors.NewValidationError("PORTICO_MAX_CROSSING_INVALID", "maxCrossingSeconds debe estar entre 1 y 3600")
		}
	}
	if strings.TrimSpace(p.Tipo) == "" {
		p.Tipo = "urbano"
	}
	p.Tipo = strings.ToLower(strings.TrimSpace(p.Tipo))
	switch p.Tipo {
	case "autopista", "peaje_manual", "peaje_automatico", "urbano":
	default:
		return domainErrors.NewValidationError("PORTICO_TIPO_INVALID", "tipo no permitido")
	}

	if strings.TrimSpace(p.Direccion) == "" {
		p.Direccion = "N"
	}
	p.Direccion = strings.ToUpper(strings.TrimSpace(p.Direccion))
	switch p.Direccion {
	case "N", "S", "E", "W", "NE", "NW", "SE", "SW":
	default:
		return domainErrors.NewValidationError("PORTICO_DIRECCION_INVALID", "direccion no permitida")
	}
	if p.VelocidadMaxima == 0 {
		p.VelocidadMaxima = 60
	}
	if p.VelocidadMaxima < 1 || p.VelocidadMaxima > 200 {
		return domainErrors.NewValidationError("PORTICO_VELOCIDAD_MAX_INVALID", "velocidadMaxima fuera de rango")
	}
	if strings.TrimSpace(p.ZonaDeteccionWKT) != "" {
		wkt := strings.ToUpper(strings.TrimSpace(p.ZonaDeteccionWKT))
		if !strings.HasPrefix(wkt, "POLYGON") {
			return domainErrors.NewValidationError("PORTICO_ZONA_WKT_INVALID", "zonaDeteccionWkt debe ser POLYGON WKT")
		}
	}

	for i := range p.Tarifas {
		if err := p.Tarifas[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}
