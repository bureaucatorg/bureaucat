package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"bereaucat/internal/database"

	"github.com/urfave/cli/v3"
)

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
				Usage: "Path to migration files",
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
					migrator, err := database.NewMigrator(dbURL, path)
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
					migrator, err := database.NewMigrator(dbURL, path)
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
