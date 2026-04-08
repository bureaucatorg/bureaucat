package cli

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/urfave/cli/v3"
)

type myTaskItem struct {
	ID           string         `json:"id"`
	ProjectKey   string         `json:"project_key"`
	TaskNumber   int            `json:"task_number"`
	TaskID       string         `json:"task_id"`
	Title        string         `json:"title"`
	StateName    string         `json:"state_name"`
	StateType    string         `json:"state_type"`
	StateColor   string         `json:"state_color"`
	Priority     int            `json:"priority"`
	Assignees    []taskAssignee `json:"assignees"`
	CommentCount int            `json:"comment_count"`
}

type myTasksResponse struct {
	Tasks      []myTaskItem `json:"tasks"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	PerPage    int          `json:"per_page"`
	TotalPages int          `json:"total_pages"`
}

func MyTasksCommand() *cli.Command {
	return &cli.Command{
		Name:  "mytasks",
		Usage: "List tasks assigned to the current user",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "page", Value: 1, Usage: "Page number"},
			&cli.IntFlag{Name: "per-page", Value: 20, Usage: "Items per page"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			client, err := NewClientFromEnv()
			if err != nil {
				return err
			}

			query := map[string]string{
				"page":     strconv.Itoa(int(cmd.Int("page"))),
				"per_page": strconv.Itoa(int(cmd.Int("per-page"))),
			}

			var raw json.RawMessage
			if err := client.Get("/me/tasks", query, &raw); err != nil {
				return err
			}

			if !IsHuman(cmd) {
				return PrintRawJSON(raw)
			}

			var resp myTasksResponse
			if err := json.Unmarshal(raw, &resp); err != nil {
				return err
			}

			rows := make([][]string, 0, len(resp.Tasks))
			for _, task := range resp.Tasks {
				rows = append(rows, []string{
					task.TaskID,
					truncate(task.Title, 40),
					task.StateName,
					priorityLabel(task.Priority),
					task.ProjectKey,
				})
			}
			return PrintTable([]string{"ID", "TITLE", "STATE", "PRIORITY", "PROJECT"}, rows)
		},
	}
}
