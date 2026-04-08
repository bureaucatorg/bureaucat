package cli

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/urfave/cli/v3"
)

type cliProjectResponse struct {
	ID          string  `json:"id"`
	ProjectKey  string  `json:"project_key"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Role        string  `json:"role,omitempty"`
}

type cliPaginatedProjectsResponse struct {
	Projects   []cliProjectResponse `json:"projects"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PerPage    int                  `json:"per_page"`
	TotalPages int                  `json:"total_pages"`
}

func ProjectsCommand() *cli.Command {
	return &cli.Command{
		Name:  "projects",
		Usage: "List projects",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "page", Value: 1, Usage: "Page number"},
			&cli.IntFlag{Name: "per-page", Value: 20, Usage: "Items per page"},
			&cli.StringFlag{Name: "search", Usage: "Search by name"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			client, err := NewClientFromEnv()
			if err != nil {
				return err
			}

			query := map[string]string{
				"page":     strconv.Itoa(int(cmd.Int("page"))),
				"per_page": strconv.Itoa(int(cmd.Int("per-page"))),
				"search":   cmd.String("search"),
			}

			var raw json.RawMessage
			if err := client.Get("/projects", query, &raw); err != nil {
				return err
			}

			if !IsHuman(cmd) {
				return PrintRawJSON(raw)
			}

			var resp cliPaginatedProjectsResponse
			if err := json.Unmarshal(raw, &resp); err != nil {
				return err
			}

			rows := make([][]string, 0, len(resp.Projects))
			for _, project := range resp.Projects {
				description := ""
				if project.Description != nil {
					description = truncate(*project.Description, 40)
				}
				rows = append(rows, []string{project.ProjectKey, project.Name, project.Role, description})
			}
			return PrintTable([]string{"KEY", "NAME", "ROLE", "DESCRIPTION"}, rows)
		},
	}
}
