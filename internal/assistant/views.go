package assistant

import (
	"context"
	"encoding/json"
	"errors"
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
	ID          uuid.UUID    `json:"id"`
	Role        string       `json:"role"`
	Content     string       `json:"content"`
	ToolCalls   []ToolCall   `json:"tool_calls"`
	Attachments []Attachment `json:"attachments"`
	Meta        Meta         `json:"meta"`
	CreatedAt   time.Time    `json:"created_at"`
}

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
	for _, r := range rows {
		if r.Role == RoleTool {
			continue
		}
		calls := []ToolCall{}
		_ = json.Unmarshal(r.ToolCalls, &calls)
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
