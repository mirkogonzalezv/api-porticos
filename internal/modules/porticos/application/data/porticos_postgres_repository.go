package data

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"

	"rea/porticos/internal/modules/porticos/domain/entities"
	"rea/porticos/internal/modules/porticos/domain/repository"
	domainErrors "rea/porticos/pkg/errors"
	"rea/porticos/pkg/logger"

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
			codigo, nombre, concesionaria_id, latitude, longitude, bearing, detection_radius_meters, entry_radius_meters, exit_radius_meters,
			entry_latitude, entry_longitude, exit_latitude, exit_longitude, max_crossing_seconds, tipo,
			direccion, velocidad_maxima, zona_de_deteccion, vehicle_types, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15,
			$16, $17, ST_GeogFromText($18), $19, $20
		)
		RETURNING id::text
	`,
		portico.Codigo,
		portico.Nombre,
		portico.ConcesionariaID,
		portico.Latitude,
		portico.Longitude,
		portico.Bearing,
		portico.DetectionRadiusMeters,
		portico.EntryRadiusMeters,
		portico.ExitRadiusMeters,
		portico.EntryLatitude,
		portico.EntryLongitude,
		portico.ExitLatitude,
		portico.ExitLongitude,
		portico.MaxCrossingSeconds,
		portico.Tipo,
		portico.Direccion,
		portico.VelocidadMaxima,
		nullableString(portico.ZonaDeteccionWKT),
		encodeVehicleTypes(portico.VehicleTypes),
		portico.IsActive,
	).Scan(&portico.ID)
	if err != nil {
		logger.Error("PORTICO_CREATE_ERROR: " + err.Error())
		if isUniqueViolation(err) {
			return nil, domainErrors.NewConflictError("PORTICO_CODIGO_DUPLICADO", "ya existe un pórtico con ese código")
		}
		return nil, domainErrors.NewInternalError("PORTICO_CREATE_ERROR", "error al crear pórtico")
	}

	if err := r.insertTarifasTx(ctx, tx, portico.ID, portico.Tarifas); err != nil {
		return nil, err
	}
	if err := r.insertViasTx(ctx, tx, portico.ID, portico.Vias); err != nil {
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
			p.detection_radius_meters,
			p.entry_radius_meters,
			p.exit_radius_meters,
			p.entry_latitude,
			p.entry_longitude,
			p.exit_latitude,
			p.exit_longitude,
			p.max_crossing_seconds,
			p.tipo,
			p.direccion,
			p.velocidad_maxima,
			ST_AsText(p.zona_de_deteccion) AS zona_wkt,
			p.vehicle_types,
			p.is_active
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
		var vehicleTypesRaw []byte
		if err := rows.Scan(
			&p.ID,
			&p.Codigo,
			&p.Nombre,
			&p.ConcesionariaID,
			&p.Concesionaria,
			&p.Latitude,
			&p.Longitude,
			&p.Bearing,
			&p.DetectionRadiusMeters,
			&p.EntryRadiusMeters,
			&p.ExitRadiusMeters,
			&p.EntryLatitude,
			&p.EntryLongitude,
			&p.ExitLatitude,
			&p.ExitLongitude,
			&p.MaxCrossingSeconds,
			&p.Tipo,
			&p.Direccion,
			&p.VelocidadMaxima,
			&p.ZonaDeteccionWKT,
			&vehicleTypesRaw,
			&p.IsActive,
		); err != nil {
			return nil, domainErrors.NewInternalError("PORTICO_LIST_SCAN_ERROR", "error al leer pórticos")
		}
		p.VehicleTypes = decodeVehicleTypes(vehicleTypesRaw)

		tarifas, err := r.getTarifasByPorticoID(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		vias, err := r.getViasByPorticoID(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		p.Vias = vias
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
	var vehicleTypesRaw []byte
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
			p.detection_radius_meters,
			p.entry_radius_meters,
			p.exit_radius_meters,
			p.entry_latitude,
			p.entry_longitude,
			p.exit_latitude,
			p.exit_longitude,
			p.max_crossing_seconds,
			p.tipo,
			p.direccion,
			p.velocidad_maxima,
			ST_AsText(p.zona_de_deteccion) AS zona_wkt,
			p.vehicle_types,
			p.is_active
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
		&p.DetectionRadiusMeters,
		&p.EntryRadiusMeters,
		&p.ExitRadiusMeters,
		&p.EntryLatitude,
		&p.EntryLongitude,
		&p.ExitLatitude,
		&p.ExitLongitude,
		&p.MaxCrossingSeconds,
		&p.Tipo,
		&p.Direccion,
		&p.VelocidadMaxima,
		&p.ZonaDeteccionWKT,
		&vehicleTypesRaw,
		&p.IsActive,
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
	vias, err := r.getViasByPorticoID(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.VehicleTypes = decodeVehicleTypes(vehicleTypesRaw)
	p.Tarifas = tarifas
	p.Vias = vias

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
			p.detection_radius_meters,
			p.entry_radius_meters,
			p.exit_radius_meters,
			p.entry_latitude,
			p.entry_longitude,
			p.exit_latitude,
			p.exit_longitude,
			p.max_crossing_seconds,
			p.tipo,
			p.direccion,
			p.velocidad_maxima,
			ST_AsText(p.zona_de_deteccion) AS zona_wkt,
			p.vehicle_types,
			p.is_active
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
		var vehicleTypesRaw []byte
		if err := rows.Scan(
			&p.ID,
			&p.Codigo,
			&p.Nombre,
			&p.ConcesionariaID,
			&p.Concesionaria,
			&p.Latitude,
			&p.Longitude,
			&p.Bearing,
			&p.DetectionRadiusMeters,
			&p.EntryRadiusMeters,
			&p.ExitRadiusMeters,
			&p.EntryLatitude,
			&p.EntryLongitude,
			&p.ExitLatitude,
			&p.ExitLongitude,
			&p.MaxCrossingSeconds,
			&p.Tipo,
			&p.Direccion,
			&p.VelocidadMaxima,
			&p.ZonaDeteccionWKT,
			&vehicleTypesRaw,
			&p.IsActive,
		); err != nil {
			return nil, domainErrors.NewInternalError("PORTICO_NEARBY_SCAN_ERROR", "error al leer pórticos cercanos")
		}
		p.VehicleTypes = decodeVehicleTypes(vehicleTypesRaw)
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, domainErrors.NewInternalError("PORTICO_NEARBY_ROWS_ERROR", "error iterando pórticos cercanos")
	}

	return out, nil
}

func (r *PostgresPorticoRepository) FindByTrajectory(ctx context.Context, lineWKT string) ([]entities.Portico, error) {
	lineWKT = strings.TrimSpace(lineWKT)
	if lineWKT == "" {
		return nil, domainErrors.NewValidationError("PORTICO_TRAJECTORY_REQUIRED", "trajectory es obligatoria")
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
			p.detection_radius_meters,
			p.entry_radius_meters,
			p.exit_radius_meters,
			p.entry_latitude,
			p.entry_longitude,
			p.exit_latitude,
			p.exit_longitude,
			p.max_crossing_seconds,
			p.tipo,
			p.direccion,
			p.velocidad_maxima,
			ST_AsText(p.zona_de_deteccion) AS zona_wkt,
			p.vehicle_types,
			p.is_active
		FROM porticos p
		LEFT JOIN concesionarias c ON c.id = p.concesionaria_id
		WHERE p.is_active = TRUE
		  AND p.zona_de_deteccion IS NOT NULL
		  AND ST_Intersects(p.zona_de_deteccion, ST_GeogFromText($1))
	`, lineWKT)
	if err != nil {
		return nil, domainErrors.NewInternalError("PORTICO_TRAJECTORY_ERROR", "error al buscar pórticos por trayectoria")
	}
	defer rows.Close()

	out := make([]entities.Portico, 0)
	for rows.Next() {
		var p entities.Portico
		var vehicleTypesRaw []byte
		if err := rows.Scan(
			&p.ID,
			&p.Codigo,
			&p.Nombre,
			&p.ConcesionariaID,
			&p.Concesionaria,
			&p.Latitude,
			&p.Longitude,
			&p.Bearing,
			&p.DetectionRadiusMeters,
			&p.EntryRadiusMeters,
			&p.ExitRadiusMeters,
			&p.EntryLatitude,
			&p.EntryLongitude,
			&p.ExitLatitude,
			&p.ExitLongitude,
			&p.MaxCrossingSeconds,
			&p.Tipo,
			&p.Direccion,
			&p.VelocidadMaxima,
			&p.ZonaDeteccionWKT,
			&vehicleTypesRaw,
			&p.IsActive,
		); err != nil {
			return nil, domainErrors.NewInternalError("PORTICO_TRAJECTORY_SCAN_ERROR", "error al leer pórticos por trayectoria")
		}
		p.VehicleTypes = decodeVehicleTypes(vehicleTypesRaw)

		tarifas, err := r.getTarifasByPorticoID(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		vias, err := r.getViasByPorticoID(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		p.Vias = vias
		p.Tarifas = tarifas
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, domainErrors.NewInternalError("PORTICO_TRAJECTORY_ROWS_ERROR", "error iterando pórticos por trayectoria")
	}

	return out, nil
}

func (r *PostgresPorticoRepository) GetByCodigo(ctx context.Context, codigo string) (*entities.Portico, error) {
	codigo = strings.TrimSpace(codigo)
	if codigo == "" {
		return nil, domainErrors.NewValidationError("PORTICO_CODIGO_REQUIRED", "codigo es obligatorio")
	}

	var p entities.Portico
	var vehicleTypesRaw []byte
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
			p.detection_radius_meters,
			p.entry_radius_meters,
			p.exit_radius_meters,
			p.entry_latitude,
			p.entry_longitude,
			p.exit_latitude,
			p.exit_longitude,
			p.max_crossing_seconds,
			p.tipo,
			p.direccion,
			p.velocidad_maxima,
			ST_AsText(p.zona_de_deteccion) AS zona_wkt,
			p.vehicle_types,
			p.is_active
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
		&p.DetectionRadiusMeters,
		&p.EntryRadiusMeters,
		&p.ExitRadiusMeters,
		&p.EntryLatitude,
		&p.EntryLongitude,
		&p.ExitLatitude,
		&p.ExitLongitude,
		&p.MaxCrossingSeconds,
		&p.Tipo,
		&p.Direccion,
		&p.VelocidadMaxima,
		&p.ZonaDeteccionWKT,
		&vehicleTypesRaw,
		&p.IsActive,
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
	vias, err := r.getViasByPorticoID(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.VehicleTypes = decodeVehicleTypes(vehicleTypesRaw)
	p.Tarifas = tarifas
	p.Vias = vias

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
			detection_radius_meters = $8,
			entry_radius_meters = $9,
			exit_radius_meters = $10,
			entry_latitude = $11,
			entry_longitude = $12,
			exit_latitude = $13,
			exit_longitude = $14,
			max_crossing_seconds = $15,
			tipo = $16,
			direccion = $17,
			velocidad_maxima = $18,
			zona_de_deteccion = ST_GeogFromText($19),
			vehicle_types = $20,
			is_active = $21,
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
		portico.DetectionRadiusMeters,
		portico.EntryRadiusMeters,
		portico.ExitRadiusMeters,
		portico.EntryLatitude,
		portico.EntryLongitude,
		portico.ExitLatitude,
		portico.ExitLongitude,
		portico.MaxCrossingSeconds,
		portico.Tipo,
		portico.Direccion,
		portico.VelocidadMaxima,
		nullableString(portico.ZonaDeteccionWKT),
		encodeVehicleTypes(portico.VehicleTypes),
		portico.IsActive,
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
	_, err = tx.Exec(ctx, `DELETE FROM portico_vias WHERE portico_id = $1`, portico.ID)
	if err != nil {
		return nil, domainErrors.NewInternalError("PORTICO_VIAS_DELETE_ERROR", "error al reemplazar vias")
	}

	if err := r.insertTarifasTx(ctx, tx, portico.ID, portico.Tarifas); err != nil {
		return nil, err
	}
	if err := r.insertViasTx(ctx, tx, portico.ID, portico.Vias); err != nil {
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

func encodeVehicleTypes(items []string) []byte {
	if len(items) == 0 {
		return []byte("[]")
	}
	raw, err := json.Marshal(items)
	if err != nil {
		return []byte("[]")
	}
	return raw
}

func decodeVehicleTypes(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func nullableString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (r *PostgresPorticoRepository) insertViasTx(
	ctx context.Context,
	tx pgx.Tx,
	porticoID string,
	vias []entities.Via,
) error {
	for i := range vias {
		v := vias[i]
		if err := v.Validate(); err != nil {
			return err
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO portico_vias (
				portico_id,
				way_name,
				direction_deg,
				center_line,
				entry_line,
				exit_line,
				entry_distance_m,
				exit_distance_m,
				auto_calculate,
				is_active
			) VALUES (
				$1, $2, $3,
				ST_GeogFromText($4),
				ST_GeogFromText($5),
				ST_GeogFromText($6),
				$7, $8, $9, $10
			)
		`,
			porticoID,
			v.WayName,
			v.DirectionDeg,
			nullableString(v.CenterLineWKT),
			nullableString(v.EntryLineWKT),
			nullableString(v.ExitLineWKT),
			v.EntryDistanceM,
			v.ExitDistanceM,
			v.AutoCalculate,
			v.IsActive,
		)
		if err != nil {
			logger.Error("PORTICO_VIA_CREATE_ERROR: " + err.Error())
			return domainErrors.NewInternalError("PORTICO_VIA_CREATE_ERROR", "error al crear via")
		}
	}

	return nil
}

func (r *PostgresPorticoRepository) getViasByPorticoID(ctx context.Context, porticoID string) ([]entities.Via, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			id::text,
			way_name,
			direction_deg,
			ST_AsText(center_line) AS center_wkt,
			ST_AsText(entry_line) AS entry_wkt,
			ST_AsText(exit_line) AS exit_wkt,
			entry_distance_m,
			exit_distance_m,
			auto_calculate,
			is_active
		FROM portico_vias
		WHERE portico_id = $1
		ORDER BY created_at ASC
	`, porticoID)
	if err != nil {
		return nil, domainErrors.NewInternalError("PORTICO_VIA_LIST_ERROR", "error al cargar vias")
	}
	defer rows.Close()

	out := make([]entities.Via, 0)
	for rows.Next() {
		var v entities.Via
		if err := rows.Scan(
			&v.ID,
			&v.WayName,
			&v.DirectionDeg,
			&v.CenterLineWKT,
			&v.EntryLineWKT,
			&v.ExitLineWKT,
			&v.EntryDistanceM,
			&v.ExitDistanceM,
			&v.AutoCalculate,
			&v.IsActive,
		); err != nil {
			return nil, domainErrors.NewInternalError("PORTICO_VIA_SCAN_ERROR", "error al leer vias")
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, domainErrors.NewInternalError("PORTICO_VIA_ROWS_ERROR", "error iterando vias")
	}

	return out, nil
}
