package data

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	"rea/porticos/internal/modules/porticos/domain/entities"
	"rea/porticos/internal/modules/porticos/domain/repository"
)

type cacheItem struct {
	key       string
	value     any
	expiresAt time.Time
}

type lruTTLCache struct {
	mu         sync.Mutex
	maxEntries int
	ttl        time.Duration
	ll         *list.List
	items      map[string]*list.Element
}

func newLRUTTLCache(maxEntries int, ttl time.Duration) *lruTTLCache {
	if maxEntries <= 0 {
		maxEntries = 200
	}
	if ttl <= 0 {
		ttl = 20 * time.Second
	}
	return &lruTTLCache{
		maxEntries: maxEntries,
		ttl:        ttl,
		ll:         list.New(),
		items:      make(map[string]*list.Element, maxEntries),
	}
}

func (c *lruTTLCache) get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return nil, false
	}
	item := elem.Value.(*cacheItem)
	if time.Now().After(item.expiresAt) {
		c.removeElement(elem)
		return nil, false
	}
	c.ll.MoveToFront(elem)
	return item.value, true
}

func (c *lruTTLCache) set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		item := elem.Value.(*cacheItem)
		item.value = value
		item.expiresAt = time.Now().Add(c.ttl)
		c.ll.MoveToFront(elem)
		return
	}

	elem := c.ll.PushFront(&cacheItem{
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	})
	c.items[key] = elem

	if c.ll.Len() > c.maxEntries {
		c.removeOldest()
	}
}

func (c *lruTTLCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ll.Init()
	c.items = make(map[string]*list.Element, c.maxEntries)
}

