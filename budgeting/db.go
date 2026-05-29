package budgeting

import "encore.dev/storage/sqldb"

var db = sqldb.NewDatabase("budgeting", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
