package data

import (
	"context"
	"errors"
	"math"
	"strings"

	"rea/porticos/internal/modules/porticos/domain/entities"
	"rea/porticos/internal/modules/porticos/domain/repository"
	domainErrors "rea/porticos/pkg/errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPorticoRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPorticoRepository(pool *pgxpool.Pool) repository.PorticoRepository {
	return &PostgresPorticoRepository{
		pool: pool,
	}
}

func (r *PostgresPorticoRepository) Create(ctx context.Context, portico *entities.Portico) (*entities.Portico, error) {
	if portico == nil {
		return nil, domainErrors.NewValidationError("PORTICO_REQUIRED", "portico es obligatorio")
	}
	if err := portico.Validate(); err != nil {
		return nil, err
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, domainErrors.NewInternalError("PORTICO_TX_BEGIN_ERROR", "no se pudo iniciar transacción")
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO porticos (
			codigo, nombre, concesionaria_id, latitude, longitude, bearing, bearing_tolerance_deg, detection_radius_meters, entry_radius_meters, exit_radius_meters
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id::text
	`,
		portico.Codigo,
		portico.Nombre,
		portico.ConcesionariaID,
		portico.Latitude,
		portico.Longitude,
		portico.Bearing,
		portico.BearingToleranceDeg,
		portico.DetectionRadiusMeters,
		portico.EntryRadiusMeters,
		portico.ExitRadiusMeters,
	).Scan(&portico.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domainErrors.NewConflictError("PORTICO_CODIGO_DUPLICADO", "ya existe un pórtico con ese código")
		}
		return nil, domainErrors.NewInternalError("PORTICO_CREATE_ERROR", "error al crear pórtico")
	}

	if err := r.insertTarifasTx(ctx, tx, portico.ID, portico.Tarifas); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, domainErrors.NewInternalError("PORTICO_TX_COMMIT_ERROR", "no se pudo confirmar transacción")
	}

	return r.GetByID(ctx, portico.ID)
}

func (r *PostgresPorticoRepository) List(ctx context.Context, filter repository.ListPorticosFilter) ([]entities.Portico, error) {
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
		SELECT
			p.id::text,
			p.codigo,
			p.nombre,
			p.concesionaria_id::text,
			c.nombre AS concesionaria_nombre,
			p.latitude,
			p.longitude,
			p.bearing,
			p.bearing_tolerance_deg,
			p.detection_radius_meters,
			p.entry_radius_meters,
			p.exit_radius_meters
		FROM porticos p
		LEFT JOIN concesionarias c ON c.id = p.concesionaria_id
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, domainErrors.NewInternalError("PORTICO_LIST_ERROR", "error al listar pórticos")
	}
	defer rows.Close()

	porticos := make([]entities.Portico, 0)
	for rows.Next() {
		var p entities.Portico
		if err := rows.Scan(
			&p.ID,
			&p.Codigo,
			&p.Nombre,
			&p.ConcesionariaID,
			&p.Concesionaria,
			&p.Latitude,
			&p.Longitude,
			&p.Bearing,
			&p.BearingToleranceDeg,
			&p.DetectionRadiusMeters,
			&p.EntryRadiusMeters,
			&p.ExitRadiusMeters,
		); err != nil {
			return nil, domainErrors.NewInternalError("PORTICO_LIST_SCAN_ERROR", "error al leer pórticos")
		}

		tarifas, err := r.getTarifasByPorticoID(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		p.Tarifas = tarifas

		porticos = append(porticos, p)
	}

	if err := rows.Err(); err != nil {
		return nil, domainErrors.NewInternalError("PORTICO_LIST_ROWS_ERROR", "error iterando pórticos")
	}

	return porticos, nil
}

func (r *PostgresPorticoRepository) GetByID(ctx context.Context, id string) (*entities.Portico, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, domainErrors.NewValidationError("PORTICO_ID_REQUIRED", "id es obligatorio")
	}

	var p entities.Portico
	err := r.pool.QueryRow(ctx, `
		SELECT
			p.id::text,
			p.codigo,
			p.nombre,
			p.concesionaria_id::text,
			c.nombre AS concesionaria_nombre,
			p.latitude,
			p.longitude,
			p.bearing,
			p.bearing_tolerance_deg,
			p.detection_radius_meters,
			p.entry_radius_meters,
			p.exit_radius_meters
		FROM porticos p
		LEFT JOIN concesionarias c ON c.id = p.concesionaria_id
		WHERE p.id = $1
	`, id).Scan(
		&p.ID,
		&p.Codigo,
		&p.Nombre,
		&p.ConcesionariaID,
		&p.Concesionaria,
		&p.Latitude,
		&p.Longitude,
		&p.Bearing,
		&p.BearingToleranceDeg,
		&p.DetectionRadiusMeters,
		&p.EntryRadiusMeters,
		&p.ExitRadiusMeters,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domainErrors.NewNotFoundError("PORTICO_NOT_FOUND", "pórtico no encontrado")
		}
		return nil, domainErrors.NewInternalError("PORTICO_GET_ERROR", "error al obtener pórtico")
	}

	tarifas, err := r.getTarifasByPorticoID(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Tarifas = tarifas

	return &p, nil
}

func (r *PostgresPorticoRepository) ListNearby(ctx context.Context, lat, lng, maxDistanceMeters float64) ([]entities.Portico, error) {
	if maxDistanceMeters <= 0 {
		maxDistanceMeters = 500
	}
	dlat := maxDistanceMeters / 111320.0
	cosLat := math.Cos(lat * math.Pi / 180)
	dlng := maxDistanceMeters / (111320.0 * cosLat)

	rows, err := r.pool.Query(ctx, `
		SELECT
			p.id::text,
			p.codigo,
			p.nombre,
			p.concesionaria_id::text,
			c.nombre AS concesionaria_nombre,
			p.latitude,
			p.longitude,
			p.bearing,
			p.bearing_tolerance_deg,
			p.detection_radius_meters,
			p.entry_radius_meters,
			p.exit_radius_meters
		FROM porticos p
		LEFT JOIN concesionarias c ON c.id = p.concesionaria_id
		WHERE p.latitude BETWEEN $1 AND $2
		  AND p.longitude BETWEEN $3 AND $4
	`, lat-dlat, lat+dlat, lng-dlng, lng+dlng)
	if err != nil {
		return nil, domainErrors.NewInternalError("PORTICO_NEARBY_ERROR", "error al buscar pórticos cercanos")
	}
	defer rows.Close()

	out := make([]entities.Portico, 0)
	for rows.Next() {
		var p entities.Portico
		if err := rows.Scan(
			&p.ID,
			&p.Codigo,
			&p.Nombre,
			&p.ConcesionariaID,
			&p.Concesionaria,
			&p.Latitude,
			&p.Longitude,
			&p.Bearing,
			&p.BearingToleranceDeg,
			&p.DetectionRadiusMeters,
			&p.EntryRadiusMeters,
			&p.ExitRadiusMeters,
		); err != nil {
			return nil, domainErrors.NewInternalError("PORTICO_NEARBY_SCAN_ERROR", "error al leer pórticos cercanos")
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, domainErrors.NewInternalError("PORTICO_NEARBY_ROWS_ERROR", "error iterando pórticos cercanos")
	}

	return out, nil
}

func (r *PostgresPorticoRepository) GetByCodigo(ctx context.Context, codigo string) (*entities.Portico, error) {
	codigo = strings.TrimSpace(codigo)
	if codigo == "" {
		return nil, domainErrors.NewValidationError("PORTICO_CODIGO_REQUIRED", "codigo es obligatorio")
	}

	var p entities.Portico
	err := r.pool.QueryRow(ctx, `
		SELECT
			p.id::text,
			p.codigo,
			p.nombre,
			p.concesionaria_id::text,
			c.nombre AS concesionaria_nombre,
			p.latitude,
			p.longitude,
			p.bearing,
			p.bearing_tolerance_deg,
			p.detection_radius_meters,
			p.entry_radius_meters,
			p.exit_radius_meters
		FROM porticos p
		LEFT JOIN concesionarias c ON c.id = p.concesionaria_id
		WHERE p.codigo = $1
	`, codigo).Scan(
		&p.ID,
		&p.Codigo,
		&p.Nombre,
		&p.ConcesionariaID,
		&p.Concesionaria,
		&p.Latitude,
		&p.Longitude,
		&p.Bearing,
		&p.BearingToleranceDeg,
		&p.DetectionRadiusMeters,
		&p.EntryRadiusMeters,
		&p.ExitRadiusMeters,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domainErrors.NewNotFoundError("PORTICO_NOT_FOUND", "pórtico no encontrado")
		}
		return nil, domainErrors.NewInternalError("PORTICO_GET_BY_CODIGO_ERROR", "error al obtener pórtico por código")
	}

	tarifas, err := r.getTarifasByPorticoID(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Tarifas = tarifas

	return &p, nil
}

func (r *PostgresPorticoRepository) Update(ctx context.Context, portico *entities.Portico) (*entities.Portico, error) {
	if portico == nil {
		return nil, domainErrors.NewValidationError("PORTICO_REQUIRED", "portico es obligatorio")
	}
	if strings.TrimSpace(portico.ID) == "" {
		return nil, domainErrors.NewValidationError("PORTICO_ID_REQUIRED", "id es obligatorio")
	}
	if err := portico.Validate(); err != nil {
		return nil, err
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, domainErrors.NewInternalError("PORTICO_TX_BEGIN_ERROR", "no se pudo iniciar transacción")
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE porticos
		SET
			codigo = $2,
			nombre = $3,
			concesionaria_id = $4,
			latitude = $5,
			longitude = $6,
			bearing = $7,
			bearing_tolerance_deg = $8,
			detection_radius_meters = $9,
			entry_radius_meters = $10,
			exit_radius_meters = $11,
			updated_at = NOW()
		WHERE id = $1
	`,
		portico.ID,
		portico.Codigo,
		portico.Nombre,
		portico.ConcesionariaID,
		portico.Latitude,
		portico.Longitude,
		portico.Bearing,
		portico.BearingToleranceDeg,
		portico.DetectionRadiusMeters,
		portico.EntryRadiusMeters,
		portico.ExitRadiusMeters,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domainErrors.NewConflictError("PORTICO_CODIGO_DUPLICADO", "ya existe un pórtico con ese código")
		}
		return nil, domainErrors.NewInternalError("PORTICO_UPDATE_ERROR", "error al actualizar pórtico")
	}
	if tag.RowsAffected() == 0 {
		return nil, domainErrors.NewNotFoundError("PORTICO_NOT_FOUND", "pórtico no encontrado")
	}

	_, err = tx.Exec(ctx, `DELETE FROM tarifas_portico WHERE portico_id = $1`, portico.ID)
	if err != nil {
		return nil, domainErrors.NewInternalError("PORTICO_TARIFAS_DELETE_ERROR", "error al reemplazar tarifas")
	}

	if err := r.insertTarifasTx(ctx, tx, portico.ID, portico.Tarifas); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, domainErrors.NewInternalError("PORTICO_TX_COMMIT_ERROR", "no se pudo confirmar transacción")
	}

	return r.GetByID(ctx, portico.ID)
}

