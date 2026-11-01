package models

import "github.com/google/uuid"

type UserRole struct {
	UserID uuid.UUID
	Role   string
}
