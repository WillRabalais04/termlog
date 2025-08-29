package db

import (
	_ "embed"
)

//go:embed migrations/sqlite/000001_create_logs_table.up.sql
var SqliteSchema string

//go:embed migrations/postgres/000001_create_logs_table.up.sql
var PostgresSchema string
