package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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

	sourceJSON, err := encodeSourcePosition(paso.SourcePosition)
	if err != nil {
		return nil, domainErrors.NewInternalError("PASO_SOURCE_JSON_ERROR", "error serializando posición de origen")
	}

	err = r.pool.QueryRow(ctx, `
		INSERT INTO pasos_portico (
			owner_supabase_user_id,
			vehiculo_id,
			portico_id,
			fecha_hora_paso,
			direccion_paso,
			entry_timestamp,
			exit_timestamp,
			latitud,
			longitud,
			heading,
			speed,
			monto_cobrado,
			moneda,
			fuente,
			source_position
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id::text
	`,
		paso.OwnerSupabaseUserID,
		paso.VehiculoID,
		paso.PorticoID,
		paso.FechaHoraPaso,
		nullableString(paso.DireccionPaso),
		paso.EntryTimestamp,
		paso.ExitTimestamp,
		paso.Latitud,
		paso.Longitud,
		paso.Heading,
		paso.Speed,
		paso.MontoCobrado,
		paso.Moneda,
		paso.Fuente,
		sourceJSON,
	).Scan(&paso.ID)
	if err != nil {
		return nil, domainErrors.NewInternalError("PASO_CREATE_ERROR", "error al registrar paso de pórtico")
	}

	return r.GetByID(ctx, paso.OwnerSupabaseUserID, paso.ID)
}

func (r *PasosPostgresRepository) CreateCapture(ctx context.Context, paso *entities.PasoCapturado) (*entities.PasoCapturado, error) {
	if paso == nil {
		return nil, domainErrors.NewValidationError("CAPTURA_REQUIRED", "captura es obligatoria")
	}
	if err := paso.ValidateForCreate(); err != nil {
		return nil, err
	}

	sourceJSON, err := encodeSourcePosition(paso.SourcePosition)
	if err != nil {
		return nil, domainErrors.NewInternalError("CAPTURA_SOURCE_JSON_ERROR", "error serializando posiciones")
	}

	err = r.pool.QueryRow(ctx, `
		INSERT INTO pasos_capturados (
			owner_supabase_user_id,
			vehiculo_id,
			portico_id,
			via_id,
			fecha_hora_inicio,
			fecha_hora_fin,
			entry_timestamp,
			exit_timestamp,
			entry_hit,
			exit_hit,
			heading_avg,
			speed_avg,
			direccion_paso,
			status,
			source_position
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15
		)
		RETURNING id::text
	`,
		paso.OwnerSupabaseUserID,
		paso.VehiculoID,
		paso.PorticoID,
		nullableString(paso.ViaID),
		paso.FechaHoraInicio,
		paso.FechaHoraFin,
		paso.EntryTimestamp,
		paso.ExitTimestamp,
		paso.EntryHit,
		paso.ExitHit,
		paso.HeadingAvg,
		paso.SpeedAvg,
		nullableString(paso.DireccionPaso),
		paso.Status,
		sourceJSON,
	).Scan(&paso.ID)
	if err != nil {
		return nil, domainErrors.NewInternalError("CAPTURA_CREATE_ERROR", "error al registrar captura")
	}

	return paso, nil
}

