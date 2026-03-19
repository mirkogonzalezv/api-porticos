package data

import (
	"context"
	"errors"
	"strings"

	"rea/porticos/internal/modules/vehiculos/domain/entities"
	"rea/porticos/internal/modules/vehiculos/domain/repository"
	domainErrors "rea/porticos/pkg/errors"
	"rea/porticos/pkg/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type VehiculosPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewVehiculosPostgresRepository(pool *pgxpool.Pool) repository.VehiculoRepository {
	return &VehiculosPostgresRepository{pool: pool}
}

func (r *VehiculosPostgresRepository) Create(ctx context.Context, vehiculo *entities.Vehiculo) (*entities.Vehiculo, error) {
	if vehiculo == nil {
		return nil, domainErrors.NewValidationError("VEHICULO_REQUIRED", "vehiculo es obligatorio")
	}
	if err := vehiculo.Validate(); err != nil {
		return nil, err
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO vehiculos (
			owner_supabase_user_id, patente, tipo_vehiculo, alias, activo
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text
	`, vehiculo.OwnerSupabaseUserID, vehiculo.Patente, vehiculo.TipoVehiculo, vehiculo.Alias, vehiculo.Activo).Scan(&vehiculo.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domainErrors.NewConflictError("VEHICULO_PATENTE_DUPLICADA", "ya existe un vehículo con esa patente")
		}
		logger.L().Error("Vehiculo create error", zap.Error(err))
		return nil, domainErrors.NewInternalError("VEHICULO_CREATE_ERROR", "error al crear vehículo")
	}
	return r.GetByID(ctx, vehiculo.OwnerSupabaseUserID, vehiculo.ID)
}

func (r *VehiculosPostgresRepository) ListByOwner(ctx context.Context, ownerID string, filter repository.ListVehiculosFilter) ([]entities.Vehiculo, error) {
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return nil, domainErrors.NewValidationError("VEHICULO_OWNER_REQUIRED", "usuario no autenticado")
	}
	limit := filter.Limit
	offset := filter.Offset
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id::text, owner_supabase_user_id::text, patente, tipo_vehiculo, alias, activo
		FROM vehiculos
		WHERE owner_supabase_user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, ownerID, limit, offset)
	if err != nil {
		logger.L().Error("Vehiculo list error", zap.Error(err))
		return nil, domainErrors.NewInternalError("VEHICULO_LIST_ERROR", "error al listar vehículos")
	}
	defer rows.Close()

	out := make([]entities.Vehiculo, 0)
	for rows.Next() {
		var v entities.Vehiculo
		if err := rows.Scan(&v.ID, &v.OwnerSupabaseUserID, &v.Patente, &v.TipoVehiculo, &v.Alias, &v.Activo); err != nil {
			logger.L().Error("Vehiculo list scan error", zap.Error(err))
			return nil, domainErrors.NewInternalError("VEHICULO_LIST_SCAN_ERROR", "error al leer vehículos")
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		logger.L().Error("Vehiculo list rows error", zap.Error(err))
		return nil, domainErrors.NewInternalError("VEHICULO_LIST_ROWS_ERROR", "error iterando vehículos")
	}
	return out, nil
}

func (r *VehiculosPostgresRepository) ListAll(ctx context.Context, filter repository.ListVehiculosFilter) ([]entities.Vehiculo, error) {
	limit := filter.Limit
	offset := filter.Offset
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id::text, owner_supabase_user_id::text, patente, tipo_vehiculo, alias, activo
		FROM vehiculos
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		logger.L().Error("Vehiculo list all error", zap.Error(err))
		return nil, domainErrors.NewInternalError("VEHICULO_LIST_ERROR", "error al listar vehículos")
	}
	defer rows.Close()

	out := make([]entities.Vehiculo, 0)
	for rows.Next() {
		var v entities.Vehiculo
		if err := rows.Scan(&v.ID, &v.OwnerSupabaseUserID, &v.Patente, &v.TipoVehiculo, &v.Alias, &v.Activo); err != nil {
			logger.L().Error("Vehiculo list all scan error", zap.Error(err))
			return nil, domainErrors.NewInternalError("VEHICULO_LIST_SCAN_ERROR", "error al leer vehículos")
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		logger.L().Error("Vehiculo list all rows error", zap.Error(err))
		return nil, domainErrors.NewInternalError("VEHICULO_LIST_ROWS_ERROR", "error iterando vehículos")
	}
	return out, nil
}

func (r *VehiculosPostgresRepository) GetByID(ctx context.Context, ownerID, id string) (*entities.Vehiculo, error) {
	ownerID = strings.TrimSpace(ownerID)
	id = strings.TrimSpace(id)
	if ownerID == "" || id == "" {
		return nil, domainErrors.NewValidationError("VEHICULO_REQUIRED_FIELDS", "usuario e id son obligatorios")
	}

	var v entities.Vehiculo
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, owner_supabase_user_id::text, patente, tipo_vehiculo, alias, activo
		FROM vehiculos
		WHERE owner_supabase_user_id = $1
		  AND id = $2
		LIMIT 1
	`, ownerID, id).Scan(&v.ID, &v.OwnerSupabaseUserID, &v.Patente, &v.TipoVehiculo, &v.Alias, &v.Activo)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domainErrors.NewNotFoundError("VEHICULO_NOT_FOUND", "vehículo no encontrado")
		}
		logger.L().Error("Vehiculo get error", zap.Error(err))
		return nil, domainErrors.NewInternalError("VEHICULO_GET_ERROR", "error al obtener vehículo")
	}
	return &v, nil
}

