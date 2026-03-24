package requests

import (
	"strings"
	"time"

	"rea/porticos/internal/modules/porticos/domain/entities"
	domainErrors "rea/porticos/pkg/errors"
)

type PorticoUpsertRequest struct {
	Codigo                string          `json:"codigo"`
	Nombre                string          `json:"nombre"`
	ConcesionariaID       string          `json:"concesionariaId"`
	Latitude              float64         `json:"latitude"`
	Longitude             float64         `json:"longitude"`
	Bearing               *float64        `json:"bearing,omitempty"`
	DetectionRadiusMeters *float64        `json:"detectionRadiusMeters,omitempty"`
	EntryRadiusMeters     *float64        `json:"entryRadiusMeters,omitempty"`
	ExitRadiusMeters      *float64        `json:"exitRadiusMeters,omitempty"`
	EntryLatitude         *float64        `json:"entryLatitude,omitempty"`
	EntryLongitude        *float64        `json:"entryLongitude,omitempty"`
	ExitLatitude          *float64        `json:"exitLatitude,omitempty"`
	ExitLongitude         *float64        `json:"exitLongitude,omitempty"`
	MaxCrossingSeconds    *int            `json:"maxCrossingSeconds,omitempty"`
	Tipo                  string          `json:"tipo,omitempty"`
	Direccion             string          `json:"direccion,omitempty"`
	VelocidadMaxima       int             `json:"velocidadMaxima,omitempty"`
	ZonaDeteccionWKT      string          `json:"zonaDeteccionWkt,omitempty"`
	VehicleTypes          []string        `json:"vehicleTypes,omitempty"`
	IsActive              *bool           `json:"isActive,omitempty"`
	Vias                  []ViaRequest    `json:"vias,omitempty"`
	Tarifas               []TarifaRequest `json:"tarifas,omitempty"`
}

type TarifaRequest struct {
	TipoVehiculo string                 `json:"tipoVehiculo"`
	Moneda       string                 `json:"moneda"`
	Horarios     []TarifaHorarioRequest `json:"horarios"`
}

type TarifaHorarioRequest struct {
	Inicio string `json:"inicio"`
	Fin    string `json:"fin"`
	Monto  int    `json:"monto"`
}

func (r *PorticoUpsertRequest) ToEntity() (*entities.Portico, error) {
	out := &entities.Portico{
		Codigo:                strings.TrimSpace(r.Codigo),
		Nombre:                strings.TrimSpace(r.Nombre),
		ConcesionariaID:       strings.TrimSpace(r.ConcesionariaID),
		Latitude:              r.Latitude,
		Longitude:             r.Longitude,
		Bearing:               r.Bearing,
		DetectionRadiusMeters: r.DetectionRadiusMeters,
		EntryRadiusMeters:     r.EntryRadiusMeters,
		ExitRadiusMeters:      r.ExitRadiusMeters,
		EntryLatitude:         r.EntryLatitude,
		EntryLongitude:        r.EntryLongitude,
		ExitLatitude:          r.ExitLatitude,
		ExitLongitude:         r.ExitLongitude,
		MaxCrossingSeconds:    r.MaxCrossingSeconds,
		Tipo:                  strings.TrimSpace(r.Tipo),
		Direccion:             strings.TrimSpace(r.Direccion),
		VelocidadMaxima:       r.VelocidadMaxima,
		ZonaDeteccionWKT:      strings.TrimSpace(r.ZonaDeteccionWKT),
		VehicleTypes:          r.VehicleTypes,
		IsActive:              true,
		Vias:                  make([]entities.Via, 0, len(r.Vias)),
		Tarifas:               make([]entities.Tarifa, 0, len(r.Tarifas)),
	}
	if r.IsActive != nil {
		out.IsActive = *r.IsActive
	}

	for _, v := range r.Vias {
		via, err := v.ToEntity()
		if err != nil {
			return nil, err
		}
		out.Vias = append(out.Vias, via)
	}

	for _, tr := range r.Tarifas {
		tarifa := entities.Tarifa{
			TipoVehiculo: strings.TrimSpace(tr.TipoVehiculo),
			Moneda:       strings.TrimSpace(tr.Moneda),
			Horarios:     make([]entities.TarifaHorario, 0, len(tr.Horarios)),
		}

		for _, hr := range tr.Horarios {
			inicio, err := parseClock(hr.Inicio)
			if err != nil {
				return nil, domainErrors.NewValidationError("TARIFA_HORARIO_INICIO_INVALID", "inicio debe usar formato HH:MM o HH:MM:SS")
			}

			fin, err := parseClock(hr.Fin)
			if err != nil {
				return nil, domainErrors.NewValidationError("TARIFA_HORARIO_FIN_INVALID", "fin debe usar formato HH:MM o HH:MM:SS")
			}

			tarifa.Horarios = append(tarifa.Horarios, entities.TarifaHorario{
				Inicio: inicio,
				Fin:    fin,
				Monto:  hr.Monto,
			})
		}

		out.Tarifas = append(out.Tarifas, tarifa)
	}

	return out, nil
}

func parseClock(v string) (time.Time, error) {
	v = strings.TrimSpace(v)
	if t, err := time.Parse("15:04", v); err == nil {
		return t, nil
	}

	return time.Parse("15:04:05", v)
}
