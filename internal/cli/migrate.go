package cli

import (
	"context"
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"bereaucat/internal/database"

	"github.com/urfave/cli/v3"
)

// MigrationsFS is set by the main package to provide embedded migration files
var MigrationsFS fs.FS

// createMigrator creates a migrator - uses filesystem path if it exists, otherwise embedded FS
func createMigrator(dbURL, path string) (*database.Migrator, error) {
	// Try filesystem path first (for dev mode)
	if migrator, err := database.NewMigrator(dbURL, path); err == nil {
		return migrator, nil
	}

	// Fall back to embedded FS (for prod mode)
	if MigrationsFS != nil {
		return database.NewMigratorFromFS(dbURL, MigrationsFS)
	}

	// Neither available
	return nil, fmt.Errorf("no migrations found at path %q and no embedded migrations available", path)
}

func MigrateCommand() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Database migration commands",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "database-url",
				Usage:   "Database connection URL",
				Sources: cli.EnvVars("DATABASE_URL"),
			},
			&cli.StringFlag{
				Name:  "path",
				Value: "migrations",
				Usage: "Path to migration files (ignored when using embedded migrations)",
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "up",
				Usage:     "Run migrations (all or up to specific version)",
				ArgsUsage: "[version]",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					dbURL := cmd.String("database-url")
					if dbURL == "" {
						return fmt.Errorf("database-url is required (use --database-url or DATABASE_URL env var)")
					}

					path := cmd.String("path")
					migrator, err := createMigrator(dbURL, path)
					if err != nil {
						return fmt.Errorf("failed to create migrator: %w", err)
					}
					defer migrator.Close()

					if cmd.Args().Len() > 0 {
						version, err := parseVersion(cmd.Args().First())
						if err != nil {
							return err
						}
						return migrator.MigrateToVersion(version)
					}

					return migrator.Up()
				},
			},
			{
				Name:      "down",
				Usage:     "Revert migrations (all or down to specific version)",
				ArgsUsage: "[version]",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					dbURL := cmd.String("database-url")
					if dbURL == "" {
						return fmt.Errorf("database-url is required (use --database-url or DATABASE_URL env var)")
					}

					path := cmd.String("path")
					migrator, err := createMigrator(dbURL, path)
					if err != nil {
						return fmt.Errorf("failed to create migrator: %w", err)
					}
					defer migrator.Close()

					if cmd.Args().Len() > 0 {
						version, err := parseVersion(cmd.Args().First())
						if err != nil {
							return err
						}
						return migrator.MigrateToVersion(version)
					}

					return migrator.Down()
				},
			},
		},
	}
}

// parseVersion handles both "0002" and "2" formats
func parseVersion(s string) (uint, error) {
	s = strings.TrimLeft(s, "0")
	if s == "" {
		return 0, nil
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid version: %s", s)
	}
	return uint(v), nil
}
