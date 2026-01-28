package main

import (
	"context"
	"log"
	"os"

	"bereaucat/internal/cli"

	ucli "github.com/urfave/cli/v3"
)

func main() {
	app := &ucli.Command{
		Name:  "bureaucat",
		Usage: "A CLI tool for managing the Bureaucat application",
		Commands: []*ucli.Command{
			cli.ServeCommand(),
			cli.MigrateCommand(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
