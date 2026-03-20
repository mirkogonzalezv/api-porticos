package data

import (
	"context"
	"sync"
	"time"

	"rea/porticos/internal/modules/tracking/domain/entities"
)

type memorySessionEntry struct {
	session   *entities.TrackingSession
	expiresAt time.Time
}

type memoryLastPassEntry struct {
	ts        time.Time
	expiresAt time.Time
}

type TrackingMemoryRepository struct {
	mu       sync.Mutex
	sessions map[string]memorySessionEntry
	lastPass map[string]memoryLastPassEntry
}

func NewTrackingMemoryRepository() *TrackingMemoryRepository {
	return &TrackingMemoryRepository{
		sessions: make(map[string]memorySessionEntry),
		lastPass: make(map[string]memoryLastPassEntry),
	}
}

func (r *TrackingMemoryRepository) GetSession(ctx context.Context, vehiculoID, porticoID string) (*entities.TrackingSession, error) {
	_ = ctx
	key := sessionKey(vehiculoID, porticoID)
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.sessions[key]
	if !ok {
		return nil, nil
	}
	if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
		delete(r.sessions, key)
		return nil, nil
	}
	return entry.session, nil
}

func (r *TrackingMemoryRepository) SetSession(ctx context.Context, session *entities.TrackingSession, ttl time.Duration) error {
	_ = ctx
	key := sessionKey(session.VehiculoID, session.PorticoID)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[key] = memorySessionEntry{
		session:   session,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (r *TrackingMemoryRepository) DeleteSession(ctx context.Context, vehiculoID, porticoID string) error {
	_ = ctx
	key := sessionKey(vehiculoID, porticoID)

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.sessions, key)
	return nil
}

func (r *TrackingMemoryRepository) GetLastPass(ctx context.Context, vehiculoID, porticoID string) (time.Time, bool, error) {
	_ = ctx
	key := lastPassKey(vehiculoID, porticoID)
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.lastPass[key]
	if !ok {
		return time.Time{}, false, nil
	}
	if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
		delete(r.lastPass, key)
		return time.Time{}, false, nil
	}
	return entry.ts, true, nil
}

func (r *TrackingMemoryRepository) SetLastPass(ctx context.Context, vehiculoID, porticoID string, ts time.Time, ttl time.Duration) error {
	_ = ctx
	key := lastPassKey(vehiculoID, porticoID)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastPass[key] = memoryLastPassEntry{
		ts:        ts,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}
