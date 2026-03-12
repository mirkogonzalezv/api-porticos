package repository

import (
	"context"

	"rea/porticos/internal/modules/kpis/domain/entities"
)

type KPIRepository interface {
	GetBasicKPIs(ctx context.Context) (*entities.BasicKPIs, error)
}
