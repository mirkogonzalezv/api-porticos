package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"rea/porticos/internal/modules/tracking/domain/entities"
	domainErrors "rea/porticos/pkg/errors"

	"github.com/redis/go-redis/v9"
)

type TrackingStore interface {
	GetSession(ctx context.Context, vehiculoID, porticoID string) (*entities.TrackingSession, error)
	SetSession(ctx context.Context, session *entities.TrackingSession, ttl time.Duration) error
	DeleteSession(ctx context.Context, vehiculoID, porticoID string) error
	GetLastPass(ctx context.Context, vehiculoID, porticoID string) (time.Time, bool, error)
	SetLastPass(ctx context.Context, vehiculoID, porticoID string, ts time.Time, ttl time.Duration) error
}

type TrackingRedisRepository struct {
	client *redis.Client
}

func NewTrackingRedisRepository(client *redis.Client) *TrackingRedisRepository {
	return &TrackingRedisRepository{client: client}
}

func (r *TrackingRedisRepository) GetSession(ctx context.Context, vehiculoID, porticoID string) (*entities.TrackingSession, error) {
	key := sessionKey(vehiculoID, porticoID)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, domainErrors.NewInternalError("TRACKING_REDIS_GET_ERROR", "error consultando sesión")
	}
	var s entities.TrackingSession
	if err := json.Unmarshal([]byte(val), &s); err != nil {
		return nil, domainErrors.NewInternalError("TRACKING_REDIS_PARSE_ERROR", "error parseando sesión")
	}
	return &s, nil
}

func (r *TrackingRedisRepository) SetSession(ctx context.Context, session *entities.TrackingSession, ttl time.Duration) error {
	key := sessionKey(session.VehiculoID, session.PorticoID)
	payload, err := json.Marshal(session)
	if err != nil {
		return domainErrors.NewInternalError("TRACKING_REDIS_SERIALIZE_ERROR", "error serializando sesión")
	}
	if err := r.client.Set(ctx, key, payload, ttl).Err(); err != nil {
		return domainErrors.NewInternalError("TRACKING_REDIS_SET_ERROR", "error guardando sesión")
	}
	return nil
}

func (r *TrackingRedisRepository) DeleteSession(ctx context.Context, vehiculoID, porticoID string) error {
	key := sessionKey(vehiculoID, porticoID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return domainErrors.NewInternalError("TRACKING_REDIS_DEL_ERROR", "error eliminando sesión")
	}
	return nil
}

func (r *TrackingRedisRepository) GetLastPass(ctx context.Context, vehiculoID, porticoID string) (time.Time, bool, error) {
	key := lastPassKey(vehiculoID, porticoID)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, domainErrors.NewInternalError("TRACKING_REDIS_GET_ERROR", "error consultando último paso")
	}
	parsed, err := time.Parse(time.RFC3339Nano, val)
	if err != nil {
		return time.Time{}, false, domainErrors.NewInternalError("TRACKING_REDIS_PARSE_ERROR", "error parseando último paso")
	}
	return parsed, true, nil
}

func (r *TrackingRedisRepository) SetLastPass(ctx context.Context, vehiculoID, porticoID string, ts time.Time, ttl time.Duration) error {
	key := lastPassKey(vehiculoID, porticoID)
	if err := r.client.Set(ctx, key, ts.Format(time.RFC3339Nano), ttl).Err(); err != nil {
		return domainErrors.NewInternalError("TRACKING_REDIS_SET_ERROR", "error guardando último paso")
	}
	return nil
}

func sessionKey(vehiculoID, porticoID string) string {
	return fmt.Sprintf("track:%s:%s", vehiculoID, porticoID)
}

func lastPassKey(vehiculoID, porticoID string) string {
	return fmt.Sprintf("lastpass:%s:%s", vehiculoID, porticoID)
}
