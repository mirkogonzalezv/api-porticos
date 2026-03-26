package usecases

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"rea/porticos/internal/modules/geo/domain/dtos/requests"
	pasosEntities "rea/porticos/internal/modules/pasos/domain/entities"
	pasosRepo "rea/porticos/internal/modules/pasos/domain/repository"
	porticosEntities "rea/porticos/internal/modules/porticos/domain/entities"
	porticosRepo "rea/porticos/internal/modules/porticos/domain/repository"
	vehiculosRepo "rea/porticos/internal/modules/vehiculos/domain/repository"
	domainErrors "rea/porticos/pkg/errors"
)

type GeoBatchUseCase struct {
	vehiculos vehiculosRepo.VehiculoRepository
	porticos  porticosRepo.PorticoRepository
	pasos     pasosRepo.PasoPorticoRepository
}

type GeoBatchResult struct {
	Status     string   `json:"status"`
	Created    int      `json:"created"`
	Duplicates int      `json:"duplicates"`
	Skipped    int      `json:"skipped"`
	PasoIDs    []string `json:"pasoIds,omitempty"`
	PorticoIDs []string `json:"porticoIds,omitempty"`
	VehiculoID string   `json:"vehiculoId"`
	DeviceID   string   `json:"deviceId,omitempty"`
	Positions  int      `json:"positions"`
}

type parsedPosition struct {
	Lat       float64
	Lng       float64
	Speed     float64
	Heading   float64
	Timestamp time.Time
}

func NewGeoBatchUseCase(
	vehiculos vehiculosRepo.VehiculoRepository,
	porticos porticosRepo.PorticoRepository,
	pasos pasosRepo.PasoPorticoRepository,
) *GeoBatchUseCase {
	return &GeoBatchUseCase{vehiculos: vehiculos, porticos: porticos, pasos: pasos}
}

