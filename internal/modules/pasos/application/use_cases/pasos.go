package usecases

import (
	"context"
	"strings"
	"time"

	"rea/porticos/internal/modules/pasos/domain/dtos/requests"
	"rea/porticos/internal/modules/pasos/domain/entities"
	pasosRepository "rea/porticos/internal/modules/pasos/domain/repository"
	porticosRepository "rea/porticos/internal/modules/porticos/domain/repository"
	vehiculosRepository "rea/porticos/internal/modules/vehiculos/domain/repository"
	domainErrors "rea/porticos/pkg/errors"
)

type PasosUseCase struct {
	pasosRepo     pasosRepository.PasoPorticoRepository
	vehiculosRepo vehiculosRepository.VehiculoRepository
	porticosRepo  porticosRepository.PorticoRepository
}

func NewPasosUseCase(
	pasosRepo pasosRepository.PasoPorticoRepository,
	vehiculosRepo vehiculosRepository.VehiculoRepository,
	porticosRepo porticosRepository.PorticoRepository,
) *PasosUseCase {
	return &PasosUseCase{
		pasosRepo:     pasosRepo,
		vehiculosRepo: vehiculosRepo,
		porticosRepo:  porticosRepo,
	}
}

func (uc *PasosUseCase) Create(ctx context.Context, paso *entities.PasoPortico) (*entities.PasoPortico, error) {
	if paso == nil {
		return nil, domainErrors.NewValidationError("PASO_REQUIRED", "paso es obligatorio")
	}
	if err := paso.ValidateForCreate(); err != nil {
		return nil, err
	}

	if _, err := uc.vehiculosRepo.GetByID(ctx, paso.OwnerSupabaseUserID, paso.VehiculoID); err != nil {
		return nil, err
	}
	if _, err := uc.porticosRepo.GetByID(ctx, paso.PorticoID); err != nil {
		return nil, err
	}

	return uc.pasosRepo.Create(ctx, paso)
}

func (uc *PasosUseCase) CreateBatch(ctx context.Context, ownerID string, items []requests.CreatePasoBatchItem) ([]entities.PasoPortico, error) {
	return nil, domainErrors.NewValidationError("PASO_BATCH_DISABLED", "pasos/batch no está habilitado; usa geo/batch")
}

func (uc *PasosUseCase) GetByID(ctx context.Context, ownerID, id string) (*entities.PasoPortico, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(id) == "" {
		return nil, domainErrors.NewValidationError("PASO_REQUIRED_FIELDS", "usuario e id son obligatorios")
	}
	return uc.pasosRepo.GetByID(ctx, ownerID, id)
}

func (uc *PasosUseCase) ListByOwnerRange(
	ctx context.Context,
	ownerID string,
	from, to time.Time,
	vehiculoID, porticoID string,
	limit, offset int,
) ([]entities.PasoPortico, error) {
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return nil, domainErrors.NewValidationError("PASO_OWNER_REQUIRED", "usuario no autenticado")
	}
	if from.IsZero() || to.IsZero() {
		return nil, domainErrors.NewValidationError("PASO_RANGE_REQUIRED", "from y to son obligatorios")
	}
	if to.Before(from) {
		return nil, domainErrors.NewValidationError("PASO_RANGE_INVALID", "to no puede ser menor que from")
	}

	filter := pasosRepository.ListPasosFilter{
		From:       from,
		To:         to,
		VehiculoID: strings.TrimSpace(vehiculoID),
		PorticoID:  strings.TrimSpace(porticoID),
		Limit:      limit,
		Offset:     offset,
	}
	return uc.pasosRepo.ListByOwnerRange(ctx, ownerID, filter)
}

