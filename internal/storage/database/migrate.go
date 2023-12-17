package database

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func (db *Database) migrate() error {
	if !db.doMigration {
		return nil
	}
	db.log.Debug("starting migration")
	change, err := runMigrations(db.dsn)
	if err != nil {
		return err
	}
	if change {
		db.log.Info("migration is complete")
	} else {
		db.log.Debug("db is up to date")
	}
	return nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) (bool, error) {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return false, fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return false, fmt.Errorf("failed to get a new migrate instance: %w", err)
	}

	if err = m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return false, fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
		return false, nil
	}
	return true, nil
}
