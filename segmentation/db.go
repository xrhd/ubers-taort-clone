package segmentation

import "encore.dev/storage/sqldb"

var db = sqldb.NewDatabase("segmentation", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