func (uc *PasosUseCase) ListAllRange(
	ctx context.Context,
	from, to time.Time,
	vehiculoID, porticoID string,
	limit, offset int,
) ([]entities.PasoPortico, error) {
	if from.IsZero() || to.IsZero() {
		return nil, domainErrors.NewValidationError("PASO_RANGE_REQUIRED", "from y to son obligatorios")
	}
	if to.Before(from) {
		return nil, domainErrors.NewValidationError("PASO_RANGE_INVALID", "to no puede ser menor que from")
	}

	filter := pasosRepository.ListPasosFilter{
		From:       from,
		To:         to,
		VehiculoID: strings.TrimSpace(vehiculoID),
		PorticoID:  strings.TrimSpace(porticoID),
		Limit:      limit,
		Offset:     offset,
	}
	return uc.pasosRepo.ListAllRange(ctx, filter)
}

func (uc *PasosUseCase) ListCapturadosByOwnerRange(
	ctx context.Context,
	ownerID string,
	from, to time.Time,
	vehiculoID, porticoID string,
	limit, offset int,
) ([]entities.PasoCapturado, error) {
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return nil, domainErrors.NewValidationError("CAPTURA_OWNER_REQUIRED", "usuario no autenticado")
	}
	if from.IsZero() || to.IsZero() {
		return nil, domainErrors.NewValidationError("CAPTURA_RANGE_REQUIRED", "from y to son obligatorios")
	}
	if to.Before(from) {
		return nil, domainErrors.NewValidationError("CAPTURA_RANGE_INVALID", "to no puede ser menor que from")
	}

	filter := pasosRepository.ListPasosFilter{
		From:       from,
		To:         to,
		VehiculoID: strings.TrimSpace(vehiculoID),
		PorticoID:  strings.TrimSpace(porticoID),
		Limit:      limit,
		Offset:     offset,
	}
	return uc.pasosRepo.ListCapturadosByOwnerRange(ctx, ownerID, filter)
}

func (uc *PasosUseCase) ListCapturadosAllRange(
	ctx context.Context,
	from, to time.Time,
	vehiculoID, porticoID string,
	limit, offset int,
) ([]entities.PasoCapturado, error) {
	if from.IsZero() || to.IsZero() {
		return nil, domainErrors.NewValidationError("CAPTURA_RANGE_REQUIRED", "from y to son obligatorios")
	}
	if to.Before(from) {
		return nil, domainErrors.NewValidationError("CAPTURA_RANGE_INVALID", "to no puede ser menor que from")
	}

	filter := pasosRepository.ListPasosFilter{
		From:       from,
		To:         to,
		VehiculoID: strings.TrimSpace(vehiculoID),
		PorticoID:  strings.TrimSpace(porticoID),
		Limit:      limit,
		Offset:     offset,
	}
	return uc.pasosRepo.ListCapturadosAllRange(ctx, filter)
}

func (uc *PasosUseCase) SummaryByOwnerRange(
	ctx context.Context,
	ownerID string,
	from, to time.Time,
	vehiculoID, porticoID, groupBy string,
) ([]entities.ResumenPeriodo, error) {
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return nil, domainErrors.NewValidationError("PASO_OWNER_REQUIRED", "usuario no autenticado")
	}
	if from.IsZero() || to.IsZero() {
		return nil, domainErrors.NewValidationError("PASO_RANGE_REQUIRED", "from y to son obligatorios")
	}
	if to.Before(from) {
		return nil, domainErrors.NewValidationError("PASO_RANGE_INVALID", "to no puede ser menor que from")
	}

	groupBy = strings.ToLower(strings.TrimSpace(groupBy))
	switch groupBy {
	case "day", "week", "month":
	default:
		return nil, domainErrors.NewValidationError("PASO_GROUPBY_INVALID", "groupBy debe ser day, week o month")
	}

	filter := pasosRepository.SummaryPasosFilter{
		From:       from,
		To:         to,
		VehiculoID: strings.TrimSpace(vehiculoID),
		PorticoID:  strings.TrimSpace(porticoID),
		GroupBy:    groupBy,
	}
	return uc.pasosRepo.SummaryByOwnerRange(ctx, ownerID, filter)
}