func (r *PostgresPorticoRepository) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return domainErrors.NewValidationError("PORTICO_ID_REQUIRED", "id es obligatorio")
	}

	tag, err := r.pool.Exec(ctx, `DELETE FROM porticos WHERE id = $1`, id)
	if err != nil {
		return domainErrors.NewInternalError("PORTICO_DELETE_ERROR", "error al eliminar pórtico")
	}
	if tag.RowsAffected() == 0 {
		return domainErrors.NewNotFoundError("PORTICO_NOT_FOUND", "pórtico no encontrado")
	}

	return nil
}

func (r *PostgresPorticoRepository) insertTarifasTx(
	ctx context.Context,
	tx pgx.Tx,
	porticoID string,
	tarifas []entities.Tarifa,
) error {
	for i := range tarifas {
		t := tarifas[i]
		if err := t.Validate(); err != nil {
			return err
		}

		var tarifaID string
		err := tx.QueryRow(ctx, `
			INSERT INTO tarifas_portico (
				portico_id, tipo_vehiculo, moneda
			) VALUES ($1, $2, $3)
			RETURNING id::text
		`,
			porticoID,
			t.TipoVehiculo,
			t.Moneda,
		).Scan(&tarifaID)
		if err != nil {
			if isUniqueViolation(err) {
				return domainErrors.NewConflictError("TARIFA_DUPLICADA", "ya existe tarifa para ese tipo de vehículo")
			}
			return domainErrors.NewInternalError("TARIFA_CREATE_ERROR", "error al crear tarifa")
		}

		for j := range t.Horarios {
			h := t.Horarios[j]
			if err := h.Validate(); err != nil {
				return err
			}

			_, err := tx.Exec(ctx, `
				INSERT INTO tarifa_horarios (
					tarifa_id, inicio, fin, monto
				) VALUES ($1, $2, $3, $4)
			`,
				tarifaID,
				h.Inicio.Format("15:04:05"),
				h.Fin.Format("15:04:05"),
				h.Monto,
			)
			if err != nil {
				if isUniqueViolation(err) {
					return domainErrors.NewConflictError("TARIFA_HORARIO_DUPLICADO", "horario de tarifa duplicado")
				}
				return domainErrors.NewInternalError("TARIFA_HORARIO_CREATE_ERROR", "error al crear horario de tarifa")
			}
		}
	}

	return nil
}

