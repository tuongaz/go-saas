package app

import (
	_ "embed"
)

//go:embed store/persistence/schemas/sqlite.sql
var SqliteSchema string

//go:embed store/persistence/schemas/postgres.sql
var PostgresSchema string

//go:embed store/persistence/schemas/mysql.sql
var MysqlSchema string