func (r *PasosPostgresRepository) CreateCapturesBatch(ctx context.Context, pasos []*entities.PasoCapturado) error {
	if len(pasos) == 0 {
		return nil
	}
	rows := make([][]any, 0, len(pasos))
	for i := range pasos {
		if pasos[i] == nil {
			return domainErrors.NewValidationError("CAPTURA_REQUIRED", "captura es obligatoria")
		}
		if err := pasos[i].ValidateForCreate(); err != nil {
			return err
		}
		sourceJSON, err := encodeSourcePosition(pasos[i].SourcePosition)
		if err != nil {
			return domainErrors.NewInternalError("CAPTURA_SOURCE_JSON_ERROR", "error serializando posiciones")
		}
		rows = append(rows, []any{
			pasos[i].OwnerSupabaseUserID,
			pasos[i].VehiculoID,
			pasos[i].PorticoID,
			nullableString(pasos[i].ViaID),
			pasos[i].FechaHoraInicio,
			pasos[i].FechaHoraFin,
			pasos[i].EntryTimestamp,
			pasos[i].ExitTimestamp,
			pasos[i].EntryHit,
			pasos[i].ExitHit,
			pasos[i].HeadingAvg,
			pasos[i].SpeedAvg,
			nullableString(pasos[i].DireccionPaso),
			pasos[i].Status,
			sourceJSON,
		})
	}

	_, err := r.pool.CopyFrom(
		ctx,
		pgx.Identifier{"pasos_capturados"},
		[]string{
			"owner_supabase_user_id",
			"vehiculo_id",
			"portico_id",
			"via_id",
			"fecha_hora_inicio",
			"fecha_hora_fin",
			"entry_timestamp",
			"exit_timestamp",
			"entry_hit",
			"exit_hit",
			"heading_avg",
			"speed_avg",
			"direccion_paso",
			"status",
			"source_position",
		},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return domainErrors.NewInternalError("CAPTURA_BATCH_ERROR", "error al registrar capturas")
	}
	return nil
}

func (r *PasosPostgresRepository) CreateBatch(ctx context.Context, pasos []*entities.PasoPortico) ([]entities.PasoPortico, error) {
	if len(pasos) == 0 {
		return nil, domainErrors.NewValidationError("PASO_BATCH_EMPTY", "items es obligatorio")
	}

	ownerID := strings.TrimSpace(pasos[0].OwnerSupabaseUserID)
	if ownerID == "" {
		return nil, domainErrors.NewValidationError("PASO_OWNER_REQUIRED", "usuario no autenticado")
	}

	for i := range pasos {
		if pasos[i] == nil {
			return nil, domainErrors.NewValidationError("PASO_REQUIRED", "paso es obligatorio")
		}
		if strings.TrimSpace(pasos[i].OwnerSupabaseUserID) != ownerID {
			return nil, domainErrors.NewValidationError("PASO_OWNER_MISMATCH", "todos los pasos deben pertenecer al mismo usuario")
		}
		if err := pasos[i].ValidateForCreate(); err != nil {
			return nil, err
		}
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, domainErrors.NewInternalError("PASO_TX_BEGIN_ERROR", "no se pudo iniciar transacción")
	}
	defer tx.Rollback(ctx)

	ids := make([]string, 0, len(pasos))
	for i := range pasos {
		sourceJSON, err := encodeSourcePosition(pasos[i].SourcePosition)
		if err != nil {
			return nil, domainErrors.NewInternalError("PASO_SOURCE_JSON_ERROR", "error serializando posición de origen")
		}
		var id string
		err = tx.QueryRow(ctx, `
			INSERT INTO pasos_portico (
				owner_supabase_user_id,
				vehiculo_id,
				portico_id,
				fecha_hora_paso,
				direccion_paso,
				entry_timestamp,
				exit_timestamp,
				latitud,
				longitud,
				heading,
				speed,
				monto_cobrado,
				moneda,
				fuente,
				source_position
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			RETURNING id::text
		`,
			pasos[i].OwnerSupabaseUserID,
			pasos[i].VehiculoID,
			pasos[i].PorticoID,
			pasos[i].FechaHoraPaso,
			nullableString(pasos[i].DireccionPaso),
			pasos[i].EntryTimestamp,
			pasos[i].ExitTimestamp,
			pasos[i].Latitud,
			pasos[i].Longitud,
			pasos[i].Heading,
			pasos[i].Speed,
			pasos[i].MontoCobrado,
			pasos[i].Moneda,
			pasos[i].Fuente,
			sourceJSON,
		).Scan(&id)
		if err != nil {
			return nil, domainErrors.NewInternalError("PASO_CREATE_ERROR", "error al registrar paso de pórtico")
		}
		ids = append(ids, id)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, domainErrors.NewInternalError("PASO_TX_COMMIT_ERROR", "no se pudo confirmar transacción")
	}

	return r.fetchByIDs(ctx, ownerID, ids)
}

