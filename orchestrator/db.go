package orchestrator

import "encore.dev/storage/sqldb"

var db = sqldb.NewDatabase("orchestrator", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
