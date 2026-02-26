package notifier

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// EventType represents the type of notification event.
type EventType string

const (
	EventTaskAssigned EventType = "task_assigned"
	EventMentioned    EventType = "mentioned"
)

// Notification represents a single notification to deliver.
type Notification struct {
	Event       EventType
	RecipientID uuid.UUID // Bureaucat user ID to notify
	ActorName   string    // Who triggered the notification (e.g. "John Doe")
	ProjectKey  string    // e.g. "DEVOP"
	TaskNumber  int       // e.g. 123
	TaskTitle   string
	BaseURL     string    // App base URL for building task links (e.g. "https://bureaucat.example.com")
}

// TaskURL returns the full URL to the task.
func (n Notification) TaskURL() string {
	return fmt.Sprintf("%s/projects/%s/tasks/%d", n.BaseURL, n.ProjectKey, n.TaskNumber)
}

// Notifier is the interface that all notification providers must implement.
type Notifier interface {
	// Name returns the provider name (e.g. "mattermost", "slack").
	Name() string
	// Send delivers a notification to the recipient identified by email.
	Send(ctx context.Context, recipientEmail string, notification Notification) error
}
