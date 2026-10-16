package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"bereaucat/internal/store"
)

type ConversationView struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MessageView struct {
	ID          uuid.UUID      `json:"id"`
	Role        string         `json:"role"`
	Content     string         `json:"content"`
	ToolCalls   []ToolCallView `json:"tool_calls"`
	Attachments []Attachment   `json:"attachments"`
	Meta        Meta           `json:"meta"`
	CreatedAt   time.Time      `json:"created_at"`
}

// ToolCallView is a tool call plus whether it returned output; the output itself is
// fetched separately with ToolOutput so conversations stay small.
type ToolCallView struct {
	ToolCall
	HasOutput bool `json:"has_output"`
	IsError   bool `json:"is_error"`
}

var toolErrorRe = regexp.MustCompile(`^HTTP [45]\d\d:`)

type ActionView struct {
	ID         uuid.UUID       `json:"id"`
	ToolCallID string          `json:"tool_call_id"`
	Tool       string          `json:"tool"`
	Title      string          `json:"title"`
	Args       json.RawMessage `json:"args"`
	Status     string          `json:"status"`
	Result     string          `json:"result,omitempty"`
}

type ConversationDetail struct {
	ConversationView
	Messages []MessageView `json:"messages"`
	Actions  []ActionView  `json:"actions"`
}

func (s *Service) ListConversations(ctx context.Context, userID uuid.UUID) ([]ConversationView, error) {
	rows, err := s.store.ListAssistantConversations(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]ConversationView, 0, len(rows))
	for _, r := range rows {
		out = append(out, conversationView(r))
	}
	return out, nil
}

func (s *Service) CreateConversation(ctx context.Context, userID uuid.UUID) (ConversationView, error) {
	row, err := s.store.CreateAssistantConversation(ctx, store.CreateAssistantConversationParams{UserID: userID})
	if err != nil {
		return ConversationView{}, err
	}
	return conversationView(row), nil
}

func (s *Service) DeleteConversation(ctx context.Context, userID, convID uuid.UUID) error {
	n, err := s.store.DeleteAssistantConversation(ctx, store.DeleteAssistantConversationParams{ID: convID, UserID: userID})
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

const maxTitleRunes = 100

func (s *Service) RenameConversation(ctx context.Context, userID, convID uuid.UUID, title string) (string, error) {
	title = strings.TrimSpace(strings.ReplaceAll(title, "\n", " "))
	if title == "" {
		return "", ErrEmptyTitle
	}
	if r := []rune(title); len(r) > maxTitleRunes {
		title = string(r[:maxTitleRunes])
	}
	n, err := s.store.RenameAssistantConversation(ctx, store.RenameAssistantConversationParams{ID: convID, UserID: userID, Title: title})
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", ErrNotFound
	}
	return title, nil
}

// GetConversation returns user/assistant messages (tool output is omitted) and all actions.
func (s *Service) GetConversation(ctx context.Context, userID, convID uuid.UUID) (ConversationDetail, error) {
	conv, err := s.store.GetAssistantConversation(ctx, store.GetAssistantConversationParams{ID: convID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return ConversationDetail{}, ErrNotFound
	}
	if err != nil {
		return ConversationDetail{}, err
	}
	rows, err := s.store.ListAssistantMessages(ctx, convID)
	if err != nil {
		return ConversationDetail{}, err
	}
	actions, err := s.store.ListAssistantActions(ctx, convID)
	if err != nil {
		return ConversationDetail{}, err
	}

	detail := ConversationDetail{
		ConversationView: conversationView(conv),
		Messages:         []MessageView{},
		Actions:          make([]ActionView, 0, len(actions)),
	}
	outputs := map[string]string{}
	for _, r := range rows {
		if r.Role == RoleTool {
			outputs[r.ToolCallID] = r.Content
		}
	}
	for _, r := range rows {
		if r.Role == RoleTool {
			continue
		}
		var raw []ToolCall
		_ = json.Unmarshal(r.ToolCalls, &raw)
		calls := make([]ToolCallView, 0, len(raw))
		for _, tc := range raw {
			out, ok := outputs[tc.ID]
			calls = append(calls, ToolCallView{ToolCall: tc, HasOutput: ok, IsError: toolErrorRe.MatchString(out)})
		}
		detail.Messages = append(detail.Messages, MessageView{
			ID:          r.ID,
			Role:        r.Role,
			Content:     r.Content,
			ToolCalls:   calls,
			Attachments: messageAttachments(r),
			Meta:        messageMeta(r),
			CreatedAt:   r.CreatedAt.Time,
		})
	}
	for _, a := range actions {
		detail.Actions = append(detail.Actions, *actionView(a))
	}
	return detail, nil
}

// ToolOutput returns the full stored output of one tool call in the user's conversation.
func (s *Service) ToolOutput(ctx context.Context, userID, convID uuid.UUID, callID string) (string, error) {
	_, err := s.store.GetAssistantConversation(ctx, store.GetAssistantConversationParams{ID: convID, UserID: userID})
	if err == nil {
		var out string
		out, err = s.store.GetAssistantToolOutput(ctx, store.GetAssistantToolOutputParams{ConversationID: convID, ToolCallID: callID})
		if err == nil {
			return out, nil
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return "", err
}

func conversationView(c store.AssistantConversation) ConversationView {
	return ConversationView{ID: c.ID, Title: c.Title, UpdatedAt: c.UpdatedAt.Time}
}

func actionView(a store.AssistantAction) *ActionView {
	title := a.ToolName
	if t := toolByName(a.ToolName); t != nil {
		title = t.Title
	}
	result := ""
	if a.Status == "failed" {
		result = truncate(a.Result, 500)
	}
	return &ActionView{
		ID:         a.ID,
		ToolCallID: a.ToolCallID,
		Tool:       a.ToolName,
		Title:      title,
		Args:       json.RawMessage(a.Args),
		Status:     a.Status,
		Result:     result,
	}
}
