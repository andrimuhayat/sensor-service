package migration

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

func MigrationRubenv(conn *sqlx.DB) {

	migrations := &migrate.FileMigrationSource{
		Dir: "internal/platform/storage/migration",
	}
	migrate.SetTable("claims_migrations")

	var migrateType = migrate.Up
	//if helper.GetEnv("MIGRATE_TYPE") == "down" {
	//	migrateType = migrate.Down
	//}
	n, err := migrate.Exec(conn.DB, "mysql", migrations, migrateType)
	if err != nil {
		fmt.Println("Migration Error occcured:", err)
		return
	}

	fmt.Printf("Applied %d migrations!\n", n)
}
