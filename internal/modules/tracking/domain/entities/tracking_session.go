package entities

import "time"

type TrackingState string

const (
	TrackingEntered   TrackingState = "ENTERED"
	TrackingInside    TrackingState = "INSIDE"
	TrackingExited    TrackingState = "EXITED"
	TrackingValidated TrackingState = "VALIDATED"
	TrackingDiscarded TrackingState = "DISCARDED"
)

type TrackingSession struct {
	ID           string        `json:"id"`
	VehiculoID   string        `json:"vehiculoId"`
	PorticoID    string        `json:"porticoId"`
	State        TrackingState `json:"state"`
	EnteredAt    time.Time     `json:"enteredAt"`
	ExitAt       time.Time     `json:"exitAt"`
	LastSeenAt   time.Time     `json:"lastSeenAt"`
	FirstLat     float64       `json:"firstLat"`
	FirstLng     float64       `json:"firstLng"`
	LastLat      float64       `json:"lastLat"`
	LastLng      float64       `json:"lastLng"`
	LastHeading  float64       `json:"lastHeading"`
	LastSpeed    float64       `json:"lastSpeed"`
	InsideCount  int           `json:"insideCount"`
	OutsideCount int           `json:"outsideCount"`
	ExitCount    int           `json:"exitCount"`
}
