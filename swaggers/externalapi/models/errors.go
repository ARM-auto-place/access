package models

// Описание JSON-RPC-ошибок access. Ошибки передаются в поле error ответа HTTP 200.
// Структуры используются схемами в swaggers/externalapi/swagger.yaml.
type RPCError struct {
	// @tg desc=`Код JSON-RPC. -32700 — Некорректный JSON. Генератор также использует этот код при ошибке декодирования параметров или версии протокола.; -32600 — Некорректный JSON-RPC-запрос; зарезервированный код протокола.; -32601 — Запрошенный JSON-RPC-метод не найден или не соответствует маршруту.; -32602 — Некорректные параметры: отсутствующий или нулевой UUID, пустая роль либо роль из пробелов.; -32603 — Внутренняя ошибка: ошибка PostgreSQL, отмена операции или истечение её таймаута. Технические причины остаются в логах.; 1001 — GetRole: у пользователя отсутствует запись о роли в users_role.; 1002 — SetRole: внешний ключ users_role не позволяет назначить роль несуществующему пользователю.`
	// @tg example=-32603
	Code int `json:"code"`
	// @tg desc=`Публичное сообщение без технических причин ошибки`
	// @tg example=`internal server error`
	Message string `json:"message"`
	// @tg desc=`Необязательные публичные данные. Для ошибок сервисного слоя tg возвращает пустой объект; исходная ошибка и Cause сюда не попадают.`
	Data map[string]any `json:"data,omitempty"`
}

// RPCErrorResponse описывает общую оболочку ошибки.
type RPCErrorResponse struct {
	// @tg example=`2.0`
	JSONRPC string `json:"jsonrpc"`
	// @tg desc=`Идентификатор исходного запроса: число, строка или null`
	// @tg example=1
	ID    any      `json:"id"`
	Error RPCError `json:"error"`
}

// ParseError — Некорректный JSON. Генератор также использует этот код при ошибке декодирования параметров или версии протокола.
type ParseError struct {
	// @tg desc=`Некорректный JSON. Генератор также использует этот код при ошибке декодирования параметров или версии протокола.`
	// @tg example=-32700
	Code int `json:"code"`
	// @tg example=`parse error`
	Message string `json:"message"`
	// @tg desc=`Необязательные публичные данные без исходной ошибки и Cause`
	Data map[string]any `json:"data,omitempty"`
}

// ErrParseError описывает ответ с JSON-RPC-кодом -32700.
type ErrParseError struct {
	// @tg example=`2.0`
	JSONRPC string `json:"jsonrpc"`
	// @tg desc=`Идентификатор исходного запроса: число, строка или null`
	// @tg example=1
	ID    any        `json:"id"`
	Error ParseError `json:"error"`
}

// InvalidRequest — Некорректный JSON-RPC-запрос; зарезервированный код протокола.
type InvalidRequest struct {
	// @tg desc=`Некорректный JSON-RPC-запрос; зарезервированный код протокола.`
	// @tg example=-32600
	Code int `json:"code"`
	// @tg example=`invalid request`
	Message string `json:"message"`
	// @tg desc=`Необязательные публичные данные без исходной ошибки и Cause`
	Data map[string]any `json:"data,omitempty"`
}

// ErrInvalidRequest описывает ответ с JSON-RPC-кодом -32600.
type ErrInvalidRequest struct {
	// @tg example=`2.0`
	JSONRPC string `json:"jsonrpc"`
	// @tg desc=`Идентификатор исходного запроса: число, строка или null`
	// @tg example=1
	ID    any            `json:"id"`
	Error InvalidRequest `json:"error"`
}

// MethodNotFound — Запрошенный JSON-RPC-метод не найден или не соответствует маршруту.
type MethodNotFound struct {
	// @tg desc=`Запрошенный JSON-RPC-метод не найден или не соответствует маршруту.`
	// @tg example=-32601
	Code int `json:"code"`
	// @tg example=`method not found`
	Message string `json:"message"`
	// @tg desc=`Необязательные публичные данные без исходной ошибки и Cause`
	Data map[string]any `json:"data,omitempty"`
}

