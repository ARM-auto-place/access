package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/ARM-auto-place/access/internal/config"
	commonErrors "github.com/ARM-auto-place/access/internal/models/errors"
	"github.com/ARM-auto-place/access/internal/storage/cache"
	"github.com/ARM-auto-place/access/internal/storage/postgres"
	interfaces "github.com/ARM-auto-place/access/pkg/interfaces/externalapi"
	customErrors "github.com/ARM-auto-place/access/pkg/models/errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type accessService struct {
	storage   postgres.AccessStorage
	roleCache *cache.RoleCache
	logger    zerolog.Logger
	cfg       config.CacheConfig
	locks     [256]chan struct{}
	refreshMu sync.Mutex
	runOnce   sync.Once
}

var _ interfaces.ExternalAPI = (*accessService)(nil)

func NewAccessService(logger zerolog.Logger, storage postgres.AccessStorage, roleCache *cache.RoleCache, cfg config.CacheConfig) *accessService {
	s := &accessService{storage: storage, roleCache: roleCache, logger: logger, cfg: cfg}
	for i := range s.locks {
		s.locks[i] = make(chan struct{}, 1)
	}
	return s
}
func (s *accessService) SetRole(ctx context.Context, userID uuid.UUID, role string) error {
	if userID == uuid.Nil || strings.TrimSpace(role) == "" {
		return customErrors.InvalidParams()
	}
	lock := s.locks[userID[0]]
	select {
	case lock <- struct{}{}:
	case <-ctx.Done():
		return customErrors.InternalServerError().WithWrapped(ctx.Err(), "SetRole canceled")
	}
	defer func() { <-lock }()
	if err := ctx.Err(); err != nil {
		return customErrors.InternalServerError().WithWrapped(err, "SetRole canceled")
	}
	if err := s.storage.SetRole(ctx, userID, role); err != nil {
		s.roleCache.DeleteRole(userID)
		if errors.Is(err, commonErrors.ErrUserNotFound) {
			return customErrors.ErrUserNotFound()
		}
		return customErrors.InternalServerError().WithWrapped(err, "failed to save role")
	}
	s.roleCache.SetRole(userID, role)
	return nil
}
func (s *accessService) GetRole(ctx context.Context, userID uuid.UUID) (string, error) {
	if userID == uuid.Nil {
		return "", customErrors.InvalidParams()
	}
	lock := s.locks[userID[0]]
	select {
	case lock <- struct{}{}:
	case <-ctx.Done():
		return "", customErrors.InternalServerError().WithWrapped(ctx.Err(), "GetRole canceled")
	}
	defer func() { <-lock }()
	if err := ctx.Err(); err != nil {
		return "", customErrors.InternalServerError().WithWrapped(err, "GetRole canceled")
	}
	role, hit, generation := s.roleCache.GetRole(userID)
	label := "miss"
	if hit {
		label = "hit"
	}
	config.Metrics().CacheRequests.WithLabelValues(label).Inc()
	if hit {
		return role, nil
	}
	expires := time.Now().Add(s.cfg.TTL)
	role, err := s.storage.GetRole(ctx, userID)
	if err != nil {
		if errors.Is(err, commonErrors.ErrRoleNotFound) {
			return "", customErrors.ErrRoleNotFound()
		}
		return "", customErrors.InternalServerError().WithWrapped(err, "failed to read role")
	}
	if strings.TrimSpace(role) == "" {
		return "", customErrors.InternalServerError()
	}
	s.roleCache.StoreRead(userID, role, generation, expires)
	return role, nil
}
