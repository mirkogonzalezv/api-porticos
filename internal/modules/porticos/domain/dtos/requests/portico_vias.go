package requests

import (
	"fmt"
	"strings"

	"rea/porticos/internal/modules/porticos/domain/entities"
	domainErrors "rea/porticos/pkg/errors"
)

type ViaRequest struct {
	WayName    string             `json:"way_name"`
	Direction  float64            `json:"direction"`
	CenterLine GeoLineString      `json:"center_line"`
	EntryLine  GeoLineString      `json:"entry_line"`
	ExitLine   GeoLineString      `json:"exit_line"`
	Settings   ViaSettingsRequest `json:"settings"`
	IsActive   *bool              `json:"is_active,omitempty"`
}

type ViaSettingsRequest struct {
	EntryDistance float64 `json:"entry_distance"`
	ExitDistance  float64 `json:"exit_distance"`
	AutoCalculate bool    `json:"auto_calculate"`
}

type GeoLineString struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

func (r *ViaRequest) ToEntity() (entities.Via, error) {
	if r == nil {
		return entities.Via{}, domainErrors.NewValidationError("PORTICO_VIA_REQUIRED", "via es obligatoria")
	}
	centerWKT, err := lineStringToWKT(&r.CenterLine)
	if err != nil {
		return entities.Via{}, err
	}
	entryWKT, err := lineStringToWKT(&r.EntryLine)
	if err != nil {
		return entities.Via{}, err
	}
	exitWKT, err := lineStringToWKT(&r.ExitLine)
	if err != nil {
		return entities.Via{}, err
	}

	out := entities.Via{
		WayName:        strings.TrimSpace(r.WayName),
		DirectionDeg:   r.Direction,
		CenterLineWKT:  centerWKT,
		EntryLineWKT:   entryWKT,
		ExitLineWKT:    exitWKT,
		EntryDistanceM: r.Settings.EntryDistance,
		ExitDistanceM:  r.Settings.ExitDistance,
		AutoCalculate:  r.Settings.AutoCalculate,
		IsActive:       true,
	}
	if r.IsActive != nil {
		out.IsActive = *r.IsActive
	}
	if err := out.Validate(); err != nil {
		return entities.Via{}, err
	}
	return out, nil
}

func lineStringToWKT(line *GeoLineString) (string, error) {
	if line == nil {
		return "", domainErrors.NewValidationError("PORTICO_VIA_LINE_REQUIRED", "lineString es obligatorio")
	}
	if strings.ToLower(strings.TrimSpace(line.Type)) != "linestring" {
		return "", domainErrors.NewValidationError("PORTICO_VIA_LINE_TYPE", "type debe ser LineString")
	}
	if len(line.Coordinates) < 2 {
		return "", domainErrors.NewValidationError("PORTICO_VIA_LINE_POINTS", "LineString requiere al menos 2 puntos")
	}
	b := strings.Builder{}
	b.WriteString("LINESTRING(")
	for i := range line.Coordinates {
		coord := line.Coordinates[i]
		if len(coord) < 2 {
			return "", domainErrors.NewValidationError("PORTICO_VIA_LINE_COORD", "coordenadas inválidas")
		}
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(formatCoord(coord[0], coord[1]))
	}
	b.WriteString(")")
	return b.String(), nil
}

func formatCoord(lng, lat float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f %.6f", lng, lat), "0"), ".")
}