// ErrMethodNotFound описывает ответ с JSON-RPC-кодом -32601.
type ErrMethodNotFound struct {
	// @tg example=`2.0`
	JSONRPC string `json:"jsonrpc"`
	// @tg desc=`Идентификатор исходного запроса: число, строка или null`
	// @tg example=1
	ID    any            `json:"id"`
	Error MethodNotFound `json:"error"`
}

// InvalidParams — Некорректные параметры: отсутствующий или нулевой UUID, пустая роль либо роль из пробелов.
type InvalidParams struct {
	// @tg desc=`Некорректные параметры: отсутствующий или нулевой UUID, пустая роль либо роль из пробелов.`
	// @tg example=-32602
	Code int `json:"code"`
	// @tg example=`invalid parameters`
	Message string `json:"message"`
	// @tg desc=`Необязательные публичные данные без исходной ошибки и Cause`
	Data map[string]any `json:"data,omitempty"`
}

// ErrInvalidParams описывает ответ с JSON-RPC-кодом -32602.
type ErrInvalidParams struct {
	// @tg example=`2.0`
	JSONRPC string `json:"jsonrpc"`
	// @tg desc=`Идентификатор исходного запроса: число, строка или null`
	// @tg example=1
	ID    any           `json:"id"`
	Error InvalidParams `json:"error"`
}

// InternalError — Внутренняя ошибка: ошибка PostgreSQL, отмена операции или истечение её таймаута. Технические причины остаются в логах.
type InternalError struct {
	// @tg desc=`Внутренняя ошибка: ошибка PostgreSQL, отмена операции или истечение её таймаута. Технические причины остаются в логах.`
	// @tg example=-32603
	Code int `json:"code"`
	// @tg example=`internal server error`
	Message string `json:"message"`
	// @tg desc=`Необязательные публичные данные без исходной ошибки и Cause`
	Data map[string]any `json:"data,omitempty"`
}

// ErrInternalError описывает ответ с JSON-RPC-кодом -32603.
type ErrInternalError struct {
	// @tg example=`2.0`
	JSONRPC string `json:"jsonrpc"`
	// @tg desc=`Идентификатор исходного запроса: число, строка или null`
	// @tg example=1
	ID    any           `json:"id"`
	Error InternalError `json:"error"`
}

// RoleNotFound — GetRole: у пользователя отсутствует запись о роли в users_role.
type RoleNotFound struct {
	// @tg desc=`GetRole: у пользователя отсутствует запись о роли в users_role.`
	// @tg example=1001
	Code int `json:"code"`
	// @tg example=`role not assigned`
	Message string `json:"message"`
	// @tg desc=`Необязательные публичные данные без исходной ошибки и Cause`
	Data map[string]any `json:"data,omitempty"`
}

// ErrRoleNotFound описывает ответ с JSON-RPC-кодом 1001.
type ErrRoleNotFound struct {
	// @tg example=`2.0`
	JSONRPC string `json:"jsonrpc"`
	// @tg desc=`Идентификатор исходного запроса: число, строка или null`
	// @tg example=1
	ID    any          `json:"id"`
	Error RoleNotFound `json:"error"`
}

// UserNotFound — SetRole: внешний ключ users_role не позволяет назначить роль несуществующему пользователю.
type UserNotFound struct {
	// @tg desc=`SetRole: внешний ключ users_role не позволяет назначить роль несуществующему пользователю.`
	// @tg example=1002
	Code int `json:"code"`
	// @tg example=`user not found`
	Message string `json:"message"`
	// @tg desc=`Необязательные публичные данные без исходной ошибки и Cause`
	Data map[string]any `json:"data,omitempty"`
}

// ErrUserNotFound описывает ответ с JSON-RPC-кодом 1002.
type ErrUserNotFound struct {
	// @tg example=`2.0`
	JSONRPC string `json:"jsonrpc"`
	// @tg desc=`Идентификатор исходного запроса: число, строка или null`
	// @tg example=1
	ID    any          `json:"id"`
	Error UserNotFound `json:"error"`
}
