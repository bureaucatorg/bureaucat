package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/assistant"
	"bereaucat/internal/store"
)

// AssistantHandler is the web adapter for the LLM assistant.
type AssistantHandler struct {
	store   store.Querier
	service *assistant.Service
}

func NewAssistantHandler(store store.Querier, service *assistant.Service) *AssistantHandler {
	return &AssistantHandler{store: store, service: service}
}

// AssistantStatusResponse tells the UI whether the assistant is available.
type AssistantStatusResponse struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// AssistantMessageRequest is a chat message from the user.
type AssistantMessageRequest struct {
	Text      string         `json:"text"`
	UploadIDs []uuid.UUID    `json:"upload_ids"`
	Meta      assistant.Meta `json:"meta"`
}

// AssistantActionRequest approves or rejects a pending write.
type AssistantActionRequest struct {
	Approve bool `json:"approve"`
}

// GetStatus returns whether the assistant is enabled.
//
//	@Summary		Assistant status
//	@Tags			Assistant
//	@Produce		json
//	@Success		200	{object}	AssistantStatusResponse
//	@Security		BearerAuth
//	@Router			/assistant/status [get]
func (h *AssistantHandler) GetStatus(c *echo.Context) error {
	cfg, err := assistant.Load(c.Request().Context(), h.store)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load assistant settings")
	}
	return c.JSON(http.StatusOK, AssistantStatusResponse{Enabled: cfg.Enabled, Provider: cfg.Provider, Model: cfg.Model})
}

// ListConversations returns the current user's conversations, newest first.
//
//	@Summary		List assistant conversations
//	@Tags			Assistant
//	@Produce		json
//	@Success		200	{array}	assistant.ConversationView
//	@Security		BearerAuth
//	@Router			/assistant/conversations [get]
func (h *AssistantHandler) ListConversations(c *echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	convs, err := h.service.ListConversations(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list conversations")
	}
	return c.JSON(http.StatusOK, convs)
}

// CreateConversation starts an empty conversation.
//
//	@Summary		Create assistant conversation
//	@Tags			Assistant
//	@Produce		json
//	@Success		201	{object}	assistant.ConversationView
//	@Security		BearerAuth
//	@Router			/assistant/conversations [post]
func (h *AssistantHandler) CreateConversation(c *echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	conv, err := h.service.CreateConversation(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create conversation")
	}
	return c.JSON(http.StatusCreated, conv)
}

// GetConversation returns a conversation's messages and actions.
//
//	@Summary		Get assistant conversation
//	@Tags			Assistant
//	@Produce		json
//	@Param			id	path		string	true	"Conversation ID"
//	@Success		200	{object}	assistant.ConversationDetail
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/assistant/conversations/{id} [get]
func (h *AssistantHandler) GetConversation(c *echo.Context) error {
	userID, convID, err := conversationParams(c)
	if err != nil {
		return err
	}
	detail, err := h.service.GetConversation(c.Request().Context(), userID, convID)
	if err != nil {
		return assistantHTTPError(err)
	}
	return c.JSON(http.StatusOK, detail)
}

