package usecases

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"math"
	"time"

	"rea/porticos/internal/modules/pasos/domain/entities"
	pasosRepo "rea/porticos/internal/modules/pasos/domain/repository"
	porticoEntities "rea/porticos/internal/modules/porticos/domain/entities"
	porticosRepo "rea/porticos/internal/modules/porticos/domain/repository"
	"rea/porticos/internal/modules/tracking/domain/dtos/requests"
	trackingEntities "rea/porticos/internal/modules/tracking/domain/entities"
	trackingRepo "rea/porticos/internal/modules/tracking/application/data"
	vehiculosRepo "rea/porticos/internal/modules/vehiculos/domain/repository"
	domainErrors "rea/porticos/pkg/errors"
)

const (
	sessionTTL     = 3 * time.Minute
	lastPassTTL    = 30 * time.Second
	minInsideTime  = 10 * time.Second
	minInsideDistM = 15.0
	fastSpeedMps   = 20.0 / 3.6
)

type TrackingUseCase struct {
	porticos porticosRepo.PorticoRepository
	vehiculos vehiculosRepo.VehiculoRepository
	pasos pasosRepo.PasoPorticoRepository
	store trackingRepo.TrackingStore
}

func NewTrackingUseCase(
	porticos porticosRepo.PorticoRepository,
	vehiculos vehiculosRepo.VehiculoRepository,
	pasos pasosRepo.PasoPorticoRepository,
	store trackingRepo.TrackingStore,
) *TrackingUseCase {
	return &TrackingUseCase{porticos: porticos, vehiculos: vehiculos, pasos: pasos, store: store}
}

type TrackingResult struct {
	Status              string `json:"status"`
	PasoID              string `json:"pasoId,omitempty"`
	PorticoID           string `json:"porticoId,omitempty"`
	PorticoCodigo       string `json:"porticoCodigo,omitempty"`
	ConcesionariaNombre string `json:"concesionariaNombre,omitempty"`
	Timestamp           string `json:"timestamp,omitempty"`
}

func (uc *TrackingUseCase) ProcessPosition(ctx context.Context, ownerID string, req *requests.TrackingPositionRequest) (*TrackingResult, error) {
	if req == nil {
		return nil, domainErrors.NewValidationError("TRACKING_REQUIRED", "request es obligatorio")
	}
	if ownerID == "" {
		return nil, domainErrors.NewUnauthorizedError("AUTH_REQUIRED", "usuario no autenticado")
	}

	ts, err := req.Validate()
	if err != nil {
		return nil, err
	}

	if _, err := uc.vehiculos.GetByID(ctx, ownerID, req.VehiculoID); err != nil {
		return nil, err
	}

	candidates, err := uc.porticos.ListNearby(ctx, req.Lat, req.Lng, 1000)
	if err != nil {
		return nil, err
	}

	for i := range candidates {
		p := candidates[i]
		if !bearingOK(&p, req.Heading) {
			continue
		}

		entryRadius := radiusOrDefault(p.EntryRadiusMeters, p.DetectionRadiusMeters, 120)
		exitRadius := radiusOrDefault(p.ExitRadiusMeters, p.DetectionRadiusMeters, entryRadius)
		distance := haversineMeters(req.Lat, req.Lng, p.Latitude, p.Longitude)

		if distance > entryRadius {
			continue
		}

		session, err := uc.store.GetSession(ctx, req.VehiculoID, p.ID)
		if err != nil {
			return nil, err
		}
		if session == nil {
			session = &trackingEntities.TrackingSession{
				ID:         newID(),
				VehiculoID: req.VehiculoID,
				PorticoID:  p.ID,
			}
		}

		updateSession(session, req, ts, distance, entryRadius, exitRadius)
		if err := uc.store.SetSession(ctx, session, sessionTTL); err != nil {
			return nil, err
		}

		if session.State == trackingEntities.TrackingExited {
			if shouldValidate(session, ts, req.Speed) {
				lastPassAt, ok, err := uc.store.GetLastPass(ctx, req.VehiculoID, p.ID)
				if err != nil {
					return nil, err
				}
				if ok && ts.Sub(lastPassAt) < lastPassTTL {
					return &TrackingResult{Status: "NO_MATCH"}, nil
				}

				paso := &entities.PasoPortico{
					OwnerSupabaseUserID: ownerID,
					VehiculoID:          req.VehiculoID,
					PorticoID:           p.ID,
					FechaHoraPaso:       ts,
					Latitud:             &req.Lat,
					Longitud:            &req.Lng,
					Heading:             &req.Heading,
					Speed:               &req.Speed,
					MontoCobrado:        0,
					Moneda:              "CLP",
					Fuente:              "mobile",
					TrackingSessionID:   session.ID,
					SourcePosition: map[string]any{
						"lat": req.Lat,
						"lng": req.Lng,
					},
				}
				if err := paso.ValidateForCreate(); err != nil {
					return nil, err
				}

				created, err := uc.pasos.Create(ctx, paso)
				if err != nil {
					return nil, err
				}

				_ = uc.store.SetLastPass(ctx, req.VehiculoID, p.ID, ts, lastPassTTL)
				_ = uc.store.DeleteSession(ctx, req.VehiculoID, p.ID)

				return &TrackingResult{
					Status:              "VALIDATED",
					PasoID:              created.ID,
					PorticoID:           p.ID,
					PorticoCodigo:       p.Codigo,
					ConcesionariaNombre: p.Concesionaria,
					Timestamp:           ts.Format(time.RFC3339),
				}, nil
			}
		}
	}

	return &TrackingResult{Status: "NO_MATCH"}, nil
}

