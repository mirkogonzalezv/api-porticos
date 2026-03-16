package data

import (
	"context"

	"rea/porticos/internal/modules/kpis/domain/entities"
	"rea/porticos/internal/modules/kpis/domain/repository"
	domainErrors "rea/porticos/pkg/errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type KPIsPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewKPIsPostgresRepository(pool *pgxpool.Pool) repository.KPIRepository {
	return &KPIsPostgresRepository{pool: pool}
}

func (r *KPIsPostgresRepository) GetBasicKPIs(ctx context.Context) (*entities.BasicKPIs, error) {
	kpis := &entities.BasicKPIs{}

	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM porticos) AS total_porticos,
			(SELECT COUNT(*) FROM pasos_portico) AS total_registros_porticos,
			(SELECT COUNT(DISTINCT patente) FROM vehiculos) AS total_patentes,
			(SELECT COUNT(*) FROM concesionarias) AS total_concesionarias
	`).Scan(
		&kpis.TotalPorticos,
		&kpis.TotalRegistrosPorticos,
		&kpis.TotalPatentes,
		&kpis.TotalConcesionarias,
	)
	if err != nil {
		return nil, domainErrors.NewInternalError("KPI_BASIC_QUERY_ERROR", "error al obtener KPIs básicos")
	}

	return kpis, nil
}