func (uc *GeoBatchUseCase) ProcessBatch(ctx context.Context, ownerID string, req *requests.GeoBatchRequest, idempotencyKey string) (*GeoBatchResult, error) {
	if strings.TrimSpace(ownerID) == "" {
		return nil, domainErrors.NewUnauthorizedError("AUTH_REQUIRED", "usuario no autenticado")
	}
	if req == nil {
		return nil, domainErrors.NewValidationError("GEO_BATCH_REQUIRED", "batch es obligatorio")
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	positions := req.PointsList()
	if len(positions) < 3 {
		return nil, domainErrors.NewValidationError("GEO_POSITIONS_MIN", "points debe tener al menos 3 puntos")
	}

	vehiculo, err := uc.vehiculos.GetByID(ctx, ownerID, req.VehiculoID)
	if err != nil {
		return nil, err
	}

	ok, err := uc.pasos.AcquireIdempotencyKey(ctx, ownerID, idempotencyKey, "geo_batch", 15*time.Minute)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &GeoBatchResult{
			Status:     "DUPLICATE",
			VehiculoID: req.VehiculoID,
			DeviceID:   strings.TrimSpace(req.DeviceID),
			Positions:  len(positions),
		}, nil
	}

	parsed, err := parsePositions(positions)
	if err != nil {
		return nil, err
	}

	lineWKT := buildLineString(parsed)
	porticos, err := uc.porticos.FindByTrajectory(ctx, lineWKT)
	if err != nil {
		return nil, err
	}

	maxSpeed := maxSpeed(parsed)
	from, to := timeRange(parsed)
	windowFrom := from.Add(-30 * time.Second)
	windowTo := to.Add(30 * time.Second)

	result := &GeoBatchResult{
		Status:     "PROCESSED",
		VehiculoID: req.VehiculoID,
		DeviceID:   strings.TrimSpace(req.DeviceID),
		Positions:  len(positions),
	}

	capturas := make([]*pasosEntities.PasoCapturado, 0)
	confirmados := make([]*pasosEntities.PasoPortico, 0)

	avgH := avgHeading(parsed)
	avgS := avgSpeed(parsed)
	direccion := headingToDirection(avgH)
	sourcePositions := req.PointsList()

	for i := range porticos {
		p := porticos[i]
		if !p.IsActive {
			result.Skipped++
			continue
		}
		if !vehicleTypeAllowed(p.VehicleTypes, vehiculo.TipoVehiculo) {
			result.Skipped++
			continue
		}
		if p.VelocidadMaxima > 0 && maxSpeed > float64(p.VelocidadMaxima) {
			result.Skipped++
			continue
		}

		crossings, err := uc.porticos.FindViaCrossingsByTrajectory(ctx, p.ID, lineWKT)
		if err != nil {
			return nil, err
		}
		if len(crossings) == 0 {
			captura := &pasosEntities.PasoCapturado{
				OwnerSupabaseUserID: ownerID,
				VehiculoID:          req.VehiculoID,
				PorticoID:           p.ID,
				FechaHoraInicio:     from,
				FechaHoraFin:        to,
				EntryTimestamp:      nil,
				ExitTimestamp:       nil,
				EntryHit:            false,
				ExitHit:             false,
				HeadingAvg:          avgH,
				SpeedAvg:            avgS,
				DireccionPaso:       direccion,
				Status:              "CAPTURED",
				SourcePosition:      sourcePositions,
			}
			if err := captura.ValidateForCreate(); err != nil {
				return nil, err
			}
			capturas = append(capturas, captura)
			result.Skipped++
			continue
		}

		for _, crossing := range crossings {
			status := "CAPTURED"
			entryTS := (*time.Time)(nil)
			exitTS := (*time.Time)(nil)
			if crossing.EntryHit {
				entryTS = &from
			}
			if crossing.ExitHit {
				exitTS = &to
			}
			monto, moneda := 0, "CLP"
			if crossing.EntryHit && crossing.ExitHit {
				status = "CONFIRMED"
				monto, moneda = resolveTarifa(p.Tarifas, vehiculo.TipoVehiculo, to)
			}

			captura := &pasosEntities.PasoCapturado{
				OwnerSupabaseUserID: ownerID,
				VehiculoID:          req.VehiculoID,
				PorticoID:           p.ID,
				ViaID:               crossing.ViaID,
				FechaHoraInicio:     from,
				FechaHoraFin:        to,
				EntryTimestamp:      entryTS,
				ExitTimestamp:       exitTS,
				EntryHit:            crossing.EntryHit,
				ExitHit:             crossing.ExitHit,
				HeadingAvg:          avgH,
				SpeedAvg:            avgS,
				DireccionPaso:       direccion,
				Status:              status,
				SourcePosition:      sourcePositions,
			}
			if err := captura.ValidateForCreate(); err != nil {
				return nil, err
			}
			capturas = append(capturas, captura)

			if status != "CONFIRMED" {
				result.Skipped++
				continue
			}

			dup, err := uc.hasDuplicate(ctx, ownerID, req.VehiculoID, p.ID, windowFrom, windowTo)
			if err != nil {
				return nil, err
			}
			if dup {
				result.Duplicates++
				continue
			}

			paso := &pasosEntities.PasoPortico{
				OwnerSupabaseUserID: ownerID,
				VehiculoID:          req.VehiculoID,
				PorticoID:           p.ID,
				FechaHoraPaso:       to,
				EntryTimestamp:      entryTS,
				ExitTimestamp:       exitTS,
				DireccionPaso:       p.Direccion,
				Latitud:             &parsed[len(parsed)-1].Lat,
				Longitud:            &parsed[len(parsed)-1].Lng,
				Heading:             &parsed[len(parsed)-1].Heading,
				Speed:               &parsed[len(parsed)-1].Speed,
				MontoCobrado:        monto,
				Moneda:              moneda,
				Fuente:              "batch",
			}
			if err := paso.ValidateForCreate(); err != nil {
				return nil, err
			}
			result.Created++
			confirmados = append(confirmados, paso)
		}
	}

	if err := uc.pasos.CreateCapturesBatch(ctx, capturas); err != nil {
		return nil, err
	}
	if err := uc.pasos.CreateConfirmadosBatch(ctx, confirmados); err != nil {
		return nil, err
	}

	return result, nil
}

func (uc *GeoBatchUseCase) hasDuplicate(ctx context.Context, ownerID, vehiculoID, porticoID string, from, to time.Time) (bool, error) {
	items, err := uc.pasos.ListByOwnerRange(ctx, ownerID, pasosRepo.ListPasosFilter{
		From:       from,
		To:         to,
		VehiculoID: vehiculoID,
		PorticoID:  porticoID,
		Limit:      1,
		Offset:     0,
	})
	if err != nil {
		return false, err
	}
	return len(items) > 0, nil
}

