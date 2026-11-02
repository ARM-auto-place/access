package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ARM-auto-place/access/internal/models"
	commonErrors "github.com/ARM-auto-place/access/internal/models/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"
)

var defaultTimeout = 5 * time.Second

type storage struct {
	connectManager ConnectManager
	logger         zerolog.Logger
}

func NewAccessStorage(logger zerolog.Logger, connectManager ConnectManager) AccessStorage {
	return &storage{
		connectManager: connectManager,
		logger:         logger,
	}
}

func (s *storage) SetRole(ctx context.Context, userID uuid.UUID, role string) (err error) {
	masterConn := s.connectManager.GetMasterConn()
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	_, err = masterConn.Exec(ctx, sqlUpsertRole, userID, role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.TableName == "users_role" {
			return commonErrors.ErrUserNotFound
		}
		return fmt.Errorf("SetRole: postgresql: %w", err)
	}
	return nil
}

func (s *storage) GetRole(ctx context.Context, userID uuid.UUID) (string, error) {
	masterConn := s.connectManager.GetMasterConn()
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var role string
	err := masterConn.QueryRow(ctx, sqlSelectRole, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", commonErrors.ErrRoleNotFound
		}
		return "", fmt.Errorf("GetRole: postgresql: %w", err)
	}
	return role, nil
}

func (s *storage) WalkRoles(ctx context.Context, limit int, consume func([]models.UserRole) error) error {
	if limit <= 0 {
		return fmt.Errorf("WalkRoles: limit must be positive")
	}

	masterConn := s.connectManager.GetMasterConn()
	tx, err := masterConn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return fmt.Errorf("WalkRoles: begin: %w", err)
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		_ = tx.Rollback(cleanupCtx)
	}()

	var afterID *uuid.UUID
	for {
		batch, err := s.getRolesBatch(ctx, tx, afterID, limit)
		if err != nil {
			return err
		}
		if len(batch) == 0 {
			break
		}
		if err = consume(batch); err != nil {
			return err
		}
		lastID := batch[len(batch)-1].UserID
		afterID = &lastID
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("WalkRoles: commit: %w", err)
	}
	return nil
}

func (s *storage) getRolesBatch(ctx context.Context, tx pgx.Tx, afterID *uuid.UUID, limit int) ([]models.UserRole, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	rows, err := tx.Query(ctx, sqlSelectRoles, afterID, limit)
	if err != nil {
		return nil, fmt.Errorf("GetRolesBatch: query: %w", err)
	}
	defer rows.Close()

	batch := make([]models.UserRole, 0, limit)
	for rows.Next() {
		var role models.UserRole
		if err = rows.Scan(&role.UserID, &role.Role); err != nil {
			return nil, fmt.Errorf("GetRolesBatch: scan: %w", err)
		}
		batch = append(batch, role)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetRolesBatch: rows: %w", err)
	}
	return batch, nil
}
