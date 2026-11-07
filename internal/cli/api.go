package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"

	"bereaucat/internal/assistant"
)

// APICommand exposes every assistant catalog tool as `bureaucat api <tool>`, so the CLI
// tracks the same API surface as the chat assistant and the MCP server.
func APICommand() *cli.Command {
	var cmds []*cli.Command
	for _, t := range assistant.Catalog() {
		if t.UsesUpload() {
			continue
		}
		cmds = append(cmds, toolCommand(t))
	}
	return &cli.Command{
		Name:     "api",
		Usage:    "Call a Bureaucat API operation (JSON output)",
		Commands: cmds,
	}
}

func toolCommand(t assistant.Tool) *cli.Command {
	flags := make([]cli.Flag, 0, len(t.Params))
	for _, p := range t.Params {
		flags = append(flags, paramFlag(p))
	}
	return &cli.Command{
		Name:  kebab(t.Name),
		Usage: t.Description,
		Flags: flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if t.Method != http.MethodGet {
				if err := RequireYes(cmd); err != nil {
					return err
				}
			}
			args, err := toolArgs(cmd, t.Params)
			if err != nil {
				return err
			}
			target, body, err := t.BuildRequest(args)
			if err != nil {
				return err
			}
			client, err := NewClientFromEnv()
			if err != nil {
				return err
			}
			raw, err := client.Send(t.Method, target, body)
			if err != nil || len(bytes.TrimSpace(raw)) == 0 {
				return err
			}
			return PrintRawJSON(raw)
		},
	}
}

func paramFlag(p assistant.Param) cli.Flag {
	usage := p.Description
	if len(p.Enum) > 0 {
		usage += " (one of: " + strings.Join(p.Enum, ", ") + ")"
	}
	switch p.Type {
	case "integer":
		return &cli.IntFlag{Name: kebab(p.Name), Usage: usage, Required: p.Required}
	case "boolean":
		return &cli.BoolFlag{Name: kebab(p.Name), Usage: usage, Required: p.Required}
	case "string[]", "integer[]":
		return &cli.StringSliceFlag{Name: kebab(p.Name), Usage: usage + " (repeatable)", Required: p.Required}
	case "object", "object[]":
		return &cli.StringFlag{Name: kebab(p.Name), Usage: usage + " (JSON)", Required: p.Required}
	default:
		return &cli.StringFlag{Name: kebab(p.Name), Usage: usage, Required: p.Required}
	}
}

func toolArgs(cmd *cli.Command, params []assistant.Param) (map[string]any, error) {
	args := map[string]any{}
	for _, p := range params {
		name := kebab(p.Name)
		if !cmd.IsSet(name) {
			continue
		}
		switch p.Type {
		case "integer":
			args[p.Name] = cmd.Int(name)
		case "boolean":
			args[p.Name] = cmd.Bool(name)
		case "string[]":
			args[p.Name] = cmd.StringSlice(name)
		case "integer[]":
			var ints []int64
			for _, s := range cmd.StringSlice(name) {
				n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
				if err != nil {
					return nil, fmt.Errorf("--%s: %q is not an integer", name, s)
				}
				ints = append(ints, n)
			}
			args[p.Name] = ints
		case "object", "object[]":
			var v any
			if err := json.Unmarshal([]byte(cmd.String(name)), &v); err != nil {
				return nil, fmt.Errorf("--%s: invalid JSON: %w", name, err)
			}
			args[p.Name] = v
		case "string|null":
			if v := cmd.String(name); v != "null" {
				args[p.Name] = v
			} else {
				args[p.Name] = nil
			}
		default:
			args[p.Name] = cmd.String(name)
		}
	}
	return args, nil
}

func kebab(s string) string { return strings.ReplaceAll(s, "_", "-") }