func parsePositions(items []requests.GeoPosition) ([]parsedPosition, error) {
	out := make([]parsedPosition, 0, len(items))
	for i := range items {
		p := items[i]
		ts, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(p.Timestamp))
		if err != nil {
			return nil, domainErrors.NewValidationError("GEO_TIMESTAMP_INVALID", "timestamp debe usar RFC3339")
		}
		out = append(out, parsedPosition{
			Lat:       p.Lat,
			Lng:       p.Lng,
			Speed:     p.Speed,
			Heading:   p.Heading,
			Timestamp: ts,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Timestamp.Before(out[j].Timestamp)
	})
	return out, nil
}

func buildLineString(points []parsedPosition) string {
	b := strings.Builder{}
	b.WriteString("LINESTRING(")
	for i := range points {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(formatCoord(points[i].Lng, points[i].Lat))
	}
	b.WriteString(")")
	return b.String()
}

func avgHeading(points []parsedPosition) *float64 {
	if len(points) == 0 {
		return nil
	}
	var sum float64
	for i := range points {
		sum += points[i].Heading
	}
	avg := sum / float64(len(points))
	return &avg
}

func avgSpeed(points []parsedPosition) *float64 {
	if len(points) == 0 {
		return nil
	}
	var sum float64
	for i := range points {
		sum += points[i].Speed
	}
	avg := sum / float64(len(points))
	return &avg
}

func headingToDirection(heading *float64) string {
	if heading == nil {
		return ""
	}
	h := *heading
	for h < 0 {
		h += 360
	}
	for h >= 360 {
		h -= 360
	}
	switch {
	case h >= 337.5 || h < 22.5:
		return "N"
	case h < 67.5:
		return "NE"
	case h < 112.5:
		return "E"
	case h < 157.5:
		return "SE"
	case h < 202.5:
		return "S"
	case h < 247.5:
		return "SW"
	case h < 292.5:
		return "W"
	default:
		return "NW"
	}
}

func formatCoord(lng, lat float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f %.6f", lng, lat), "0"), ".")
}

func maxSpeed(points []parsedPosition) float64 {
	max := 0.0
	for i := range points {
		if points[i].Speed > max {
			max = points[i].Speed
		}
	}
	return max
}

func timeRange(points []parsedPosition) (time.Time, time.Time) {
	if len(points) == 0 {
		return time.Time{}, time.Time{}
	}
	return points[0].Timestamp, points[len(points)-1].Timestamp
}

func vehicleTypeAllowed(allowed []string, tipoVehiculo string) bool {
	if len(allowed) == 0 {
		return true
	}
	needle := strings.ToLower(strings.TrimSpace(tipoVehiculo))
	for _, t := range allowed {
		if strings.ToLower(strings.TrimSpace(t)) == needle {
			return true
		}
	}
	return false
}

func resolveTarifa(tarifas []porticosEntities.Tarifa, tipoVehiculo string, ts time.Time) (int, string) {
	needle := strings.ToLower(strings.TrimSpace(tipoVehiculo))
	for _, t := range tarifas {
		if strings.ToLower(strings.TrimSpace(t.TipoVehiculo)) != needle {
			continue
		}
		for _, h := range t.Horarios {
			if withinHorario(ts, h.Inicio, h.Fin) {
				return h.Monto, strings.ToUpper(strings.TrimSpace(t.Moneda))
			}
		}
		if t.Moneda != "" {
			return 0, strings.ToUpper(strings.TrimSpace(t.Moneda))
		}
		return 0, "CLP"
	}
	return 0, "CLP"
}

func withinHorario(ts time.Time, inicio time.Time, fin time.Time) bool {
	value := time.Date(2000, 1, 1, ts.Hour(), ts.Minute(), ts.Second(), 0, time.UTC)
	start := time.Date(2000, 1, 1, inicio.Hour(), inicio.Minute(), inicio.Second(), 0, time.UTC)
	end := time.Date(2000, 1, 1, fin.Hour(), fin.Minute(), fin.Second(), 0, time.UTC)
	if end.Before(start) {
		return !value.Before(start) || !value.After(end)
	}
	return !value.Before(start) && !value.After(end)
}
