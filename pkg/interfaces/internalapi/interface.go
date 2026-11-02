package internalapi

import (
	"context"

	"github.com/google/uuid"
)

// @tg version=0.0.1
// @tg backend=access-api
// @tg title=`access API`
// @tg servers=
//
//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/internalapi --outSwagger ../../../swaggers/internalapi/swagger.yaml
//go:generate tg client --services . --outPath ../../client/transport -go
// InternalAPI
// @tg jsonRPC-server metrics log
// @tg uuidPackage=github.com/google/uuid
type InternalAPI interface {
	// SetRole ...
	// @tg summary=`Назначение роли пользователя`
	// @tg desc=`Назначает или обновляет роль пользователя. Ошибки JSON-RPC: -32602 — некорректные параметры; 1002 — пользователь не найден; -32603 — внутренняя ошибка.`
	SetRole(ctx context.Context, userID uuid.UUID, role string) (err error)

	// GetRole ...
	// @tg summary=`Получение роли пользователя`
	// @tg desc=`Возвращает текущую роль пользователя. Ошибки JSON-RPC: -32602 — некорректные параметры; 1001 — роль не назначена; -32603 — внутренняя ошибка.`
	GetRole(ctx context.Context, userID uuid.UUID) (role string, err error)
}
