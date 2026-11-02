package postgres

import (
	"context"

	"github.com/ARM-auto-place/access/internal/models"
	"github.com/google/uuid"
)

type AccessStorage interface {
	SetRole(context.Context, uuid.UUID, string) error
	GetRole(context.Context, uuid.UUID) (string, error)
	WalkRoles(context.Context, int, func([]models.UserRole) error) error
}