func (c *lruTTLCache) removeOldest() {
	elem := c.ll.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

func (c *lruTTLCache) removeElement(elem *list.Element) {
	c.ll.Remove(elem)
	item := elem.Value.(*cacheItem)
	delete(c.items, item.key)
}

type CachedPorticoRepository struct {
	inner repository.PorticoRepository
	byID  *lruTTLCache
	byKey *lruTTLCache
}

func NewCachedPorticoRepository(inner repository.PorticoRepository, ttl time.Duration, maxEntries int) repository.PorticoRepository {
	return &CachedPorticoRepository{
		inner: inner,
		byID:  newLRUTTLCache(maxEntries, ttl),
		byKey: newLRUTTLCache(maxEntries, ttl),
	}
}

func (r *CachedPorticoRepository) Create(ctx context.Context, portico *entities.Portico) (*entities.Portico, error) {
	created, err := r.inner.Create(ctx, portico)
	if err != nil {
		return nil, err
	}
	r.invalidateAll()
	return clonePortico(created), nil
}

func (r *CachedPorticoRepository) List(ctx context.Context, filter repository.ListPorticosFilter) ([]entities.Portico, error) {
	key := fmt.Sprintf("list:%d:%d", filter.Limit, filter.Offset)
	if cached, ok := r.byKey.get(key); ok {
		if out, okCast := cached.([]entities.Portico); okCast {
			return clonePorticos(out), nil
		}
	}

	items, err := r.inner.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	cloned := clonePorticos(items)
	r.byKey.set(key, cloned)
	return cloned, nil
}

func (r *CachedPorticoRepository) GetByID(ctx context.Context, id string) (*entities.Portico, error) {
	key := "id:" + id
	if cached, ok := r.byID.get(key); ok {
		if out, okCast := cached.(*entities.Portico); okCast {
			return clonePortico(out), nil
		}
	}

	item, err := r.inner.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	cloned := clonePortico(item)
	r.byID.set(key, cloned)
	return cloned, nil
}

func (r *CachedPorticoRepository) GetByCodigo(ctx context.Context, codigo string) (*entities.Portico, error) {
	key := "codigo:" + codigo
	if cached, ok := r.byID.get(key); ok {
		if out, okCast := cached.(*entities.Portico); okCast {
			return clonePortico(out), nil
		}
	}

	item, err := r.inner.GetByCodigo(ctx, codigo)
	if err != nil {
		return nil, err
	}
	cloned := clonePortico(item)
	r.byID.set(key, cloned)
	return cloned, nil
}

func (r *CachedPorticoRepository) ListNearby(ctx context.Context, lat, lng, maxDistanceMeters float64) ([]entities.Portico, error) {
	return r.inner.ListNearby(ctx, lat, lng, maxDistanceMeters)
}

func (r *CachedPorticoRepository) FindByTrajectory(ctx context.Context, lineWKT string) ([]entities.Portico, error) {
	return r.inner.FindByTrajectory(ctx, lineWKT)
}

func (r *CachedPorticoRepository) FindViaCrossingsByTrajectory(ctx context.Context, porticoID, lineWKT string) ([]entities.ViaCrossing, error) {
	return r.inner.FindViaCrossingsByTrajectory(ctx, porticoID, lineWKT)
}

func (r *CachedPorticoRepository) Update(ctx context.Context, portico *entities.Portico) (*entities.Portico, error) {
	updated, err := r.inner.Update(ctx, portico)
	if err != nil {
		return nil, err
	}
	r.invalidateAll()
	return clonePortico(updated), nil
}

func (r *CachedPorticoRepository) Delete(ctx context.Context, id string) error {
	if err := r.inner.Delete(ctx, id); err != nil {
		return err
	}
	r.invalidateAll()
	return nil
}

func (r *CachedPorticoRepository) invalidateAll() {
	r.byID.clear()
	r.byKey.clear()
}

func clonePorticos(items []entities.Portico) []entities.Portico {
	out := make([]entities.Portico, 0, len(items))
	for i := range items {
		out = append(out, *clonePortico(&items[i]))
	}
	return out
}

func clonePortico(in *entities.Portico) *entities.Portico {
	if in == nil {
		return nil
	}
	out := *in
	if in.Bearing != nil {
		v := *in.Bearing
		out.Bearing = &v
	}
	if in.DetectionRadiusMeters != nil {
		v := *in.DetectionRadiusMeters
		out.DetectionRadiusMeters = &v
	}
	if in.EntryRadiusMeters != nil {
		v := *in.EntryRadiusMeters
		out.EntryRadiusMeters = &v
	}
	if in.ExitRadiusMeters != nil {
		v := *in.ExitRadiusMeters
		out.ExitRadiusMeters = &v
	}
	if in.EntryLatitude != nil {
		v := *in.EntryLatitude
		out.EntryLatitude = &v
	}
	if in.EntryLongitude != nil {
		v := *in.EntryLongitude
		out.EntryLongitude = &v
	}
	if in.ExitLatitude != nil {
		v := *in.ExitLatitude
		out.ExitLatitude = &v
	}
	if in.ExitLongitude != nil {
		v := *in.ExitLongitude
		out.ExitLongitude = &v
	}
	if in.MaxCrossingSeconds != nil {
		v := *in.MaxCrossingSeconds
		out.MaxCrossingSeconds = &v
	}
	out.Tipo = in.Tipo
	out.Direccion = in.Direccion
	out.VelocidadMaxima = in.VelocidadMaxima
	out.ZonaDeteccionWKT = in.ZonaDeteccionWKT
	out.VehicleTypes = append([]string(nil), in.VehicleTypes...)
	out.IsActive = in.IsActive
	out.Vias = make([]entities.Via, 0, len(in.Vias))
	for _, v := range in.Vias {
		out.Vias = append(out.Vias, v)
	}
	out.Tarifas = make([]entities.Tarifa, 0, len(in.Tarifas))
	for _, t := range in.Tarifas {
		tarifa := t
		tarifa.Horarios = append([]entities.TarifaHorario(nil), t.Horarios...)
		out.Tarifas = append(out.Tarifas, tarifa)
	}
	return &out
}
