package main

import (
	"context"
	"log"
	"os"

	"bereaucat/internal/buildinfo"
	"bereaucat/internal/cli"

	ucli "github.com/urfave/cli/v3"
)

// Version is set at build time via ldflags.
var Version = "dev"

func init() {
	// Set the embedded dist filesystem for production mode
	distFS, err := GetDistFS()
	if err != nil {
		log.Printf("Warning: could not load embedded dist: %v", err)
	}
	cli.DistFS = distFS

	// Set the embedded migrations filesystem for production mode
	migrationsFS, err := GetMigrationsFS()
	if err != nil {
		log.Printf("Warning: could not load embedded migrations: %v", err)
	}
	cli.MigrationsFS = migrationsFS
}

//	@title						Bureaucat API
//	@version					1.0
//	@description				Bureaucat - Approval workflow application API
//	@host						localhost:1341
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Enter your bearer token in the format: Bearer {token}
func main() {
	buildinfo.Version = Version

	app := &ucli.Command{
		Name:    "bureaucat",
		Usage:   "A CLI tool for managing the Bureaucat application",
		Version: Version,
		Commands: []*ucli.Command{
			cli.ServeCommand(),
			cli.MigrateCommand(),
			cli.NuxtCommand(),
			cli.PopulateDBCommand(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