func (r *VehiculosPostgresRepository) GetByIDAny(ctx context.Context, id string) (*entities.Vehiculo, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, domainErrors.NewValidationError("VEHICULO_ID_REQUIRED", "id es obligatorio")
	}

	var v entities.Vehiculo
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, owner_supabase_user_id::text, patente, tipo_vehiculo, alias, activo
		FROM vehiculos
		WHERE id = $1
		LIMIT 1
	`, id).Scan(&v.ID, &v.OwnerSupabaseUserID, &v.Patente, &v.TipoVehiculo, &v.Alias, &v.Activo)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domainErrors.NewNotFoundError("VEHICULO_NOT_FOUND", "vehículo no encontrado")
		}
		logger.L().Error("Vehiculo get any error", zap.Error(err))
		return nil, domainErrors.NewInternalError("VEHICULO_GET_ERROR", "error al obtener vehículo")
	}
	return &v, nil
}

func (r *VehiculosPostgresRepository) Update(ctx context.Context, vehiculo *entities.Vehiculo) (*entities.Vehiculo, error) {
	if vehiculo == nil {
		return nil, domainErrors.NewValidationError("VEHICULO_REQUIRED", "vehiculo es obligatorio")
	}
	if err := vehiculo.Validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(vehiculo.ID) == "" {
		return nil, domainErrors.NewValidationError("VEHICULO_ID_REQUIRED", "id es obligatorio")
	}

	tag, err := r.pool.Exec(ctx, `
		UPDATE vehiculos
		SET
			patente = $3,
			tipo_vehiculo = $4,
			alias = $5,
			activo = $6,
			updated_at = NOW()
		WHERE owner_supabase_user_id = $1
		  AND id = $2
	`, vehiculo.OwnerSupabaseUserID, vehiculo.ID, vehiculo.Patente, vehiculo.TipoVehiculo, vehiculo.Alias, vehiculo.Activo)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domainErrors.NewConflictError("VEHICULO_PATENTE_DUPLICADA", "ya existe un vehículo con esa patente")
		}
		logger.L().Error("Vehiculo update error", zap.Error(err))
		return nil, domainErrors.NewInternalError("VEHICULO_UPDATE_ERROR", "error al actualizar vehículo")
	}
	if tag.RowsAffected() == 0 {
		return nil, domainErrors.NewNotFoundError("VEHICULO_NOT_FOUND", "vehículo no encontrado")
	}
	return r.GetByID(ctx, vehiculo.OwnerSupabaseUserID, vehiculo.ID)
}

func (r *VehiculosPostgresRepository) Delete(ctx context.Context, ownerID, id string) error {
	ownerID = strings.TrimSpace(ownerID)
	id = strings.TrimSpace(id)
	if ownerID == "" || id == "" {
		return domainErrors.NewValidationError("VEHICULO_REQUIRED_FIELDS", "usuario e id son obligatorios")
	}

	tag, err := r.pool.Exec(ctx, `DELETE FROM vehiculos WHERE owner_supabase_user_id = $1 AND id = $2`, ownerID, id)
	if err != nil {
		logger.L().Error("Vehiculo delete error", zap.Error(err))
		return domainErrors.NewInternalError("VEHICULO_DELETE_ERROR", "error al eliminar vehículo")
	}
	if tag.RowsAffected() == 0 {
		return domainErrors.NewNotFoundError("VEHICULO_NOT_FOUND", "vehículo no encontrado")
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505"
}