// DeleteConversation deletes one of the user's conversations.
//
//	@Summary		Delete assistant conversation
//	@Tags			Assistant
//	@Param			id	path	string	true	"Conversation ID"
//	@Success		204
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/assistant/conversations/{id} [delete]
func (h *AssistantHandler) DeleteConversation(c *echo.Context) error {
	userID, convID, err := conversationParams(c)
	if err != nil {
		return err
	}
	if err := h.service.DeleteConversation(c.Request().Context(), userID, convID); err != nil {
		return assistantHTTPError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// AssistantRenameRequest sets a conversation's title.
type AssistantRenameRequest struct {
	Title string `json:"title"`
}

// RenameConversation changes the title of one of the user's conversations.
//
//	@Summary		Rename assistant conversation
//	@Tags			Assistant
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Conversation ID"
//	@Param			body	body		AssistantRenameRequest	true	"New title"
//	@Success		200		{object}	AssistantRenameRequest
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/assistant/conversations/{id} [patch]
func (h *AssistantHandler) RenameConversation(c *echo.Context) error {
	userID, convID, err := conversationParams(c)
	if err != nil {
		return err
	}
	var req AssistantRenameRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	title, err := h.service.RenameConversation(c.Request().Context(), userID, convID, req.Title)
	if errors.Is(err, assistant.ErrEmptyTitle) {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err != nil {
		return assistantHTTPError(err)
	}
	return c.JSON(http.StatusOK, AssistantRenameRequest{Title: title})
}

// AssistantToolOutputResponse is the full output of one tool call.
type AssistantToolOutputResponse struct {
	Output string `json:"output"`
}

// GetToolOutput returns the full output of a tool call in one of the user's conversations.
//
//	@Summary		Get assistant tool call output
//	@Tags			Assistant
//	@Produce		json
//	@Param			id		path		string	true	"Conversation ID"
//	@Param			callId	path		string	true	"Tool call ID"
//	@Success		200		{object}	AssistantToolOutputResponse
//	@Failure		404		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/assistant/conversations/{id}/tool-calls/{callId}/output [get]
func (h *AssistantHandler) GetToolOutput(c *echo.Context) error {
	userID, convID, err := conversationParams(c)
	if err != nil {
		return err
	}
	out, err := h.service.ToolOutput(c.Request().Context(), userID, convID, c.Param("callId"))
	if errors.Is(err, assistant.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "tool output not found")
	}
	if err != nil {
		return assistantHTTPError(err)
	}
	return c.JSON(http.StatusOK, AssistantToolOutputResponse{Output: out})
}

// SendMessage posts a user message and streams agent events as server-sent events.
//
//	@Summary		Send assistant message
//	@Description	Streams events (step, message, pending, action, error, done) as text/event-stream.
//	@Tags			Assistant
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			id		path	string					true	"Conversation ID"
//	@Param			body	body	AssistantMessageRequest	true	"Message"
//	@Success		200
//	@Security		BearerAuth
//	@Router			/assistant/conversations/{id}/messages [post]
func (h *AssistantHandler) SendMessage(c *echo.Context) error {
	userID, convID, err := conversationParams(c)
	if err != nil {
		return err
	}
	var req AssistantMessageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	text := strings.TrimSpace(req.Text)
	if (text == "" && len(req.UploadIDs) == 0) || len(text) > 20000 {
		return echo.NewHTTPError(http.StatusBadRequest, "message must be 1-20000 characters or include a file")
	}
	return h.stream(c, func(ctx context.Context, emit assistant.Emitter) error {
		return h.service.Send(ctx, userID, convID, text, req.UploadIDs, assistant.Meta{Page: req.Meta.Page}, emit)
	})
}

// ResolveAction approves or rejects a pending write and streams the resumed run.
//
//	@Summary		Resolve assistant action
//	@Tags			Assistant
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			id			path	string					true	"Conversation ID"
//	@Param			actionId	path	string					true	"Action ID"
//	@Param			body		body	AssistantActionRequest	true	"Decision"
//	@Success		200
//	@Security		BearerAuth
//	@Router			/assistant/conversations/{id}/actions/{actionId} [post]
func (h *AssistantHandler) ResolveAction(c *echo.Context) error {
	userID, convID, err := conversationParams(c)
	if err != nil {
		return err
	}
	actionID, err := uuid.Parse(c.Param("actionId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid action ID")
	}
	var req AssistantActionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	return h.stream(c, func(ctx context.Context, emit assistant.Emitter) error {
		return h.service.Resolve(ctx, userID, convID, actionID, req.Approve, emit)
	})
}

type sseEmitter struct {
	w  http.ResponseWriter
	rc *http.ResponseController
}

func (e *sseEmitter) Emit(ev assistant.Event) {
	data, err := json.Marshal(ev)
	if err != nil {
		return
	}
	fmt.Fprintf(e.w, "data: %s\n\n", data)
	_ = e.rc.Flush()
}

func (h *AssistantHandler) stream(c *echo.Context, run func(context.Context, assistant.Emitter) error) error {
	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	emit := &sseEmitter{w: w, rc: http.NewResponseController(w)}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Minute)
	defer cancel()

	if err := run(ctx, emit); err != nil {
		emit.Emit(assistant.Event{Type: "error", Text: assistant.ErrorText(err)})
	}
	emit.Emit(assistant.Event{Type: "done"})
	return nil
}

func conversationParams(c *echo.Context) (uuid.UUID, uuid.UUID, error) {
	userID, err := getUserID(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "invalid conversation ID")
	}
	return userID, convID, nil
}

