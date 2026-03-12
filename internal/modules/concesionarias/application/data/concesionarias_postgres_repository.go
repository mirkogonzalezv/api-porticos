package data

import (
	"context"
	"errors"
	"strings"

	"rea/porticos/internal/modules/concesionarias/domain/entities"
	"rea/porticos/internal/modules/concesionarias/domain/repository"
	domainErrors "rea/porticos/pkg/errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConcesionariasPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewConcesionariasPostgresRepository(pool *pgxpool.Pool) repository.ConcesionariaRepository {
	return &ConcesionariasPostgresRepository{pool: pool}
}

func (r *ConcesionariasPostgresRepository) Create(ctx context.Context, concesionaria *entities.Concesionaria) (*entities.Concesionaria, error) {
	if concesionaria == nil {
		return nil, domainErrors.NewValidationError("CONCESIONARIA_REQUIRED", "concesionaria es obligatoria")
	}
	if err := concesionaria.Validate(); err != nil {
		return nil, err
	}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO concesionarias (codigo, nombre, estado)
		VALUES ($1, $2, $3)
		RETURNING id::text
	`, concesionaria.Codigo, concesionaria.Nombre, concesionaria.Estado).Scan(&concesionaria.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domainErrors.NewConflictError("CONCESIONARIA_DUPLICADA", "ya existe una concesionaria con ese codigo o nombre")
		}
		return nil, domainErrors.NewInternalError("CONCESIONARIA_CREATE_ERROR", "error al crear concesionaria")
	}

	return r.GetByID(ctx, concesionaria.ID)
}

func (r *ConcesionariasPostgresRepository) List(ctx context.Context, filter repository.ListConcesionariasFilter) ([]entities.Concesionaria, error) {
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

	estado := strings.ToLower(strings.TrimSpace(filter.Estado))
	if estado != "" && estado != "active" && estado != "inactive" {
		return nil, domainErrors.NewValidationError("CONCESIONARIA_ESTADO_INVALID", "estado inválido")
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id::text, codigo, nombre, estado
		FROM concesionarias
		WHERE ($1 = '' OR estado = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, estado, limit, offset)
	if err != nil {
		return nil, domainErrors.NewInternalError("CONCESIONARIA_LIST_ERROR", "error al listar concesionarias")
	}
	defer rows.Close()

	out := make([]entities.Concesionaria, 0)
	for rows.Next() {
		var item entities.Concesionaria
		if err := rows.Scan(&item.ID, &item.Codigo, &item.Nombre, &item.Estado); err != nil {
			return nil, domainErrors.NewInternalError("CONCESIONARIA_LIST_SCAN_ERROR", "error al leer concesionarias")
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, domainErrors.NewInternalError("CONCESIONARIA_LIST_ROWS_ERROR", "error iterando concesionarias")
	}

	return out, nil
}

func (r *ConcesionariasPostgresRepository) GetByID(ctx context.Context, id string) (*entities.Concesionaria, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, domainErrors.NewValidationError("CONCESIONARIA_ID_REQUIRED", "id es obligatorio")
	}

	var item entities.Concesionaria
	err := r.pool.QueryRow(ctx, `
		SELECT id::text, codigo, nombre, estado
		FROM concesionarias
		WHERE id = $1
		LIMIT 1
	`, id).Scan(&item.ID, &item.Codigo, &item.Nombre, &item.Estado)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domainErrors.NewNotFoundError("CONCESIONARIA_NOT_FOUND", "concesionaria no encontrada")
		}
		return nil, domainErrors.NewInternalError("CONCESIONARIA_GET_ERROR", "error al obtener concesionaria")
	}

	return &item, nil
}

func (r *ConcesionariasPostgresRepository) Update(ctx context.Context, concesionaria *entities.Concesionaria) (*entities.Concesionaria, error) {
	if concesionaria == nil {
		return nil, domainErrors.NewValidationError("CONCESIONARIA_REQUIRED", "concesionaria es obligatoria")
	}
	if strings.TrimSpace(concesionaria.ID) == "" {
		return nil, domainErrors.NewValidationError("CONCESIONARIA_ID_REQUIRED", "id es obligatorio")
	}
	if err := concesionaria.Validate(); err != nil {
		return nil, err
	}

	tag, err := r.pool.Exec(ctx, `
		UPDATE concesionarias
		SET codigo = $2, nombre = $3, estado = $4, updated_at = NOW()
		WHERE id = $1
	`, concesionaria.ID, concesionaria.Codigo, concesionaria.Nombre, concesionaria.Estado)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domainErrors.NewConflictError("CONCESIONARIA_DUPLICADA", "ya existe una concesionaria con ese codigo o nombre")
		}
		return nil, domainErrors.NewInternalError("CONCESIONARIA_UPDATE_ERROR", "error al actualizar concesionaria")
	}
	if tag.RowsAffected() == 0 {
		return nil, domainErrors.NewNotFoundError("CONCESIONARIA_NOT_FOUND", "concesionaria no encontrada")
	}

	return r.GetByID(ctx, concesionaria.ID)
}

func (r *ConcesionariasPostgresRepository) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return domainErrors.NewValidationError("CONCESIONARIA_ID_REQUIRED", "id es obligatorio")
	}

	tag, err := r.pool.Exec(ctx, `DELETE FROM concesionarias WHERE id = $1`, id)
	if err != nil {
		if isForeignKeyViolation(err) {
			return domainErrors.NewConflictError("CONCESIONARIA_IN_USE", "existen pórticos asociados a esta concesionaria")
		}
		return domainErrors.NewInternalError("CONCESIONARIA_DELETE_ERROR", "error al eliminar concesionaria")
	}
	if tag.RowsAffected() == 0 {
		return domainErrors.NewNotFoundError("CONCESIONARIA_NOT_FOUND", "concesionaria no encontrada")
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

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23503"
}
