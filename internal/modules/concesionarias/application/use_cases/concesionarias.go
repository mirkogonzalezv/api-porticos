package usecases

import (
	"context"
	"strings"

	"rea/porticos/internal/modules/concesionarias/domain/entities"
	"rea/porticos/internal/modules/concesionarias/domain/repository"
	domainErrors "rea/porticos/pkg/errors"
)

type ConcesionariasUseCase struct {
	repo repository.ConcesionariaRepository
}

func NewConcesionariasUseCase(repo repository.ConcesionariaRepository) *ConcesionariasUseCase {
	return &ConcesionariasUseCase{repo: repo}
}

func (uc *ConcesionariasUseCase) Create(ctx context.Context, concesionaria *entities.Concesionaria) (*entities.Concesionaria, error) {
	if concesionaria == nil {
		return nil, domainErrors.NewValidationError("CONCESIONARIA_REQUIRED", "concesionaria es obligatoria")
	}
	if err := concesionaria.Validate(); err != nil {
		return nil, err
	}
	return uc.repo.Create(ctx, concesionaria)
}

func (uc *ConcesionariasUseCase) List(ctx context.Context, limit, offset int, estado string) ([]entities.Concesionaria, error) {
	filter := repository.ListConcesionariasFilter{
		Limit:  limit,
		Offset: offset,
		Estado: strings.TrimSpace(estado),
	}
	return uc.repo.List(ctx, filter)
}

func (uc *ConcesionariasUseCase) GetByID(ctx context.Context, id string) (*entities.Concesionaria, error) {
	if strings.TrimSpace(id) == "" {
		return nil, domainErrors.NewValidationError("CONCESIONARIA_ID_REQUIRED", "id es obligatorio")
	}
	return uc.repo.GetByID(ctx, id)
}

func (uc *ConcesionariasUseCase) Update(ctx context.Context, concesionaria *entities.Concesionaria) (*entities.Concesionaria, error) {
	if concesionaria == nil {
		return nil, domainErrors.NewValidationError("CONCESIONARIA_REQUIRED", "concesionaria es obligatoria")
	}
	if strings.TrimSpace(concesionaria.ID) == "" {
		return nil, domainErrors.NewValidationError("CONCESIONARIA_ID_REQUIRED", "id es obligatorio")
	}
	if err := concesionaria.Validate(); err != nil {
		return nil, err
	}
	return uc.repo.Update(ctx, concesionaria)
}

func (uc *ConcesionariasUseCase) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return domainErrors.NewValidationError("CONCESIONARIA_ID_REQUIRED", "id es obligatorio")
	}
	return uc.repo.Delete(ctx, id)
}
