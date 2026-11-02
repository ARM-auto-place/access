package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ARM-auto-place/access/internal/config"
	"github.com/ARM-auto-place/access/internal/models"
)

func (s *accessService) RunCacheRefresh(ctx context.Context) {
	s.runOnce.Do(func() {
		ticker := time.NewTicker(s.cfg.RefreshInterval)
		defer ticker.Stop()
		for {
			if ctx.Err() != nil {
				return
			}
			s.refresh(ctx)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	})
}
func (s *accessService) refresh(parent context.Context) {
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()
	ctx, cancel := context.WithTimeout(parent, s.cfg.RefreshTimeout)
	defer cancel()
	started := time.Now()
	var err error
	defer func() {
		m := config.Metrics()
		m.RefreshDuration.Observe(time.Since(started).Seconds())
		m.RefreshRuns.WithLabelValues(strconv.FormatBool(err == nil)).Inc()
		if err != nil {
			s.logger.Error().Err(err).Msg("role cache refresh failed")
		} else {
			m.RefreshLastSuccess.SetToCurrentTime()
		}
	}()
	if err = s.roleCache.BeginRefresh(); err != nil {
		return
	}
	defer s.roleCache.AbortRefresh()
	err = s.storage.WalkRoles(ctx, s.cfg.RefreshBatchSize, func(batch []models.UserRole) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		for _, item := range batch {
			if strings.TrimSpace(item.Role) == "" {
				return fmt.Errorf("invalid role in PostgreSQL snapshot")
			}
		}
		return s.roleCache.AddBatch(batch)
	})
	if err != nil {
		return
	}
	if err = ctx.Err(); err != nil {
		return
	}
	err = s.roleCache.Publish()
}
