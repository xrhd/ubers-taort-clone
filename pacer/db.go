package pacer

import "encore.dev/storage/sqldb"

var db = sqldb.NewDatabase("pacer", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
