package app

import (
	_ "embed"
)

//go:embed service/auth/store/schemas/sqlite.sql
var SqliteSchema string

//go:embed service/auth/store/schemas/postgres.sql
var PostgresSchema string

//go:embed service/auth/store/schemas/mysql.sql
var MySQLSchema string
