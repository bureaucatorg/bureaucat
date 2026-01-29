package main

import (
	"context"
	"log"
	"os"

	"bereaucat/internal/cli"

	ucli "github.com/urfave/cli/v3"
)

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

func main() {
	app := &ucli.Command{
		Name:  "bureaucat",
		Usage: "A CLI tool for managing the Bureaucat application",
		Commands: []*ucli.Command{
			cli.ServeCommand(),
			cli.MigrateCommand(),
			cli.NuxtCommand(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
