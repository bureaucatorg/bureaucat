package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"
)

type taskAssignee struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type taskLabel struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type cliTaskResponse struct {
	ID               string         `json:"id"`
	ProjectKey       string         `json:"project_key"`
	TaskNumber       int            `json:"task_number"`
	TaskID           string         `json:"task_id"`
	Title            string         `json:"title"`
	Description      *string        `json:"description,omitempty"`
	StateID          string         `json:"state_id"`
	StateName        string         `json:"state_name"`
	StateType        string         `json:"state_type"`
	StateColor       string         `json:"state_color"`
	Priority         int            `json:"priority"`
	CreatedBy        string         `json:"created_by"`
	CreatorUsername  string         `json:"creator_username"`
	CreatorFirstName string         `json:"creator_first_name"`
	CreatorLastName  string         `json:"creator_last_name"`
	Assignees        []taskAssignee `json:"assignees,omitempty"`
	Labels           []taskLabel    `json:"labels,omitempty"`
	CommentCount     int            `json:"comment_count"`
	CreatedAt        string         `json:"created_at"`
	UpdatedAt        string         `json:"updated_at"`
}

type cliPaginatedTasksResponse struct {
	Tasks      []cliTaskResponse `json:"tasks"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	TotalPages int               `json:"total_pages"`
}

func TasksCommand() *cli.Command {
	return &cli.Command{
		Name:      "tasks",
		Usage:     "List tasks in a project",
		ArgsUsage: "<project-key>",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "page", Value: 1, Usage: "Page number"},
			&cli.IntFlag{Name: "per-page", Value: 20, Usage: "Items per page"},
			&cli.StringFlag{Name: "state-id", Usage: "Filter by state ID"},
			&cli.StringFlag{Name: "state-type", Usage: "Filter by state type"},
			&cli.StringFlag{Name: "created-by", Usage: "Filter by creator ID"},
			&cli.StringFlag{Name: "assigned-to", Usage: "Filter by assignee ID"},
			&cli.StringFlag{Name: "priority", Usage: "Filter by priority"},
			&cli.StringFlag{Name: "q", Usage: "Search text"},
			&cli.StringFlag{Name: "from-date", Usage: "Filter from date (YYYY-MM-DD)"},
			&cli.StringFlag{Name: "to-date", Usage: "Filter to date (YYYY-MM-DD)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("project key is required")
			}
			projectKey := strings.TrimSpace(cmd.Args().First())

			client, err := NewClientFromEnv()
			if err != nil {
				return err
			}

			query := map[string]string{
				"page":        strconv.Itoa(int(cmd.Int("page"))),
				"per_page":    strconv.Itoa(int(cmd.Int("per-page"))),
				"state_id":    cmd.String("state-id"),
				"state_type":  cmd.String("state-type"),
				"created_by":  cmd.String("created-by"),
				"assigned_to": cmd.String("assigned-to"),
				"priority":    cmd.String("priority"),
				"q":           cmd.String("q"),
				"from_date":   cmd.String("from-date"),
				"to_date":     cmd.String("to-date"),
			}

			var raw json.RawMessage
			if err := client.Get("/projects/"+projectKey+"/tasks", query, &raw); err != nil {
				return err
			}

			if !IsHuman(cmd) {
				return PrintRawJSON(raw)
			}

			var resp cliPaginatedTasksResponse
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
					truncate(joinAssignees(task.Assignees), 28),
				})
			}

			return PrintTable([]string{"ID", "TITLE", "STATE", "PRIORITY", "ASSIGNEES"}, rows)
		},
	}
}

func TaskCommand() *cli.Command {
	return &cli.Command{
		Name:      "task",
		Usage:     "Show or mutate a task",
		ArgsUsage: "<project-num>",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return fmt.Errorf("task ID is required")
			}
			projectKey, taskNum, err := parseTaskID(strings.TrimSpace(cmd.Args().First()))
			if err != nil {
				return err
			}

			client, err := NewClientFromEnv()
			if err != nil {
				return err
			}

			var raw json.RawMessage
			if err := client.Get("/projects/"+projectKey+"/tasks/"+taskNum, nil, &raw); err != nil {
				return err
			}

			if !IsHuman(cmd) {
				return PrintRawJSON(raw)
			}

			var task cliTaskResponse
			if err := json.Unmarshal(raw, &task); err != nil {
				return err
			}

			labels := make([]string, 0, len(task.Labels))
			for _, label := range task.Labels {
				labels = append(labels, label.Name)
			}

			if err := PrintKeyValue([][2]string{
				{"Task", task.TaskID},
				{"Title", task.Title},
				{"State", fmt.Sprintf("%s (%s)", task.StateName, task.StateType)},
				{"Priority", fmt.Sprintf("%d (%s)", task.Priority, priorityLabel(task.Priority))},
				{"Created by", "@" + task.CreatorUsername},
				{"Assignees", joinAssignees(task.Assignees)},
				{"Labels", strings.Join(labels, ", ")},
				{"Comments", fmt.Sprintf("%d", task.CommentCount)},
				{"Created", task.CreatedAt},
				{"Updated", task.UpdatedAt},
			}); err != nil {
				return err
			}

			if task.Description != nil && strings.TrimSpace(*task.Description) != "" {
				fmt.Println()
				fmt.Println("Description:")
				fmt.Println(*task.Description)
			}

			return nil
		},
		Commands: []*cli.Command{TaskCreateCommand(), TaskUpdateCommand()},
	}
}