func updateSession(s *trackingEntities.TrackingSession, req *requests.TrackingPositionRequest, ts time.Time, dist, entryRadius, exitRadius float64) {
	s.LastSeenAt = ts
	s.LastLat = req.Lat
	s.LastLng = req.Lng
	s.LastHeading = req.Heading
	s.LastSpeed = req.Speed

	inside := dist <= entryRadius
	outside := dist > exitRadius

	if inside {
		s.InsideCount++
		s.OutsideCount = 0
		if s.State == "" && s.InsideCount >= 2 {
			s.State = trackingEntities.TrackingEntered
			s.EnteredAt = ts
			s.FirstLat = req.Lat
			s.FirstLng = req.Lng
		}
		if s.State == trackingEntities.TrackingEntered {
			s.State = trackingEntities.TrackingInside
		}
	}

	if outside {
		s.OutsideCount++
		s.InsideCount = 0
		if s.State == trackingEntities.TrackingInside && s.OutsideCount >= 2 {
			s.State = trackingEntities.TrackingExited
		}
	}
}

func shouldValidate(s *trackingEntities.TrackingSession, ts time.Time, speed float64) bool {
	if s.EnteredAt.IsZero() {
		return false
	}
	dt := ts.Sub(s.EnteredAt)
	if dt >= minInsideTime {
		return true
	}
	dist := haversineMeters(s.FirstLat, s.FirstLng, s.LastLat, s.LastLng)
	if dist >= minInsideDistM {
		return true
	}
	if speed >= fastSpeedMps && dt >= 2*time.Second {
		return true
	}
	return false
}

func bearingOK(p *porticoEntities.Portico, heading float64) bool {
	if p.Bearing == nil {
		return true
	}
	tol := 25.0
	if p.BearingToleranceDeg != nil {
		tol = float64(*p.BearingToleranceDeg)
	}
	diff := math.Abs(*p.Bearing - heading)
	if diff > 180 {
		diff = 360 - diff
	}
	return diff <= tol
}

func radiusOrDefault(primary *float64, fallback *float64, def float64) float64 {
	if primary != nil && *primary > 0 {
		return *primary
	}
	if fallback != nil && *fallback > 0 {
		return *fallback
	}
	return def
}

func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	la1 := lat1 * math.Pi / 180
	la2 := lat2 * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(la1)*math.Cos(la2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	hexStr := hex.EncodeToString(b)
	return hexStr[0:8] + "-" + hexStr[8:12] + "-" + hexStr[12:16] + "-" + hexStr[16:20] + "-" + hexStr[20:32]
}
