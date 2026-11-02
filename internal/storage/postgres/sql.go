package postgres

import (
	_ "embed"
)

//go:embed sql/select_role.sql
var sqlSelectRole string

//go:embed sql/select_roles.sql
var sqlSelectRoles string

//go:embed sql/upsert_role.sql
var sqlUpsertRole string
