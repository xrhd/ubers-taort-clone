package domain

import "encore.dev/storage/sqldb"

var db = sqldb.NewDatabase("domain", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
