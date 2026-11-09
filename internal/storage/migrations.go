package storage

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	// Драйвер "file" необходим для чтения файлов миграций из локальной папки
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations выполняет миграции базы данных PostgreSQL
func RunMigrations(db *sql.DB) error {

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://internal/migrations", "postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
