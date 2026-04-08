package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

type cliUserResponse struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	UserType  string  `json:"user_type"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type loginResult struct {
	URL        string          `json:"url"`
	ConfigPath string          `json:"config_path"`
	AuthMethod string          `json:"auth_method"`
	User       cliUserResponse `json:"user"`
}

func LoginCommand() *cli.Command {
	return &cli.Command{
		Name:  "login",
		Usage: "Authenticate with a Bureaucat server",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "url", Usage: "Bureaucat server URL"},
			&cli.StringFlag{Name: "token", Usage: "Personal access token"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			url := strings.TrimSpace(cmd.String("url"))
			if url == "" {
				value, err := promptString("Bureaucat URL")
				if err != nil {
					return fmt.Errorf("read URL: %w", err)
				}
				url = value
			}

			token := strings.TrimSpace(cmd.String("token"))
			if token == "" {
				value, err := promptString("Personal access token")
				if err != nil {
					return fmt.Errorf("read personal access token: %w", err)
				}
				token = value
			}

			return loginWithPAT(url, token, IsHuman(cmd))
		},
	}
}

func loginWithPAT(baseURL, token string, human bool) error {
	client := NewClient(baseURL, token)
	var user cliUserResponse
	if err := client.Get("/me", nil, &user); err != nil {
		if err == ErrUnauthorized {
			return fmt.Errorf("invalid token")
		}
		return err
	}

	path, err := SaveConfig(Config{URL: baseURL, Token: token})
	if err != nil {
		return err
	}

	if human {
		fmt.Fprintf(os.Stdout, "Logged in as @%s (%s)\n", user.Username, user.UserType)
		fmt.Fprintf(os.Stdout, "Config saved to %s\n", path)
		return nil
	}

	return PrintJSON(loginResult{
		URL:        baseURL,
		ConfigPath: path,
		AuthMethod: "pat",
		User:       user,
	})
}
