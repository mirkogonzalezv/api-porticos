package usecases

import (
	"context"

	"rea/porticos/internal/modules/kpis/domain/entities"
	"rea/porticos/internal/modules/kpis/domain/repository"
)

type KPIsUseCase struct {
	repo repository.KPIRepository
}

func NewKPIsUseCase(repo repository.KPIRepository) *KPIsUseCase {
	return &KPIsUseCase{repo: repo}
}

func (uc *KPIsUseCase) GetBasicKPIs(ctx context.Context) (*entities.BasicKPIs, error) {
	return uc.repo.GetBasicKPIs(ctx)
}
