package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

func WhoamiCommand() *cli.Command {
	return &cli.Command{
		Name:  "whoami",
		Usage: "Show the authenticated user",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			client, err := NewClientFromEnv()
			if err != nil {
				return err
			}

			var raw json.RawMessage
			if err := client.Get("/me", nil, &raw); err != nil {
				return err
			}

			if !IsHuman(cmd) {
				return PrintRawJSON(raw)
			}

			var user cliUserResponse
			if err := json.Unmarshal(raw, &user); err != nil {
				return err
			}

			name := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
			return PrintKeyValue([][2]string{
				{"Username", "@" + user.Username},
				{"Name", strings.TrimSpace(name)},
				{"Email", user.Email},
				{"Role", user.UserType},
			})
		},
	}
}
