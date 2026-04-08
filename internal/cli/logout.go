package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func LogoutCommand() *cli.Command {
	return &cli.Command{
		Name:  "logout",
		Usage: "Clear locally stored Bureaucat credentials",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg, err := LoadConfig()
			if err != nil {
				return err
			}
			if cfg.URL == "" && cfg.Token == "" {
				if IsHuman(cmd) {
					fmt.Fprintln(os.Stdout, "Already logged out.")
					return nil
				}
				return PrintJSON(map[string]any{"logged_out": true, "message": "already logged out"})
			}

			if err := ClearConfig(); err != nil {
				return err
			}
			if IsHuman(cmd) {
				fmt.Fprintln(os.Stdout, "Logged out. Stored credentials cleared.")
				return nil
			}
			return PrintJSON(map[string]any{"logged_out": true, "message": "stored credentials cleared"})
		},
	}
}
