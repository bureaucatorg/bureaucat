package cli

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"bereaucat/internal/server"

	"github.com/urfave/cli/v3"
)

// PidFile is the location of the serve process PID file
var PidFile = filepath.Join(os.TempDir(), "bureaucat-serve.pid")

// NuxtPidFile is the location of the Nuxt process PID file
var NuxtPidFile = filepath.Join(os.TempDir(), "bureaucat-nuxt.pid")

// DistFS is set by the main package to provide embedded static files
var DistFS fs.FS

func ServeCommand() *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "Start the HTTP server",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "api-port",
				Sources: cli.EnvVars("API_PORT", "PORT"),
				Value:   1341,
				Usage:   "Port for the API server",
			},
			&cli.StringFlag{
				Name:    "api-host",
				Sources: cli.EnvVars("API_HOST"),
				Value:   "0.0.0.0",
				Usage:   "Host for the API server",
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
			&cli.StringFlag{
				Name:     "jwt-secret",
				Sources:  cli.EnvVars("JWT_SECRET"),
				Usage:    "Secret key for JWT signing (32+ characters recommended)",
				Required: true,
			},
			&cli.IntFlag{
				Name:    "access-token-expiry-mins",
				Sources: cli.EnvVars("ACCESS_TOKEN_EXPIRY_MINS"),
				Value:   5,
				Usage:   "Access token expiry time in minutes",
			},
			&cli.IntFlag{
				Name:    "refresh-token-expiry-days",
				Sources: cli.EnvVars("REFRESH_TOKEN_EXPIRY_DAYS"),
				Value:   7,
				Usage:   "Refresh token expiry time in days",
			},
			&cli.BoolFlag{
				Name:  "migrate",
				Value: false,
				Usage: "Run pending database migrations before starting the server (equivalent to `./bureaucat migrate up`)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			host := cmd.String("api-host")
			port := cmd.Int("api-port")
			dev := cmd.Bool("dev")
			dbURL := cmd.String("database-url")
			jwtSecret := cmd.String("jwt-secret")
			accessTokenExpiryMins := int(cmd.Int("access-token-expiry-mins"))
			refreshTokenExpiryDays := int(cmd.Int("refresh-token-expiry-days"))
			runMigrations := cmd.Bool("migrate")

			addr := fmt.Sprintf("%s:%d", host, port)

			// Run pending migrations before opening the main server pool, so a
			// failed migration aborts startup cleanly with a visible error
			// instead of leaving the server running against an out-of-date schema.
			// Reuses the same helper that backs `./bureaucat migrate up`.
			if runMigrations {
				if dbURL == "" {
					return fmt.Errorf("--migrate requires --database-url (or DATABASE_URL)")
				}
				fmt.Println("Running database migrations...")
				migrator, err := createMigrator(dbURL, "migrations")
				if err != nil {
					return fmt.Errorf("failed to create migrator: %w", err)
				}
				if err := migrator.Up(); err != nil {
					migrator.Close()
					return fmt.Errorf("migration failed: %w", err)
				}
				migrator.Close()
				fmt.Println("Migrations complete.")
			}

			// Write serve PID file
			if err := os.WriteFile(PidFile, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
				fmt.Printf("Warning: could not write PID file: %v\n", err)
			}
			defer os.Remove(PidFile)

			var nuxtCmd *exec.Cmd
			var nuxtMu = make(chan struct{}, 1)
			nuxtMu <- struct{}{} // Initialize mutex

			startNuxt := func() error {
				<-nuxtMu                                // Lock
				defer func() { nuxtMu <- struct{}{} }() // Unlock

				// Stop existing Nuxt if running
				if nuxtCmd != nil && nuxtCmd.Process != nil {
					fmt.Println("Stopping existing Nuxt dev server...")
					nuxtCmd.Process.Signal(syscall.SIGTERM)
					nuxtCmd.Wait()
				}

				// Start new Nuxt dev server
				nuxtCmd = exec.Command("bun", "run", "dev", "--", "--port", "3041")
				nuxtCmd.Dir = "web"
				nuxtCmd.Stdout = os.Stdout
				nuxtCmd.Stderr = os.Stderr

				if err := nuxtCmd.Start(); err != nil {
					return fmt.Errorf("failed to start Nuxt dev server: %w", err)
				}

				// Write Nuxt PID file
				if err := os.WriteFile(NuxtPidFile, []byte(strconv.Itoa(nuxtCmd.Process.Pid)), 0644); err != nil {
					fmt.Printf("Warning: could not write Nuxt PID file: %v\n", err)
				}

				fmt.Println("Started Nuxt dev server")
				return nil
			}

			if dev {
				if err := startNuxt(); err != nil {
					return err
				}
			}

			// Start the Echo server
			authConfig := server.AuthConfig{
				JWTSecret:              jwtSecret,
				AccessTokenExpiryMins:  accessTokenExpiryMins,
				RefreshTokenExpiryDays: refreshTokenExpiryDays,
			}

			// In production mode, use embedded static files
			var staticFS fs.FS
			if !dev {
				staticFS = DistFS
			}

			srv, err := server.New(dev, dbURL, authConfig, staticFS)
			if err != nil {
				return fmt.Errorf("failed to create server: %w", err)
			}

			// Handle signals
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

			go func() {
				for sig := range sigChan {
					switch sig {
					case syscall.SIGUSR1:
						// Restart Nuxt dev server
						if dev {
							fmt.Println("\nReceived SIGUSR1, restarting Nuxt dev server...")
							if err := startNuxt(); err != nil {
								fmt.Printf("Error restarting Nuxt: %v\n", err)
							}
						}
					case syscall.SIGINT, syscall.SIGTERM:
						fmt.Println("\nShutting down...")

						if nuxtCmd != nil && nuxtCmd.Process != nil {
							nuxtCmd.Process.Signal(syscall.SIGTERM)
							nuxtCmd.Wait()
						}
						os.Remove(NuxtPidFile)

						srv.Close()
						os.Exit(0)
					}
				}
			}()

			return srv.Start(addr)
		},
	}
}
