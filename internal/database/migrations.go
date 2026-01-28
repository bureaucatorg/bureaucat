package database

import (
	"fmt"
	"time"

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

// Up runs all pending migrations one at a time with logging
func (m *Migrator) Up() error {
	startVersion, dirty, _ := m.m.Version()
	if dirty {
		fmt.Printf("Warning: database is in dirty state at version %d\n", startVersion)
	}

	fmt.Printf("Current version: %s\n", formatVersion(startVersion))

	totalStart := time.Now()
	migrationCount := 0

	for {
		version, _, _ := m.m.Version()

		start := time.Now()
		err := m.m.Steps(1)
		duration := time.Since(start)

		if err != nil {
			// Check for "no change" or "file does not exist" (no more migrations)
			if err == migrate.ErrNoChange || isNoMoreMigrations(err) {
				break
			}
			return fmt.Errorf("failed to run migration: %w", err)
		}

		newVersion, _, _ := m.m.Version()
		fmt.Printf("  Applied %s (took %s)\n", formatVersion(newVersion), duration.Round(time.Millisecond))
		migrationCount++

		// Safety check to prevent infinite loop
		if newVersion == version {
			break
		}
	}

	totalDuration := time.Since(totalStart)
	endVersion, _, _ := m.m.Version()

	if migrationCount == 0 {
		fmt.Printf("No migrations to apply (at version %s)\n", formatVersion(endVersion))
		return nil
	}

	fmt.Printf("Applied %d migration(s): %s -> %s (total %s)\n",
		migrationCount, formatVersion(startVersion), formatVersion(endVersion), totalDuration.Round(time.Millisecond))
	return nil
}

// Down reverts all migrations one at a time with logging
func (m *Migrator) Down() error {
	startVersion, dirty, _ := m.m.Version()
	if dirty {
		fmt.Printf("Warning: database is in dirty state at version %d\n", startVersion)
	}

	fmt.Printf("Current version: %s\n", formatVersion(startVersion))

	totalStart := time.Now()
	migrationCount := 0

	for {
		version, _, _ := m.m.Version()

		start := time.Now()
		err := m.m.Steps(-1)
		duration := time.Since(start)

		if err != nil {
			// Check for "no change" or "file does not exist" (no more migrations)
			if err == migrate.ErrNoChange || isNoMoreMigrations(err) {
				break
			}
			return fmt.Errorf("failed to revert migration: %w", err)
		}

		fmt.Printf("  Reverted %s (took %s)\n", formatVersion(version), duration.Round(time.Millisecond))
		migrationCount++

		newVersion, _, _ := m.m.Version()
		// Safety check to prevent infinite loop
		if newVersion == version {
			break
		}
	}

	totalDuration := time.Since(totalStart)

	if migrationCount == 0 {
		fmt.Println("No migrations to revert (database is clean)")
		return nil
	}

	fmt.Printf("Reverted %d migration(s): %s -> none (total %s)\n",
		migrationCount, formatVersion(startVersion), totalDuration.Round(time.Millisecond))
	return nil
}

// MigrateToVersion migrates to a specific version (works for both up and down)
func (m *Migrator) MigrateToVersion(version uint) error {
	startVersion, dirty, _ := m.m.Version()
	if dirty {
		fmt.Printf("Warning: database is in dirty state at version %d\n", startVersion)
	}

	fmt.Printf("Current version: %s\n", formatVersion(startVersion))
	fmt.Printf("Target version: %s\n", formatVersion(version))

	totalStart := time.Now()
	migrationCount := 0
	step := 1
	if version < startVersion {
		step = -1
	}

	for {
		currentVersion, _, _ := m.m.Version()

		// Check if we've reached the target
		if (step > 0 && currentVersion >= version) || (step < 0 && currentVersion <= version) {
			break
		}

		start := time.Now()
		err := m.m.Steps(step)
		duration := time.Since(start)

		if err == migrate.ErrNoChange {
			break
		}

		if err != nil {
			return fmt.Errorf("failed to migrate: %w", err)
		}

		newVersion, _, _ := m.m.Version()
		if step > 0 {
			fmt.Printf("  Applied %s (took %s)\n", formatVersion(newVersion), duration.Round(time.Millisecond))
		} else {
			fmt.Printf("  Reverted %s (took %s)\n", formatVersion(currentVersion), duration.Round(time.Millisecond))
		}
		migrationCount++

		// Safety check
		if newVersion == currentVersion {
			break
		}
	}

	totalDuration := time.Since(totalStart)
	endVersion, _, _ := m.m.Version()

	if migrationCount == 0 {
		fmt.Printf("Already at version %s\n", formatVersion(version))
		return nil
	}

	fmt.Printf("Migrated %d step(s): %s -> %s (total %s)\n",
		migrationCount, formatVersion(startVersion), formatVersion(endVersion), totalDuration.Round(time.Millisecond))
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

// formatVersion formats a version number, handling the case of no migrations applied
func formatVersion(v uint) string {
	if v == 0 {
		return "none"
	}
	return fmt.Sprintf("%04d", v)
}

// isNoMoreMigrations checks if the error indicates there are no more migrations to apply
func isNoMoreMigrations(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return errStr == "file does not exist" || errStr == "first : file does not exist"
}
