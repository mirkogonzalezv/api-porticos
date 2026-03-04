package usecases

import (
	"context"
	"strings"

	"rea/porticos/internal/modules/porticos/domain/entities"
	"rea/porticos/internal/modules/porticos/domain/repository"
	domainErrors "rea/porticos/pkg/errors"
)

type PorticosUseCase struct {
	repo repository.PorticoRepository
}

func NewPorticosUseCase(repo repository.PorticoRepository) *PorticosUseCase {
	return &PorticosUseCase{
		repo: repo,
	}
}

func (uc *PorticosUseCase) Create(ctx context.Context, portico *entities.Portico) (*entities.Portico, error) {
	if portico == nil {
		return nil, domainErrors.NewValidationError("PORTICO_REQUIRED", "portico es obligatorio")
	}

	if err := portico.Validate(); err != nil {
		return nil, err
	}

	return uc.repo.Create(ctx, portico)
}

func (uc *PorticosUseCase) List(ctx context.Context, limit, offset int) ([]entities.Portico, error) {
	filter := repository.ListPorticosFilter{
		Limit:  limit,
		Offset: offset,
	}

	return uc.repo.List(ctx, filter)
}

func (uc *PorticosUseCase) GetByID(ctx context.Context, id string) (*entities.Portico, error) {
	if strings.TrimSpace(id) == "" {
		return nil, domainErrors.NewValidationError("PORTICO_ID_REQUIRED", "id es obligatorio")
	}

	return uc.repo.GetByID(ctx, id)
}

func (uc *PorticosUseCase) GetByCodigo(ctx context.Context, codigo string) (*entities.Portico, error) {
	if strings.TrimSpace(codigo) == "" {
		return nil, domainErrors.NewValidationError("PORTICO_CODIGO_REQUIRED", "codigo es obligatorio")
	}

	return uc.repo.GetByCodigo(ctx, codigo)
}

func (uc *PorticosUseCase) Update(ctx context.Context, id string, portico *entities.Portico) (*entities.Portico, error) {
	if strings.TrimSpace(id) == "" {
		return nil, domainErrors.NewValidationError("PORTICO_ID_REQUIRED", "id es obligatorio")
	}
	if portico == nil {
		return nil, domainErrors.NewValidationError("PORTICO_REQUIRED", "portico es obligatorio")
	}

	portico.ID = id

	if err := portico.Validate(); err != nil {
		return nil, err
	}

	return uc.repo.Update(ctx, portico)
}

func (uc *PorticosUseCase) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return domainErrors.NewValidationError("PORTICO_ID_REQUIRED", "id es obligatorio")
	}

	return uc.repo.Delete(ctx, id)
}
