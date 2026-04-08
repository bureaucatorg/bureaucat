package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

type createCommentRequest struct {
	Content string `json:"content"`
}

func CommentCommand() *cli.Command {
	return &cli.Command{
		Name:      "comment",
		Usage:     "Add a comment to a task",
		ArgsUsage: "<project-num> [message]",
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

			content := strings.TrimSpace(cmd.Args().Get(1))
			if content == "" {
				content, err = openEditor()
				if err != nil {
					return err
				}
			}

			client, err := NewClientFromEnv()
			if err != nil {
				return err
			}
			var raw json.RawMessage
			if err := client.Post("/projects/"+projectKey+"/tasks/"+taskNum+"/comments", createCommentRequest{Content: content}, &raw); err != nil {
				return err
			}

			if !IsHuman(cmd) {
				return PrintRawJSON(raw)
			}

			fmt.Fprintf(cmd.Root().Writer, "Comment added to %s-%s\n", projectKey, taskNum)
			return nil
		},
	}
}
