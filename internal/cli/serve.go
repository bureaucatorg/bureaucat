package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"bereaucat/internal/server"

	"github.com/urfave/cli/v3"
)

func ServeCommand() *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "Start the HTTP server",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "api-port",
				Value: 1323,
				Usage: "Port for the API server",
			},
			&cli.StringFlag{
				Name:  "api-host",
				Value: "0.0.0.0",
				Usage: "Host for the API server",
			},
			&cli.BoolFlag{
				Name:  "dev",
				Value: false,
				Usage: "Enable development mode (starts Nuxt dev server)",
			},
			&cli.StringFlag{
				Name:    "database-url",
				Sources: cli.EnvVars("DATABASE_URL"),
				Usage:   "PostgreSQL database connection URL",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			host := cmd.String("api-host")
			port := cmd.Int("api-port")
			dev := cmd.Bool("dev")
			dbURL := cmd.String("database-url")

			addr := fmt.Sprintf("%s:%d", host, port)

			var nuxtCmd *exec.Cmd
			if dev {
				// Start Nuxt dev server
				nuxtCmd = exec.Command("bun", "run", "dev")
				nuxtCmd.Dir = "web"
				nuxtCmd.Stdout = os.Stdout
				nuxtCmd.Stderr = os.Stderr

				if err := nuxtCmd.Start(); err != nil {
					return fmt.Errorf("failed to start Nuxt dev server: %w", err)
				}

				fmt.Println("Started Nuxt dev server")
			}

			// Start the Echo server
			srv, err := server.New(dev, dbURL)
			if err != nil {
				return fmt.Errorf("failed to create server: %w", err)
			}

			// Handle graceful shutdown
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigChan
				fmt.Println("\nShutting down...")

				if nuxtCmd != nil && nuxtCmd.Process != nil {
					nuxtCmd.Process.Signal(syscall.SIGTERM)
					nuxtCmd.Wait()
				}

				srv.Close()
				os.Exit(0)
			}()

			return srv.Start(addr)
		},
	}
}
