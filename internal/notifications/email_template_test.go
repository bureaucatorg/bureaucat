package notifications

import (
	"os"
	"strings"
	"testing"

	"bereaucat/internal/branding"
	"bereaucat/internal/mailer"
)

// Set EMAIL_PREVIEW=/path/to/file.html to dump the rendered digest for a visual check.
func TestEverythingSample(t *testing.T) {
	msg, err := SampleEmail(mailer.Settings{AppURL: "https://app.example.com"}, branding.Brand{Name: "Bureaucat"}, "user@example.com", "everything", "Pat", "Placeholder")
	if err != nil {
		t.Fatal(err)
	}
	if path := os.Getenv("EMAIL_PREVIEW"); path != "" {
		if err := os.WriteFile(path, []byte(msg.HTML), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, want := range []string{
		"Created the task", "Assigned <b style=\"color:#09090b\">Pat Placeholder</b>",
		"Workflow Engine", "Jordan Doe, Casey Tester", "29 updates", "Riley Placeholder, Morgan Demo",
		"Approval Pending", "ISec Approved", "Set priority to", "Urgent",
		"Linked a blocker", "Removed the start date", "Workflow Engine",
	} {
		if !strings.Contains(msg.HTML, want) {
			t.Errorf("HTML missing %q", want)
		}
	}
	if !strings.Contains(msg.Text, "↳ Workflow Engine unassigned Jordan Doe, Casey Tester") {
		t.Errorf("text missing nested workflow line:\n%s", msg.Text)
	}
}

func TestGroupByActor(t *testing.T) {
	events := []emailEvent{
		newEvent("Alex", "Example", "created the task"),
		newEvent("Alex", "Example", "added a comment"),
		newEvent("Jordan", "Doe", "added a comment"),
		newEvent("Alex", "Example", "changed state"),
	}
	groups := groupByActor(events)
	if len(groups) != 3 || len(groups[0].Items) != 2 {
		t.Fatalf("want 3 groups with the first holding 2 items, got %+v", groups)
	}
	if groups[0].Items[0].Verb != "Created the task" || events[0].Verb != "created the task" {
		t.Errorf("verb should be capitalised in the group only, got %q / %q", groups[0].Items[0].Verb, events[0].Verb)
	}
}
