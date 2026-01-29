package activity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"bereaucat/internal/store"
)

// SeedChecksum is the initial checksum used for the first activity entry
const SeedChecksum = "0000000000000000000000000000000000000000000000000000000000000000"

// ActivityType represents the type of activity logged
type ActivityType string

const (
	TaskCreated     ActivityType = "task_created"
	TaskUpdated     ActivityType = "task_updated"
	TaskDeleted     ActivityType = "task_deleted"
	AssigneeAdded   ActivityType = "assignee_added"
	AssigneeRemoved ActivityType = "assignee_removed"
	LabelAdded      ActivityType = "label_added"
	LabelRemoved    ActivityType = "label_removed"
	StateChanged    ActivityType = "state_changed"
	CommentCreated  ActivityType = "comment_created"
	CommentUpdated  ActivityType = "comment_updated"
	CommentDeleted  ActivityType = "comment_deleted"
)

// Service handles tamper-proof activity logging
type Service struct {
	store store.Querier
}

// NewService creates a new activity service
func NewService(queryer store.Querier) *Service {
	return &Service{store: queryer}
}

// LogActivityParams contains parameters for logging an activity
type LogActivityParams struct {
	TaskID       uuid.UUID
	ActivityType ActivityType
	ActorID      uuid.UUID
	FieldName    *string
	OldValue     interface{}
	NewValue     interface{}
}

// LogActivity creates a new activity log entry with checksum verification
func (s *Service) LogActivity(ctx context.Context, params LogActivityParams) error {
	// Get the last checksum for this task
	var previousChecksum string
	lastChecksum, err := s.store.GetLastActivityChecksum(ctx, params.TaskID)
	if err != nil {
		// No previous activity, use seed
		previousChecksum = SeedChecksum
	} else {
		previousChecksum = lastChecksum
	}

	// Marshal old and new values to JSON
	var oldValueJSON, newValueJSON []byte
	if params.OldValue != nil {
		oldValueJSON, err = json.Marshal(params.OldValue)
		if err != nil {
			return fmt.Errorf("failed to marshal old value: %w", err)
		}
	}
	if params.NewValue != nil {
		newValueJSON, err = json.Marshal(params.NewValue)
		if err != nil {
			return fmt.Errorf("failed to marshal new value: %w", err)
		}
	}

	// Generate timestamp - truncate to microseconds to match PostgreSQL precision
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Calculate checksum
	checksum := calculateChecksum(
		previousChecksum,
		params.TaskID,
		string(params.ActivityType),
		params.ActorID,
		params.FieldName,
		oldValueJSON,
		newValueJSON,
		now,
	)

	// Create activity log entry
	_, err = s.store.CreateActivityLog(ctx, store.CreateActivityLogParams{
		TaskID:       params.TaskID,
		ActivityType: string(params.ActivityType),
		ActorID:      params.ActorID,
		FieldName:    nullableText(params.FieldName),
		OldValue:     oldValueJSON,
		NewValue:     newValueJSON,
		Checksum:     checksum,
		CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create activity log: %w", err)
	}

	return nil
}

// VerifyChain verifies the integrity of the activity log chain for a task
// Returns true if the chain is valid, false if tampering is detected
func (s *Service) VerifyChain(ctx context.Context, taskID uuid.UUID) (bool, error) {
	activities, err := s.store.VerifyActivityChain(ctx, taskID)
	if err != nil {
		return false, fmt.Errorf("failed to get activity chain: %w", err)
	}

	if len(activities) == 0 {
		return true, nil
	}

	previousChecksum := SeedChecksum
	for _, activity := range activities {
		// Normalize JSON to ensure consistent key ordering
		oldValue := normalizeJSON(activity.OldValue)
		newValue := normalizeJSON(activity.NewValue)

		expectedChecksum := calculateChecksum(
			previousChecksum,
			activity.TaskID,
			activity.ActivityType,
			activity.ActorID,
			textToStringPtr(activity.FieldName),
			oldValue,
			newValue,
			activity.CreatedAt.Time,
		)

		if activity.Checksum != expectedChecksum {
			return false, nil
		}

		previousChecksum = activity.Checksum
	}

	return true, nil
}

// calculateChecksum generates a SHA-256 checksum for an activity entry
func calculateChecksum(
	previousChecksum string,
	taskID uuid.UUID,
	activityType string,
	actorID uuid.UUID,
	fieldName *string,
	oldValue []byte,
	newValue []byte,
	timestamp time.Time,
) string {
	// Build the data string to hash
	fieldNameStr := ""
	if fieldName != nil {
		fieldNameStr = *fieldName
	}

	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s",
		previousChecksum,
		taskID.String(),
		activityType,
		actorID.String(),
		fieldNameStr,
		string(oldValue),
		string(newValue),
		timestamp.Format(time.RFC3339Nano),
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// normalizeJSON unmarshals and re-marshals JSON to ensure consistent key ordering
// Go's json.Marshal sorts map keys alphabetically
func normalizeJSON(data []byte) []byte {
	if data == nil || len(data) == 0 {
		return data
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return data // Return original if unmarshal fails
	}

	normalized, err := json.Marshal(v)
	if err != nil {
		return data // Return original if marshal fails
	}

	return normalized
}

// nullableText converts a *string to pgtype.Text
func nullableText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// textToStringPtr converts a pgtype.Text to *string
func textToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// Helper function to create a string pointer
func StringPtr(s string) *string {
	return &s
}