func (r *PasosPostgresRepository) CreateConfirmadosBatch(ctx context.Context, pasos []*entities.PasoPortico) error {
	if len(pasos) == 0 {
		return nil
	}
	rows := make([][]any, 0, len(pasos))
	for i := range pasos {
		if pasos[i] == nil {
			return domainErrors.NewValidationError("PASO_REQUIRED", "paso es obligatorio")
		}
		if err := pasos[i].ValidateForCreate(); err != nil {
			return err
		}
		sourceJSON, err := encodeSourcePosition(pasos[i].SourcePosition)
		if err != nil {
			return domainErrors.NewInternalError("PASO_SOURCE_JSON_ERROR", "error serializando posición de origen")
		}
		rows = append(rows, []any{
			pasos[i].OwnerSupabaseUserID,
			pasos[i].VehiculoID,
			pasos[i].PorticoID,
			pasos[i].FechaHoraPaso,
			nullableString(pasos[i].DireccionPaso),
			pasos[i].EntryTimestamp,
			pasos[i].ExitTimestamp,
			pasos[i].Latitud,
			pasos[i].Longitud,
			pasos[i].Heading,
			pasos[i].Speed,
			pasos[i].MontoCobrado,
			pasos[i].Moneda,
			pasos[i].Fuente,
			sourceJSON,
		})
	}

	_, err := r.pool.CopyFrom(
		ctx,
		pgx.Identifier{"pasos_portico"},
		[]string{
			"owner_supabase_user_id",
			"vehiculo_id",
			"portico_id",
			"fecha_hora_paso",
			"direccion_paso",
			"entry_timestamp",
			"exit_timestamp",
			"latitud",
			"longitud",
			"heading",
			"speed",
			"monto_cobrado",
			"moneda",
			"fuente",
			"source_position",
		},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return domainErrors.NewInternalError("PASO_BATCH_ERROR", "error al registrar pasos confirmados")
	}
	return nil
}

func (r *PasosPostgresRepository) AcquireIdempotencyKey(
	ctx context.Context,
	ownerID, key, scope string,
	ttl time.Duration,
) (bool, error) {
	ownerID = strings.TrimSpace(ownerID)
	key = strings.TrimSpace(key)
	scope = strings.TrimSpace(scope)
	if ownerID == "" || key == "" || scope == "" {
		return true, nil
	}
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	tag, err := r.pool.Exec(ctx, `
		INSERT INTO idempotency_keys (
			owner_supabase_user_id,
			idempotency_key,
			scope,
			expires_at
		) VALUES ($1, $2, $3, NOW() + $4::interval)
		ON CONFLICT (owner_supabase_user_id, idempotency_key, scope) DO NOTHING
	`, ownerID, key, scope, fmt.Sprintf("%d seconds", int(ttl.Seconds())))
	if err != nil {
		return false, domainErrors.NewInternalError("IDEMPOTENCY_INSERT_ERROR", "error registrando idempotency key")
	}
	if tag.RowsAffected() == 1 {
		return true, nil
	}

	var expiresAt time.Time
	err = r.pool.QueryRow(ctx, `
		SELECT expires_at
		FROM idempotency_keys
		WHERE owner_supabase_user_id = $1
		  AND idempotency_key = $2
		  AND scope = $3
		LIMIT 1
	`, ownerID, key, scope).Scan(&expiresAt)
	if err != nil {
		return false, domainErrors.NewInternalError("IDEMPOTENCY_READ_ERROR", "error leyendo idempotency key")
	}
	if time.Now().After(expiresAt) {
		_, err = r.pool.Exec(ctx, `
			UPDATE idempotency_keys
			SET expires_at = NOW() + $4::interval
			WHERE owner_supabase_user_id = $1
			  AND idempotency_key = $2
			  AND scope = $3
		`, ownerID, key, scope, fmt.Sprintf("%d seconds", int(ttl.Seconds())))
		if err != nil {
			return false, domainErrors.NewInternalError("IDEMPOTENCY_UPDATE_ERROR", "error actualizando idempotency key")
		}
		return true, nil
	}

	return false, nil
}

