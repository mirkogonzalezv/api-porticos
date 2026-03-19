package data

import (
	"context"
	"time"

	"rea/porticos/internal/modules/kpis/domain/entities"
	"rea/porticos/internal/modules/kpis/domain/repository"
	domainErrors "rea/porticos/pkg/errors"
	"rea/porticos/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type KPIsPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewKPIsPostgresRepository(pool *pgxpool.Pool) repository.KPIRepository {
	return &KPIsPostgresRepository{pool: pool}
}

func (r *KPIsPostgresRepository) GetBasicKPIs(ctx context.Context) (*entities.BasicKPIs, error) {
	kpis := &entities.BasicKPIs{}

	var err error
	for attempt := 0; attempt < 2; attempt++ {
		err = r.pool.QueryRow(ctx, `
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
		if err == nil {
			return kpis, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		logger.FromContext(ctx).Error("Error obteniendo KPIs básicos", zap.Error(err))
		return nil, domainErrors.NewInternalError("KPI_BASIC_QUERY_ERROR", "error al obtener KPIs básicos")
	}

	return kpis, nil
}
