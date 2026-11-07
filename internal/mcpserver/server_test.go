package mcpserver

import (
	"net/http"
	"testing"

	"bereaucat/internal/assistant"
)

func TestAllowedTools(t *testing.T) {
	names := func(tools []assistant.Tool) map[string]assistant.Tool {
		m := map[string]assistant.Tool{}
		for _, tool := range tools {
			m[tool.Name] = tool
		}
		return m
	}
	rw := names(allowedTools(false))
	ro := names(allowedTools(true))

	for _, upload := range []string{"attach_file_to_task", "attach_file_to_comment"} {
		if _, ok := rw[upload]; ok {
			t.Errorf("%s needs a chat upload and must not be exposed", upload)
		}
	}
	for _, want := range []string{"list_tasks", "create_task", "delete_task"} {
		if _, ok := rw[want]; !ok {
			t.Errorf("read_write is missing %s", want)
		}
	}
	if _, ok := ro["list_tasks"]; !ok {
		t.Error("read_only is missing list_tasks")
	}
	for name, tool := range ro {
		if tool.Method != http.MethodGet {
			t.Errorf("read_only exposes write tool %s", name)
		}
	}
	if len(rw)+2 != len(assistant.Catalog()) {
		t.Errorf("read_write should expose every catalog tool except the 2 upload tools, got %d of %d", len(rw), len(assistant.Catalog()))
	}
}

func TestToMCPToolAnnotations(t *testing.T) {
	for _, tool := range allowedTools(false) {
		a := toMCPTool(tool).Annotations
		if a.ReadOnlyHint != (tool.Method == http.MethodGet) {
			t.Errorf("%s: readOnlyHint = %v", tool.Name, a.ReadOnlyHint)
		}
		if *a.DestructiveHint != tool.Confirm {
			t.Errorf("%s: destructiveHint = %v, want %v", tool.Name, *a.DestructiveHint, tool.Confirm)
		}
	}
}
