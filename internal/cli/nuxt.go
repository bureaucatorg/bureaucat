package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/urfave/cli/v3"
)

func NuxtCommand() *cli.Command {
	return &cli.Command{
		Name:  "nuxt",
		Usage: "Nuxt development server commands",
		Commands: []*cli.Command{
			{
				Name:  "restart",
				Usage: "Restart the Nuxt development server (sends signal to serve process)",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					// Read the serve process PID
					pidBytes, err := os.ReadFile(PidFile)
					if err != nil {
						return fmt.Errorf("could not read serve PID file (%s): %w\nMake sure 'bureaucat serve --dev' is running", PidFile, err)
					}

					pid, err := strconv.Atoi(string(pidBytes))
					if err != nil {
						return fmt.Errorf("invalid PID in file: %w", err)
					}

					// Find the process
					process, err := os.FindProcess(pid)
					if err != nil {
						return fmt.Errorf("could not find serve process (PID %d): %w", pid, err)
					}

					// Send SIGUSR1 to restart Nuxt
					fmt.Printf("Sending restart signal to serve process (PID %d)...\n", pid)
					if err := process.Signal(syscall.SIGUSR1); err != nil {
						return fmt.Errorf("could not send signal: %w", err)
					}

					fmt.Println("Restart signal sent. Nuxt dev server will restart.")
					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check the status of the Nuxt development server",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					// Check serve process
					servePidBytes, err := os.ReadFile(PidFile)
					if err != nil {
						fmt.Println("Serve process: not running (no PID file)")
					} else {
						servePid, _ := strconv.Atoi(string(servePidBytes))
						process, err := os.FindProcess(servePid)
						if err != nil || process.Signal(syscall.Signal(0)) != nil {
							fmt.Printf("Serve process: not running (stale PID file, PID %d)\n", servePid)
						} else {
							fmt.Printf("Serve process: running (PID %d)\n", servePid)
						}
					}

					// Check Nuxt process
					nuxtPidBytes, err := os.ReadFile(NuxtPidFile)
					if err != nil {
						fmt.Println("Nuxt process: not running (no PID file)")
					} else {
						nuxtPid, _ := strconv.Atoi(string(nuxtPidBytes))
						process, err := os.FindProcess(nuxtPid)
						if err != nil || process.Signal(syscall.Signal(0)) != nil {
							fmt.Printf("Nuxt process: not running (stale PID file, PID %d)\n", nuxtPid)
						} else {
							fmt.Printf("Nuxt process: running (PID %d)\n", nuxtPid)
						}
					}

					return nil
				},
			},
		},
	}
}
