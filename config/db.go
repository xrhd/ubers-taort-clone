package config

import "encore.dev/storage/sqldb"

var db = sqldb.NewDatabase("config", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
