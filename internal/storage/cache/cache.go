package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/ARM-auto-place/access/internal/config"
	"github.com/ARM-auto-place/access/internal/models"
	commonErrors "github.com/ARM-auto-place/access/internal/models/errors"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type RoleCache struct {
	mu               sync.Mutex
	logger           zerolog.Logger
	cfg              config.CacheConfig
	active, building *ristretto.Cache[string, string]
	dirty            map[uuid.UUID]struct{}
	snapshotExpires  time.Time
	generation       uint64
	closed           bool
}

func NewRoleCache(logger zerolog.Logger, cfg config.CacheConfig) (*RoleCache, error) {
	if cfg.TTL <= 0 {
		return nil, fmt.Errorf("CACHE_TTL must be positive")
	}
	if cfg.RefreshInterval <= 0 {
		return nil, fmt.Errorf("CACHE_REFRESH_INTERVAL must be positive")
	}
	if cfg.RefreshTimeout <= 0 {
		return nil, fmt.Errorf("CACHE_REFRESH_TIMEOUT must be positive")
	}
	if cfg.RefreshBatchSize <= 0 {
		return nil, fmt.Errorf("CACHE_REFRESH_BATCH_SIZE must be positive")
	}
	if cfg.RefreshMaxDirtyUsers <= 0 {
		return nil, fmt.Errorf("CACHE_REFRESH_MAX_DIRTY_USERS must be positive")
	}
	c := &RoleCache{logger: logger, cfg: cfg}
	var err error
	c.active, err = c.newCache()
	if err != nil {
		return nil, err
	}
	return c, nil
}
func (c *RoleCache) newCache() (*ristretto.Cache[string, string], error) {
	return ristretto.NewCache(&ristretto.Config[string, string]{NumCounters: c.cfg.NumCounters, MaxCost: c.cfg.MaxCost, BufferItems: c.cfg.BufferItems, IgnoreInternalCost: false})
}
func (c *RoleCache) GetRole(id uuid.UUID) (string, bool, uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return "", false, c.generation
	}
	role, ok := c.active.Get(id.String())
	return role, ok, c.generation
}
func (c *RoleCache) put(target *ristretto.Cache[string, string], id uuid.UUID, role string, ttl time.Duration) {
	key := id.String()
	target.Del(key)
	target.Wait()
	if ttl > 0 {
		if !target.SetWithTTL(key, role, roleCost(role), ttl) {
			c.logger.Debug().Msg("role cache write dropped")
		}
		target.Wait()
	}
}

func (c *RoleCache) StoreRead(id uuid.UUID, role string, generation uint64, expires time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed && generation == c.generation {
		c.put(c.active, id, role, time.Until(expires))
	}
}
func (c *RoleCache) SetRole(id uuid.UUID, role string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.put(c.active, id, role, c.cfg.TTL)
	if c.building != nil {
		if _, ok := c.dirty[id]; !ok && len(c.dirty) >= c.cfg.RefreshMaxDirtyUsers {
			c.abort()
			return
		}
		c.dirty[id] = struct{}{}
		c.put(c.building, id, role, c.cfg.TTL)
	}
}
// здесь считаем вес записи, пока не особо разобрался как это работает
func roleCost(role string) int64 {
	return int64(len(role)) + 36 + 32 
}

func (c *RoleCache) DeleteRole(id uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	key := id.String()
	c.active.Del(key)
	c.active.Wait()
	if c.building != nil {
		if _, ok := c.dirty[id]; !ok && len(c.dirty) >= c.cfg.RefreshMaxDirtyUsers {
			c.abort()
			return
		}
		c.dirty[id] = struct{}{}
		c.building.Del(key)
		c.building.Wait()
	}
}

func (c *RoleCache) BeginRefresh() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return fmt.Errorf("role cache closed")
	}
	if c.building != nil {
		return fmt.Errorf("refresh already running")
	}
	next, err := c.newCache()
	if err != nil {
		return err
	}
	c.building = next
	c.dirty = make(map[uuid.UUID]struct{})
	c.snapshotExpires = time.Now().Add(c.cfg.TTL)
	return nil
}
func (c *RoleCache) AddBatch(batch []models.UserRole) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.building == nil {
		return commonErrors.ErrRefreshAborted
	}
	ttl := time.Until(c.snapshotExpires)
	if ttl <= 0 {
		return fmt.Errorf("role snapshot expired before completion")
	}
	for _, item := range batch {
		if _, dirty := c.dirty[item.UserID]; !dirty {
			c.building.SetWithTTL(item.UserID.String(), item.Role, roleCost(item.Role), ttl)
		}
	}
	c.building.Wait()
	return nil
}
func (c *RoleCache) Publish() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.building == nil {
		return commonErrors.ErrRefreshAborted
	}
	if time.Now().After(c.snapshotExpires) {
		return fmt.Errorf("role snapshot expired")
	}
	c.building.Wait()
	old := c.active
	c.active = c.building
	c.building = nil
	c.dirty = nil
	c.generation++
	old.Close()
	return nil
}
func (c *RoleCache) abort() {
	if c.building != nil {
		c.building.Close()
		c.building = nil
		c.dirty = nil
	}
}
func (c *RoleCache) AbortRefresh() { 
	c.mu.Lock(); defer c.mu.Unlock(); c.abort() 
}

func (c *RoleCache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.closed = true
		c.abort()
		c.active.Close()
	}
}