func (r *PasosPostgresRepository) GetByID(ctx context.Context, ownerID, id string) (*entities.PasoPortico, error) {
	ownerID = strings.TrimSpace(ownerID)
	id = strings.TrimSpace(id)
	if ownerID == "" || id == "" {
		return nil, domainErrors.NewValidationError("PASO_REQUIRED_FIELDS", "usuario e id son obligatorios")
	}

	var out entities.PasoPortico
	var sourceBytes []byte
	err := r.pool.QueryRow(ctx, `
		SELECT
			id::text,
			owner_supabase_user_id::text,
			vehiculo_id::text,
			portico_id::text,
			fecha_hora_paso,
			direccion_paso,
			entry_timestamp,
			exit_timestamp,
			latitud,
			longitud,
			heading,
			speed,
			monto_cobrado,
			moneda,
			fuente,
			source_position,
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
		&out.DireccionPaso,
		&out.EntryTimestamp,
		&out.ExitTimestamp,
		&out.Latitud,
		&out.Longitud,
		&out.Heading,
		&out.Speed,
		&out.MontoCobrado,
		&out.Moneda,
		&out.Fuente,
		&sourceBytes,
		&out.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domainErrors.NewNotFoundError("PASO_NOT_FOUND", "paso no encontrado")
		}
		return nil, domainErrors.NewInternalError("PASO_GET_ERROR", "error al obtener paso")
	}

	out.SourcePosition = decodeSourcePosition(sourceBytes)
	return &out, nil
}

func (r *PasosPostgresRepository) fetchByIDs(ctx context.Context, ownerID string, ids []string) ([]entities.PasoPortico, error) {
	if len(ids) == 0 {
		return []entities.PasoPortico{}, nil
	}

	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids)+1)
	args = append(args, ownerID)
	for i, id := range ids {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT
			id::text,
			owner_supabase_user_id::text,
			vehiculo_id::text,
			portico_id::text,
			fecha_hora_paso,
			direccion_paso,
			entry_timestamp,
			exit_timestamp,
			latitud,
			longitud,
			heading,
			speed,
			monto_cobrado,
			moneda,
			fuente,
			source_position,
			created_at
		FROM pasos_portico
		WHERE owner_supabase_user_id = $1
		  AND id::text IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, domainErrors.NewInternalError("PASO_BATCH_FETCH_ERROR", "error al obtener pasos creados")
	}
	defer rows.Close()

	byID := make(map[string]entities.PasoPortico, len(ids))
	for rows.Next() {
		var item entities.PasoPortico
		var sourceBytes []byte
		if err := rows.Scan(
			&item.ID,
			&item.OwnerSupabaseUserID,
			&item.VehiculoID,
			&item.PorticoID,
			&item.FechaHoraPaso,
			&item.DireccionPaso,
			&item.EntryTimestamp,
			&item.ExitTimestamp,
			&item.Latitud,
			&item.Longitud,
			&item.Heading,
			&item.Speed,
			&item.MontoCobrado,
			&item.Moneda,
			&item.Fuente,
			&sourceBytes,
			&item.CreatedAt,
		); err != nil {
			return nil, domainErrors.NewInternalError("PASO_BATCH_FETCH_SCAN_ERROR", "error al leer pasos creados")
		}
		item.SourcePosition = decodeSourcePosition(sourceBytes)
		byID[item.ID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, domainErrors.NewInternalError("PASO_BATCH_FETCH_ROWS_ERROR", "error iterando pasos creados")
	}

	out := make([]entities.PasoPortico, 0, len(ids))
	for _, id := range ids {
		item, ok := byID[id]
		if !ok {
			return nil, domainErrors.NewInternalError("PASO_BATCH_FETCH_MISSING", "resultado incompleto al leer pasos creados")
		}
		out = append(out, item)
	}

	return out, nil
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
			pp.id::text,
			pp.owner_supabase_user_id::text,
			pp.vehiculo_id::text,
			v.patente,
			pp.portico_id::text,
			p.codigo,
			c.nombre AS concesionaria_nombre,
			pp.fecha_hora_paso,
			pp.direccion_paso,
			pp.entry_timestamp,
			pp.exit_timestamp,
			pp.latitud,
			pp.longitud,
			pp.heading,
			pp.speed,
			pp.monto_cobrado,
			pp.moneda,
			pp.fuente,
			pp.source_position,
			pp.created_at
		FROM pasos_portico pp
		JOIN vehiculos v ON v.id = pp.vehiculo_id
		JOIN porticos p ON p.id = pp.portico_id
		LEFT JOIN concesionarias c ON c.id = p.concesionaria_id
		WHERE pp.owner_supabase_user_id = $1
		  AND pp.fecha_hora_paso >= $2
		  AND pp.fecha_hora_paso <= $3
		  AND ($4 = '' OR pp.vehiculo_id::text = $4)
		  AND ($5 = '' OR pp.portico_id::text = $5)
		ORDER BY pp.fecha_hora_paso DESC
		LIMIT $6 OFFSET $7
	`, ownerID, filter.From, filter.To, vehiculoID, porticoID, limit, offset)
	if err != nil {
		return nil, domainErrors.NewInternalError("PASO_LIST_ERROR", "error al listar pasos")
	}
	defer rows.Close()

	out := make([]entities.PasoPortico, 0)
	for rows.Next() {
		var item entities.PasoPortico
		var sourceBytes []byte
		if err := rows.Scan(
			&item.ID,
			&item.OwnerSupabaseUserID,
			&item.VehiculoID,
			&item.VehiculoPatente,
			&item.PorticoID,
			&item.PorticoCodigo,
			&item.ConcesionariaNombre,
			&item.FechaHoraPaso,
			&item.DireccionPaso,
			&item.EntryTimestamp,
			&item.ExitTimestamp,
			&item.Latitud,
			&item.Longitud,
			&item.Heading,
			&item.Speed,
			&item.MontoCobrado,
			&item.Moneda,
			&item.Fuente,
			&sourceBytes,
			&item.CreatedAt,
		); err != nil {
			return nil, domainErrors.NewInternalError("PASO_LIST_SCAN_ERROR", "error al leer pasos")
		}
		item.SourcePosition = decodeSourcePosition(sourceBytes)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, domainErrors.NewInternalError("PASO_LIST_ROWS_ERROR", "error iterando pasos")
	}

	return out, nil
}

