package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli/v3"
)

func IsHuman(cmd *cli.Command) bool {
	return cmd.Bool("human")
}

func RequireYes(cmd *cli.Command) error {
	if cmd.Bool("yes") {
		return nil
	}
	return fmt.Errorf("this command mutates remote state; rerun with --yes to confirm")
}

func PrintJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(os.Stdout, string(data))
	return err
}

func PrintRawJSON(raw json.RawMessage) error {
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		_, writeErr := fmt.Fprintln(os.Stdout, string(raw))
		if writeErr != nil {
			return writeErr
		}
		return nil
	}
	return PrintJSON(v)
}

func PrintTable(headers []string, rows [][]string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, strings.Join(headers, "\t")); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintln(w, strings.Join(row, "\t")); err != nil {
			return err
		}
	}
	return w.Flush()
}

func PrintKeyValue(pairs [][2]string) error {
	maxWidth := 0
	for _, pair := range pairs {
		if len(pair[0]) > maxWidth {
			maxWidth = len(pair[0])
		}
	}
	for _, pair := range pairs {
		if _, err := fmt.Fprintf(os.Stdout, "%*s: %s\n", maxWidth, pair[0], pair[1]); err != nil {
			return err
		}
	}
	return nil
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return s[:max]
	}
	return s[:max-1] + "…"
}

func joinAssignees(assignees []taskAssignee) string {
	if len(assignees) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(assignees))
	for _, assignee := range assignees {
		if assignee.Username != "" {
			parts = append(parts, "@"+assignee.Username)
			continue
		}
		parts = append(parts, assignee.UserID)
	}
	return strings.Join(parts, ", ")
}

func priorityLabel(priority int) string {
	switch priority {
	case 0:
		return "none"
	case 1:
		return "low"
	case 2:
		return "medium"
	case 3:
		return "high"
	case 4:
		return "urgent"
	default:
		return fmt.Sprintf("%d", priority)
	}
}

func parseTaskID(taskID string) (string, string, error) {
	idx := strings.LastIndex(taskID, "-")
	if idx <= 0 || idx == len(taskID)-1 {
		return "", "", fmt.Errorf("invalid task ID %q; expected format PROJECT-123", taskID)
	}
	return taskID[:idx], taskID[idx+1:], nil
}

func promptString(label string) (string, error) {
	_, err := fmt.Fprintf(os.Stdout, "%s: ", label)
	if err != nil {
		return "", err
	}
	var value string
	if _, err := fmt.Fscanln(os.Stdin, &value); err != nil {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func openEditor() (string, error) {
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if editor == "" {
		editor = "vi"
	}

	file, err := os.CreateTemp("", "bureaucat-comment-*.md")
	if err != nil {
		return "", err
	}
	path := file.Name()
	if err := file.Close(); err != nil {
		return "", err
	}
	defer os.Remove(path)

	cmd := exec.Command("sh", "-c", editor+" \"$1\"", "sh", path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return "", errors.New("empty content")
	}
	return text, nil
}
