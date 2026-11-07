package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
