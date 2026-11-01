package models

type GetRoleResult struct {
	// @tg desc=`Назначенная пользователю роль`
	// @tg example=`manager`
	Role string `json:"role"`
}

type RespGetRole200 struct {
	// @tg example=`2.0`
	JSONRPC string `json:"jsonrpc"`
	// @tg desc=`Идентификатор исходного запроса`
	// @tg example=1
	ID     any           `json:"id"`
	Result GetRoleResult `json:"result"`
}

type RespSetRole200 struct {
	// @tg example=`2.0`
	JSONRPC string `json:"jsonrpc"`
	// @tg desc=`Идентификатор исходного запроса`
	// @tg example=1
	ID any `json:"id"`
	// @tg desc=`Пустой объект при успешном назначении роли`
	Result struct{} `json:"result"`
}