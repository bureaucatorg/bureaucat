package database

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migrator wraps golang-migrate for database migrations
type Migrator struct {
	m *migrate.Migrate
}

// NewMigrator creates a new Migrator instance
func NewMigrator(dbURL, migrationsPath string) (*Migrator, error) {
	sourceURL := fmt.Sprintf("file://%s", migrationsPath)

	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{m: m}, nil
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	err := m.m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		fmt.Println("No migrations to apply")
		return nil
	}

	fmt.Println("Migrations applied successfully")
	return nil
}

// Down reverts all migrations
func (m *Migrator) Down() error {
	err := m.m.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to revert migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		fmt.Println("No migrations to revert")
		return nil
	}

	fmt.Println("Migrations reverted successfully")
	return nil
}

// MigrateToVersion migrates to a specific version (works for both up and down)
func (m *Migrator) MigrateToVersion(version uint) error {
	err := m.m.Migrate(version)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}

	if err == migrate.ErrNoChange {
		fmt.Printf("Already at version %d\n", version)
		return nil
	}

	fmt.Printf("Migrated to version %d\n", version)
	return nil
}

// Version returns the current migration version
func (m *Migrator) Version() (uint, bool, error) {
	return m.m.Version()
}

// Close closes the migrator
func (m *Migrator) Close() error {
	sourceErr, dbErr := m.m.Close()
	if sourceErr != nil {
		return sourceErr
	}
	return dbErr
}
