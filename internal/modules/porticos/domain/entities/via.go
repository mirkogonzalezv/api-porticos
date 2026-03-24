package entities

import (
	"strings"

	domainErrors "rea/porticos/pkg/errors"
)

type Via struct {
	ID             string  `json:"id"`
	WayName        string  `json:"wayName"`
	DirectionDeg   float64 `json:"direction"`
	CenterLineWKT  string  `json:"centerLineWkt,omitempty"`
	EntryLineWKT   string  `json:"entryLineWkt,omitempty"`
	ExitLineWKT    string  `json:"exitLineWkt,omitempty"`
	EntryDistanceM float64 `json:"entryDistanceMeters"`
	ExitDistanceM  float64 `json:"exitDistanceMeters"`
	AutoCalculate  bool    `json:"autoCalculate"`
	IsActive       bool    `json:"isActive"`
}

func (v *Via) Validate() error {
	if strings.TrimSpace(v.WayName) == "" {
		return domainErrors.NewValidationError("PORTICO_VIA_NAME_REQUIRED", "wayName es obligatorio")
	}
	if v.DirectionDeg < 0 || v.DirectionDeg > 360 {
		return domainErrors.NewValidationError("PORTICO_VIA_DIR_INVALID", "direction debe estar entre 0 y 360")
	}
	if strings.TrimSpace(v.CenterLineWKT) == "" {
		return domainErrors.NewValidationError("PORTICO_VIA_CENTER_REQUIRED", "centerLine es obligatorio")
	}
	return nil
}
