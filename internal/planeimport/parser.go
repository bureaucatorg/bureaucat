package planeimport

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// targetTables lists the Plane tables we want to extract from the dump.
var targetTables = map[string]bool{
	"workspaces":      true,
	"projects":        true,
	"users":           true,
	"states":          true,
	"labels":          true,
	"issues":          true,
	"issue_assignees": true,
	"issue_labels":    true,
	"issue_comments":  true,
	"project_members": true,
}

// copyHeaderRe matches COPY public.<table> (col1, col2, ...) FROM stdin;
var copyHeaderRe = regexp.MustCompile(`^COPY public\.(\w+) \((.+)\) FROM stdin;$`)

// ParsedDump holds the extracted data from a Plane SQL dump.
type ParsedDump struct {
	Workspaces     []map[string]string
	Projects       []map[string]string
	Users          []map[string]string
	States         []map[string]string
	Labels         []map[string]string
	Issues         []map[string]string
	IssueAssignees []map[string]string
	IssueLabels    []map[string]string
	IssueComments  []map[string]string
	ProjectMembers []map[string]string
}

// Parse reads a pg_dump SQL file and extracts COPY data for target tables.
func Parse(r io.Reader) (*ParsedDump, error) {
	scanner := bufio.NewScanner(r)
	// Allow up to 10MB per line for wide rows.
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	dump := &ParsedDump{}

	for scanner.Scan() {
		line := scanner.Text()

		matches := copyHeaderRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		tableName := matches[1]
		if !targetTables[tableName] {
			continue
		}

		columns := parseColumns(matches[2])
		rows, err := readCopyData(scanner, columns)
		if err != nil {
			return nil, fmt.Errorf("reading COPY data for %s: %w", tableName, err)
		}

		switch tableName {
		case "workspaces":
			dump.Workspaces = rows
		case "projects":
			dump.Projects = rows
		case "users":
			dump.Users = rows
		case "states":
			dump.States = rows
		case "labels":
			dump.Labels = rows
		case "issues":
			dump.Issues = rows
		case "issue_assignees":
			dump.IssueAssignees = rows
		case "issue_labels":
			dump.IssueLabels = rows
		case "issue_comments":
			dump.IssueComments = rows
		case "project_members":
			dump.ProjectMembers = rows
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning dump file: %w", err)
	}

	return dump, nil
}

// parseColumns splits the column list from a COPY header.
func parseColumns(raw string) []string {
	parts := strings.Split(raw, ", ")
	cols := make([]string, len(parts))
	for i, p := range parts {
		// Strip surrounding quotes if present (e.g., "group", "default").
		cols[i] = strings.Trim(strings.TrimSpace(p), "\"")
	}
	return cols
}

// readCopyData reads tab-delimited rows until the \. terminator.
func readCopyData(scanner *bufio.Scanner, columns []string) ([]map[string]string, error) {
	var rows []map[string]string

	for scanner.Scan() {
		line := scanner.Text()
		if line == "\\." {
			break
		}

		fields := strings.Split(line, "\t")
		if len(fields) != len(columns) {
			return nil, fmt.Errorf("column count mismatch: expected %d, got %d", len(columns), len(fields))
		}

		row := make(map[string]string, len(columns))
		for i, col := range columns {
			val := fields[i]
			if val == "\\N" {
				val = ""
			}
			row[col] = val
		}
		rows = append(rows, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rows, nil
}
