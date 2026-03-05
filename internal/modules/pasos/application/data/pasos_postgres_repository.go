package data

import (
	"context"
	"fmt"
	"strings"

	"rea/porticos/internal/modules/pasos/domain/entities"
	"rea/porticos/internal/modules/pasos/domain/repository"
	domainErrors "rea/porticos/pkg/errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PasosPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPasosPostgresRepository(pool *pgxpool.Pool) repository.PasoPorticoRepository {
	return &PasosPostgresRepository{pool: pool}
}

func (r *PasosPostgresRepository) Create(ctx context.Context, paso *entities.PasoPortico) (*entities.PasoPortico, error) {
	if paso == nil {
		return nil, domainErrors.NewValidationError("PASO_REQUIRED", "paso es obligatorio")
	}
	if err := paso.ValidateForCreate(); err != nil {
		return nil, err
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO pasos_portico (
			owner_supabase_user_id,
			vehiculo_id,
			portico_id,
			fecha_hora_paso,
			latitud,
			longitud,
			monto_cobrado,
			moneda,
			fuente
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id::text
	`,
		paso.OwnerSupabaseUserID,
		paso.VehiculoID,
		paso.PorticoID,
		paso.FechaHoraPaso,
		paso.Latitud,
		paso.Longitud,
		paso.MontoCobrado,
		paso.Moneda,
		paso.Fuente,
	).Scan(&paso.ID)
	if err != nil {
		return nil, domainErrors.NewInternalError("PASO_CREATE_ERROR", "error al registrar paso de pórtico")
	}

	return r.GetByID(ctx, paso.OwnerSupabaseUserID, paso.ID)
}

func (r *PasosPostgresRepository) GetByID(ctx context.Context, ownerID, id string) (*entities.PasoPortico, error) {
	ownerID = strings.TrimSpace(ownerID)
	id = strings.TrimSpace(id)
	if ownerID == "" || id == "" {
		return nil, domainErrors.NewValidationError("PASO_REQUIRED_FIELDS", "usuario e id son obligatorios")
	}

	var out entities.PasoPortico
	err := r.pool.QueryRow(ctx, `
		SELECT
			id::text,
			owner_supabase_user_id::text,
			vehiculo_id::text,
			portico_id::text,
			fecha_hora_paso,
			latitud,
			longitud,
			monto_cobrado,
			moneda,
			fuente,
			created_at
		FROM pasos_portico
		WHERE owner_supabase_user_id = $1
		  AND id = $2
		LIMIT 1
	`, ownerID, id).Scan(
		&out.ID,
		&out.OwnerSupabaseUserID,
		&out.VehiculoID,
		&out.PorticoID,
		&out.FechaHoraPaso,
		&out.Latitud,
		&out.Longitud,
		&out.MontoCobrado,
		&out.Moneda,
		&out.Fuente,
		&out.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domainErrors.NewNotFoundError("PASO_NOT_FOUND", "paso no encontrado")
		}
		return nil, domainErrors.NewInternalError("PASO_GET_ERROR", "error al obtener paso")
	}

	return &out, nil
}

func (r *PasosPostgresRepository) ListByOwnerRange(
	ctx context.Context,
	ownerID string,
	filter repository.ListPasosFilter,
) ([]entities.PasoPortico, error) {
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return nil, domainErrors.NewValidationError("PASO_OWNER_REQUIRED", "usuario no autenticado")
	}

	limit := filter.Limit
	offset := filter.Offset
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	vehiculoID := strings.TrimSpace(filter.VehiculoID)
	porticoID := strings.TrimSpace(filter.PorticoID)

	rows, err := r.pool.Query(ctx, `
		SELECT
			id::text,
			owner_supabase_user_id::text,
			vehiculo_id::text,
			portico_id::text,
			fecha_hora_paso,
			latitud,
			longitud,
			monto_cobrado,
			moneda,
			fuente,
			created_at
		FROM pasos_portico
		WHERE owner_supabase_user_id = $1
		  AND fecha_hora_paso >= $2
		  AND fecha_hora_paso <= $3
		  AND ($4 = '' OR vehiculo_id::text = $4)
		  AND ($5 = '' OR portico_id::text = $5)
		ORDER BY fecha_hora_paso DESC
		LIMIT $6 OFFSET $7
	`, ownerID, filter.From, filter.To, vehiculoID, porticoID, limit, offset)
	if err != nil {
		return nil, domainErrors.NewInternalError("PASO_LIST_ERROR", "error al listar pasos")
	}
	defer rows.Close()

	out := make([]entities.PasoPortico, 0)
	for rows.Next() {
		var item entities.PasoPortico
		if err := rows.Scan(
			&item.ID,
			&item.OwnerSupabaseUserID,
			&item.VehiculoID,
			&item.PorticoID,
			&item.FechaHoraPaso,
			&item.Latitud,
			&item.Longitud,
			&item.MontoCobrado,
			&item.Moneda,
			&item.Fuente,
			&item.CreatedAt,
		); err != nil {
			return nil, domainErrors.NewInternalError("PASO_LIST_SCAN_ERROR", "error al leer pasos")
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, domainErrors.NewInternalError("PASO_LIST_ROWS_ERROR", "error iterando pasos")
	}

	return out, nil
}

func (r *PasosPostgresRepository) SummaryByOwnerRange(
	ctx context.Context,
	ownerID string,
	filter repository.SummaryPasosFilter,
) ([]entities.ResumenPeriodo, error) {
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return nil, domainErrors.NewValidationError("PASO_OWNER_REQUIRED", "usuario no autenticado")
	}

	groupBy := strings.ToLower(strings.TrimSpace(filter.GroupBy))
	switch groupBy {
	case "day", "week", "month":
	default:
		return nil, domainErrors.NewValidationError("PASO_GROUPBY_INVALID", "groupBy debe ser day, week o month")
	}

	query := fmt.Sprintf(`
		SELECT
			date_trunc('%s', fecha_hora_paso) AS periodo,
			COUNT(*) AS total_pasos,
			COALESCE(SUM(monto_cobrado), 0) AS total_monto,
			MIN(moneda) AS moneda
		FROM pasos_portico
		WHERE owner_supabase_user_id = $1
		  AND fecha_hora_paso >= $2
		  AND fecha_hora_paso <= $3
		  AND ($4 = '' OR vehiculo_id::text = $4)
		  AND ($5 = '' OR portico_id::text = $5)
		GROUP BY periodo
		ORDER BY periodo DESC
	`, groupBy)

	vehiculoID := strings.TrimSpace(filter.VehiculoID)
	porticoID := strings.TrimSpace(filter.PorticoID)

	rows, err := r.pool.Query(ctx, query, ownerID, filter.From, filter.To, vehiculoID, porticoID)
	if err != nil {
		return nil, domainErrors.NewInternalError("PASO_SUMMARY_ERROR", "error al obtener resumen de pasos")
	}
	defer rows.Close()

	out := make([]entities.ResumenPeriodo, 0)
	for rows.Next() {
		var item entities.ResumenPeriodo
		if err := rows.Scan(&item.Periodo, &item.TotalPasos, &item.TotalMonto, &item.Moneda); err != nil {
			return nil, domainErrors.NewInternalError("PASO_SUMMARY_SCAN_ERROR", "error al leer resumen de pasos")
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, domainErrors.NewInternalError("PASO_SUMMARY_ROWS_ERROR", "error iterando resumen de pasos")
	}

	return out, nil
}
