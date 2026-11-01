// Package externalapi defines the access service contract.
// @tg version=0.0.1
// @tg backend=access-api
// @tg title=`access API`
package externalapi

import (
	"context"

	"github.com/google/uuid"
)

// JSON-RPC Swagger uses the detailed error models; tg v2.3.95 does not read
// HTTP error annotations in JSON-RPC mode, so it is maintained separately.
//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/externalapi

// ExternalAPI
// @tg jsonRPC-server metrics log
// @tg http-prefix=/api/access/v1
// @tg uuidPackage=github.com/google/uuid
type ExternalAPI interface {
	// SetRole ...
	// @tg summary=`Назначение роли пользователя`
	// @tg desc=`Назначает или обновляет роль пользователя. Ошибки JSON-RPC: -32602 — некорректные параметры; 1002 — пользователь не найден; -32603 — внутренняя ошибка.`
	SetRole(ctx context.Context, userID uuid.UUID, role string) (err error)
	// GetRole ...
	// @tg summary=`Получение роли пользователя`
	// @tg desc=`Возвращает текущую роль пользователя. Ошибки JSON-RPC: -32602 — некорректные параметры; 1001 — роль не назначена; -32603 — внутренняя ошибка.`
	GetRole(ctx context.Context, userID uuid.UUID) (role string, err error)
}