func assistantHTTPError(err error) error {
	if errors.Is(err, assistant.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "conversation not found")
	}
	return echo.NewHTTPError(http.StatusInternalServerError, "assistant request failed")
}

// --- Admin settings ---

// GetAssistantSettings returns the assistant config with the API key masked (admin only).
//
//	@Summary		Get assistant settings
//	@Tags			Admin - Settings
//	@Produce		json
//	@Success		200	{object}	assistant.Settings
//	@Security		BearerAuth
//	@Router			/admin/settings/assistant [get]
func (h *AssistantHandler) GetAssistantSettings(c *echo.Context) error {
	cfg, err := assistant.Load(c.Request().Context(), h.store)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load assistant settings")
	}
	cfg.APIKey = maskSecret(cfg.APIKey)
	return c.JSON(http.StatusOK, cfg)
}

// UpdateAssistantSettings saves the assistant config (admin only). A masked key is preserved.
//
//	@Summary		Update assistant settings
//	@Tags			Admin - Settings
//	@Accept			json
//	@Produce		json
//	@Param			body	body		assistant.Settings	true	"Assistant configuration"
//	@Success		200		{object}	assistant.Settings
//	@Failure		400		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/settings/assistant [put]
func (h *AssistantHandler) UpdateAssistantSettings(c *echo.Context) error {
	var req assistant.Settings
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()

	existing, err := assistant.Load(ctx, h.store)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load assistant settings")
	}
	if isSecretMasked(req.APIKey) && req.Provider == existing.Provider {
		req.APIKey = existing.APIKey
	}
	if req.Provider == assistant.ProviderBedrock {
		req.APIKey = ""
	}
	req.Model = strings.TrimSpace(req.Model)
	req.AWSRegion = strings.TrimSpace(req.AWSRegion)

	if req.Enabled {
		if err := req.Validate(); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid assistant settings: "+err.Error())
		}
	}

	value, err := json.Marshal(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to marshal settings")
	}
	if _, err := h.store.UpsertSetting(ctx, store.UpsertSettingParams{Key: assistant.SettingKey, Value: value}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update assistant settings")
	}

	req.APIKey = maskSecret(req.APIKey)
	return c.JSON(http.StatusOK, req)
}

// TestAssistantSettings sends a one-line prompt to the saved model (admin only).
//
//	@Summary		Test assistant settings
//	@Tags			Admin - Settings
//	@Produce		json
//	@Success		200	{object}	MessageResponse
//	@Failure		400	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/settings/assistant/test [post]
func (h *AssistantHandler) TestAssistantSettings(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), time.Minute)
	defer cancel()

	cfg, err := assistant.Load(ctx, h.store)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load assistant settings")
	}
	if err := cfg.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "assistant is not configured: "+err.Error())
	}
	provider, err := assistant.NewProvider(ctx, cfg)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to create provider: "+err.Error())
	}
	resp, err := provider.Chat(ctx, assistant.ChatRequest{
		Model:    cfg.Model,
		System:   "Reply with the single word: pong",
		Messages: []assistant.Message{{Role: assistant.RoleUser, Text: "ping"}},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "model request failed: "+err.Error())
	}
	reply := strings.TrimSpace(resp.Text)
	if len(reply) > 100 {
		reply = reply[:100]
	}
	return c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("%s replied: %s", cfg.Model, reply)})
}