func (r *PasosPostgresRepository) ListCapturadosByOwnerRange(
	ctx context.Context,
	ownerID string,
	filter repository.ListPasosFilter,
) ([]entities.PasoCapturado, error) {
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return nil, domainErrors.NewValidationError("CAPTURA_OWNER_REQUIRED", "usuario no autenticado")
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
			pc.id::text,
			pc.owner_supabase_user_id::text,
			pc.vehiculo_id::text,
			v.patente,
			pc.portico_id::text,
			p.codigo,
			c.nombre AS concesionaria_nombre,
			pc.via_id::text,
			pc.fecha_hora_inicio,
			pc.fecha_hora_fin,
			pc.entry_timestamp,
			pc.exit_timestamp,
			pc.entry_hit,
			pc.exit_hit,
			pc.heading_avg,
			pc.speed_avg,
			pc.direccion_paso,
			pc.status,
			pc.source_position,
			pc.created_at
		FROM pasos_capturados pc
		JOIN vehiculos v ON v.id = pc.vehiculo_id
		JOIN porticos p ON p.id = pc.portico_id
		LEFT JOIN concesionarias c ON c.id = p.concesionaria_id
		WHERE pc.owner_supabase_user_id = $1
		  AND pc.fecha_hora_inicio >= $2
		  AND pc.fecha_hora_fin <= $3
		  AND ($4 = '' OR pc.vehiculo_id::text = $4)
		  AND ($5 = '' OR pc.portico_id::text = $5)
		ORDER BY pc.fecha_hora_inicio DESC
		LIMIT $6 OFFSET $7
	`, ownerID, filter.From, filter.To, vehiculoID, porticoID, limit, offset)
	if err != nil {
		return nil, domainErrors.NewInternalError("CAPTURA_LIST_ERROR", "error al listar capturas")
	}
	defer rows.Close()

	out := make([]entities.PasoCapturado, 0)
	for rows.Next() {
		var item entities.PasoCapturado
		var sourceBytes []byte
		if err := rows.Scan(
			&item.ID,
			&item.OwnerSupabaseUserID,
			&item.VehiculoID,
			&item.VehiculoPatente,
			&item.PorticoID,
			&item.PorticoCodigo,
			&item.ConcesionariaNombre,
			&item.ViaID,
			&item.FechaHoraInicio,
			&item.FechaHoraFin,
			&item.EntryTimestamp,
			&item.ExitTimestamp,
			&item.EntryHit,
			&item.ExitHit,
			&item.HeadingAvg,
			&item.SpeedAvg,
			&item.DireccionPaso,
			&item.Status,
			&sourceBytes,
			&item.CreatedAt,
		); err != nil {
			return nil, domainErrors.NewInternalError("CAPTURA_LIST_SCAN_ERROR", "error al leer capturas")
		}
		item.SourcePosition = decodeSourcePosition(sourceBytes)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, domainErrors.NewInternalError("CAPTURA_LIST_ROWS_ERROR", "error iterando capturas")
	}

	return out, nil
}

func (r *PasosPostgresRepository) ListCapturadosAllRange(
	ctx context.Context,
	filter repository.ListPasosFilter,
) ([]entities.PasoCapturado, error) {
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
			pc.id::text,
			pc.owner_supabase_user_id::text,
			pc.vehiculo_id::text,
			v.patente,
			pc.portico_id::text,
			p.codigo,
			c.nombre AS concesionaria_nombre,
			pc.via_id::text,
			pc.fecha_hora_inicio,
			pc.fecha_hora_fin,
			pc.entry_timestamp,
			pc.exit_timestamp,
			pc.entry_hit,
			pc.exit_hit,
			pc.heading_avg,
			pc.speed_avg,
			pc.direccion_paso,
			pc.status,
			pc.source_position,
			pc.created_at
		FROM pasos_capturados pc
		JOIN vehiculos v ON v.id = pc.vehiculo_id
		JOIN porticos p ON p.id = pc.portico_id
		LEFT JOIN concesionarias c ON c.id = p.concesionaria_id
		WHERE pc.fecha_hora_inicio >= $1
		  AND pc.fecha_hora_fin <= $2
		  AND ($3 = '' OR pc.vehiculo_id::text = $3)
		  AND ($4 = '' OR pc.portico_id::text = $4)
		ORDER BY pc.fecha_hora_inicio DESC
		LIMIT $5 OFFSET $6
	`, filter.From, filter.To, vehiculoID, porticoID, limit, offset)
	if err != nil {
		return nil, domainErrors.NewInternalError("CAPTURA_LIST_ERROR", "error al listar capturas")
	}
	defer rows.Close()

	out := make([]entities.PasoCapturado, 0)
	for rows.Next() {
		var item entities.PasoCapturado
		var sourceBytes []byte
		if err := rows.Scan(
			&item.ID,
			&item.OwnerSupabaseUserID,
			&item.VehiculoID,
			&item.VehiculoPatente,
			&item.PorticoID,
			&item.PorticoCodigo,
			&item.ConcesionariaNombre,
			&item.ViaID,
			&item.FechaHoraInicio,
			&item.FechaHoraFin,
			&item.EntryTimestamp,
			&item.ExitTimestamp,
			&item.EntryHit,
			&item.ExitHit,
			&item.HeadingAvg,
			&item.SpeedAvg,
			&item.DireccionPaso,
			&item.Status,
			&sourceBytes,
			&item.CreatedAt,
		); err != nil {
			return nil, domainErrors.NewInternalError("CAPTURA_LIST_SCAN_ERROR", "error al leer capturas")
		}
		item.SourcePosition = decodeSourcePosition(sourceBytes)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, domainErrors.NewInternalError("CAPTURA_LIST_ROWS_ERROR", "error iterando capturas")
	}

	return out, nil
}

