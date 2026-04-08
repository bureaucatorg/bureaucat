package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

type createTaskRequest struct {
	Title       string   `json:"title"`
	Description *string  `json:"description,omitempty"`
	StateID     *string  `json:"state_id,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	Assignees   []string `json:"assignees,omitempty"`
	Labels      []string `json:"labels,omitempty"`
}

func TaskCreateCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Create a task in a project",
		ArgsUsage: "<project-key>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "title", Usage: "Task title", Required: true},
			&cli.StringFlag{Name: "description", Usage: "Task description"},
			&cli.StringFlag{Name: "state-id", Usage: "Initial state ID"},
			&cli.IntFlag{Name: "priority", Usage: "Priority (0-4)"},
			&cli.StringSliceFlag{Name: "assignee", Usage: "Assignee user ID (repeatable)"},
			&cli.StringSliceFlag{Name: "label", Usage: "Label ID (repeatable)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if err := RequireYes(cmd); err != nil {
				return err
			}
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("project key is required")
			}

			req := createTaskRequest{
				Title:     strings.TrimSpace(cmd.String("title")),
				Assignees: cmd.StringSlice("assignee"),
				Labels:    cmd.StringSlice("label"),
			}
			if description := strings.TrimSpace(cmd.String("description")); description != "" {
				req.Description = &description
			}
			if stateID := strings.TrimSpace(cmd.String("state-id")); stateID != "" {
				req.StateID = &stateID
			}
			if cmd.IsSet("priority") {
				priority := int(cmd.Int("priority"))
				req.Priority = &priority
			}

			client, err := NewClientFromEnv()
			if err != nil {
				return err
			}

			projectKey := strings.TrimSpace(cmd.Args().First())
			var raw json.RawMessage
			if err := client.Post("/projects/"+projectKey+"/tasks", req, &raw); err != nil {
				return err
			}

			if !IsHuman(cmd) {
				return PrintRawJSON(raw)
			}

			var task cliTaskResponse
			if err := json.Unmarshal(raw, &task); err != nil {
				return err
			}
			fmt.Fprintf(cmd.Root().Writer, "Created %s\n", task.TaskID)
			return nil
		},
	}
}
