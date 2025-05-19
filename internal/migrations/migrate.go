package migrations

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/thangit93/echo-base/config"
)

func RunMigration() {
	dbURL := fmt.Sprintf("mysql://%s", config.MYSQL_DSN)

	m, err := migrate.New(
		"file://database/migrations",
		dbURL,
	)
	if err != nil {
		log.Fatalf("❌ Failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Println("✅ Migration applied successfully")
}