func (r *PasosPostgresRepository) ListAllRange(
	ctx context.Context,
	filter repository.ListPasosFilter,
) ([]entities.PasoPortico, error) {
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
			pp.id::text,
			pp.owner_supabase_user_id::text,
			pp.vehiculo_id::text,
			v.patente,
			pp.portico_id::text,
			p.codigo,
			c.nombre AS concesionaria_nombre,
			pp.fecha_hora_paso,
			pp.direccion_paso,
			pp.entry_timestamp,
			pp.exit_timestamp,
			pp.latitud,
			pp.longitud,
			pp.heading,
			pp.speed,
			pp.monto_cobrado,
			pp.moneda,
			pp.fuente,
			pp.source_position,
			pp.created_at
		FROM pasos_portico pp
		JOIN vehiculos v ON v.id = pp.vehiculo_id
		JOIN porticos p ON p.id = pp.portico_id
		LEFT JOIN concesionarias c ON c.id = p.concesionaria_id
		WHERE pp.fecha_hora_paso >= $1
		  AND pp.fecha_hora_paso <= $2
		  AND ($3 = '' OR pp.vehiculo_id::text = $3)
		  AND ($4 = '' OR pp.portico_id::text = $4)
		ORDER BY pp.fecha_hora_paso DESC
		LIMIT $5 OFFSET $6
	`, filter.From, filter.To, vehiculoID, porticoID, limit, offset)
	if err != nil {
		return nil, domainErrors.NewInternalError("PASO_LIST_ERROR", "error al listar pasos")
	}
	defer rows.Close()

	out := make([]entities.PasoPortico, 0)
	for rows.Next() {
		var item entities.PasoPortico
		var sourceBytes []byte
		if err := rows.Scan(
			&item.ID,
			&item.OwnerSupabaseUserID,
			&item.VehiculoID,
			&item.VehiculoPatente,
			&item.PorticoID,
			&item.PorticoCodigo,
			&item.ConcesionariaNombre,
			&item.FechaHoraPaso,
			&item.DireccionPaso,
			&item.EntryTimestamp,
			&item.ExitTimestamp,
			&item.Latitud,
			&item.Longitud,
			&item.Heading,
			&item.Speed,
			&item.MontoCobrado,
			&item.Moneda,
			&item.Fuente,
			&sourceBytes,
			&item.CreatedAt,
		); err != nil {
			return nil, domainErrors.NewInternalError("PASO_LIST_SCAN_ERROR", "error al leer pasos")
		}
		item.SourcePosition = decodeSourcePosition(sourceBytes)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, domainErrors.NewInternalError("PASO_LIST_ROWS_ERROR", "error iterando pasos")
	}

	return out, nil
}

func encodeSourcePosition(value any) (*string, error) {
	if value == nil {
		return nil, nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	payload := string(raw)
	return &payload, nil
}

func decodeSourcePosition(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	var out any
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
