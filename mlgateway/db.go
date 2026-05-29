package mlgateway

import "encore.dev/storage/sqldb"

var db = sqldb.NewDatabase("mlgateway", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
