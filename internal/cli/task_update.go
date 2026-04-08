package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

type updateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	StateID     *string `json:"state_id,omitempty"`
	Priority    *int    `json:"priority,omitempty"`
}

func TaskUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Usage:     "Update an existing task",
		ArgsUsage: "<project-num>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "title", Usage: "New title"},
			&cli.StringFlag{Name: "description", Usage: "New description"},
			&cli.StringFlag{Name: "state-id", Usage: "New state ID"},
			&cli.IntFlag{Name: "priority", Usage: "New priority (0-4)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if err := RequireYes(cmd); err != nil {
				return err
			}
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("task ID is required")
			}

			projectKey, taskNum, err := parseTaskID(strings.TrimSpace(cmd.Args().First()))
			if err != nil {
				return err
			}

			req := updateTaskRequest{}
			var setCount int
			if cmd.IsSet("title") {
				value := strings.TrimSpace(cmd.String("title"))
				req.Title = &value
				setCount++
			}
			if cmd.IsSet("description") {
				value := cmd.String("description")
				req.Description = &value
				setCount++
			}
			if cmd.IsSet("state-id") {
				value := strings.TrimSpace(cmd.String("state-id"))
				req.StateID = &value
				setCount++
			}
			if cmd.IsSet("priority") {
				value := int(cmd.Int("priority"))
				req.Priority = &value
				setCount++
			}
			if setCount == 0 {
				return fmt.Errorf("no fields provided to update")
			}

			client, err := NewClientFromEnv()
			if err != nil {
				return err
			}
			var raw json.RawMessage
			if err := client.Patch("/projects/"+projectKey+"/tasks/"+taskNum, req, &raw); err != nil {
				return err
			}

			if !IsHuman(cmd) {
				return PrintRawJSON(raw)
			}

			var task cliTaskResponse
			if err := json.Unmarshal(raw, &task); err != nil {
				return err
			}
			fmt.Fprintf(cmd.Root().Writer, "Updated %s\n", task.TaskID)
			return nil
		},
	}
}