func (r *PostgresPorticoRepository) getTarifasByPorticoID(ctx context.Context, porticoID string) ([]entities.Tarifa, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, tipo_vehiculo, moneda
		FROM tarifas_portico
		WHERE portico_id = $1
		ORDER BY created_at ASC
	`, porticoID)
	if err != nil {
		return nil, domainErrors.NewInternalError("TARIFA_LIST_ERROR", "error al cargar tarifas")
	}
	defer rows.Close()

	tarifas := make([]entities.Tarifa, 0)
	for rows.Next() {
		var t entities.Tarifa
		if err := rows.Scan(&t.ID, &t.TipoVehiculo, &t.Moneda); err != nil {
			return nil, domainErrors.NewInternalError("TARIFA_SCAN_ERROR", "error al leer tarifas")
		}

		horarios, err := r.getHorariosByTarifaID(ctx, t.ID)
		if err != nil {
			return nil, err
		}
		t.Horarios = horarios

		tarifas = append(tarifas, t)
	}

	if err := rows.Err(); err != nil {
		return nil, domainErrors.NewInternalError("TARIFA_ROWS_ERROR", "error iterando tarifas")
	}

	return tarifas, nil
}

func (r *PostgresPorticoRepository) getHorariosByTarifaID(ctx context.Context, tarifaID string) ([]entities.TarifaHorario, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, inicio, fin, monto
		FROM tarifa_horarios
		WHERE tarifa_id = $1
		ORDER BY inicio ASC
	`, tarifaID)
	if err != nil {
		return nil, domainErrors.NewInternalError("TARIFA_HORARIO_LIST_ERROR", "error al cargar horarios de tarifa")
	}
	defer rows.Close()

	horarios := make([]entities.TarifaHorario, 0)
	for rows.Next() {
		var h entities.TarifaHorario
		if err := rows.Scan(&h.ID, &h.Inicio, &h.Fin, &h.Monto); err != nil {
			return nil, domainErrors.NewInternalError("TARIFA_HORARIO_SCAN_ERROR", "error al leer horarios de tarifa")
		}
		horarios = append(horarios, h)
	}

	if err := rows.Err(); err != nil {
		return nil, domainErrors.NewInternalError("TARIFA_HORARIO_ROWS_ERROR", "error iterando horarios de tarifa")
	}

	return horarios, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505"
}
