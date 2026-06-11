package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/auth"
	"bereaucat/internal/store"
)

// NotificationHandler serves per-user in-app notifications.
type NotificationHandler struct {
	store store.Querier
}

// NewNotificationHandler creates a new notification handler.
func NewNotificationHandler(s store.Querier) *NotificationHandler {
	return &NotificationHandler{store: s}
}

// NotificationResponse is a single notification with task/project/actor context.
type NotificationResponse struct {
	ID           uuid.UUID `json:"id"`
	TaskID       uuid.UUID `json:"task_id"`
	ActivityType string    `json:"activity_type"`
	ActorID      uuid.UUID `json:"actor_id"`
	CommentID    *string   `json:"comment_id,omitempty"`
	Username     string    `json:"username"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	EventCount   int32     `json:"event_count"`
	IsRead       bool      `json:"is_read"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	TaskNumber   int32     `json:"task_number"`
	ProjectKey   string    `json:"project_key"`
	TaskTitle    string    `json:"task_title"`
}

// ListNotifications returns the current user's notifications, paginated.
func (h *NotificationHandler) ListNotifications(c *echo.Context) error {
	userID, err := uuid.Parse(c.Request().Header.Get(auth.HeaderUserID))
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	ctx := c.Request().Context()

	rows, err := h.store.ListNotifications(ctx, store.ListNotificationsParams{
		RecipientID: userID,
		Limit:       int32(perPage),
		Offset:      int32(offset),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list notifications")
	}

	total, err := h.store.CountNotifications(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count notifications")
	}

	unread, err := h.store.CountUnreadNotifications(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count unread notifications")
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	items := make([]NotificationResponse, len(rows))
	for i, n := range rows {
		var avatarURL *string
		if n.AvatarUrl.Valid {
			avatarURL = &n.AvatarUrl.String
		}
		var commentID *string
		if n.CommentID.Valid {
			s := uuid.UUID(n.CommentID.Bytes).String()
			commentID = &s
		}
		items[i] = NotificationResponse{
			ID:           n.ID,
			TaskID:       n.TaskID,
			ActivityType: n.ActivityType,
			ActorID:      n.ActorID,
			CommentID:    commentID,
			Username:     n.Username,
			FirstName:    n.FirstName,
			LastName:     n.LastName,
			AvatarURL:    avatarURL,
			EventCount:   n.EventCount,
			IsRead:       n.ReadAt.Valid,
			CreatedAt:    n.CreatedAt.Time,
			UpdatedAt:    n.UpdatedAt.Time,
			TaskNumber:   n.TaskNumber,
			ProjectKey:   n.ProjectKey,
			TaskTitle:    n.TaskTitle,
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"activities":   items,
		"unread_count": unread,
		"total":        total,
		"page":         page,
		"per_page":     perPage,
		"total_pages":  totalPages,
	})
}

// GetUnreadCount returns the number of unread notifications for the bell badge.
func (h *NotificationHandler) GetUnreadCount(c *echo.Context) error {
	userID, err := uuid.Parse(c.Request().Header.Get(auth.HeaderUserID))
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user")
	}

	count, err := h.store.CountUnreadNotifications(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count unread notifications")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"count": count})
}

// MarkRead marks a single notification as read.
func (h *NotificationHandler) MarkRead(c *echo.Context) error {
	userID, err := uuid.Parse(c.Request().Header.Get(auth.HeaderUserID))
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user")
	}

	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid notification id")
	}

	if err := h.store.MarkNotificationRead(c.Request().Context(), store.MarkNotificationReadParams{
		ID:          notificationID,
		RecipientID: userID,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to mark notification read")
	}

	return c.NoContent(http.StatusNoContent)
}

// MarkAllRead marks all of the current user's notifications as read.
func (h *NotificationHandler) MarkAllRead(c *echo.Context) error {
	userID, err := uuid.Parse(c.Request().Header.Get(auth.HeaderUserID))
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user")
	}

	if err := h.store.MarkAllNotificationsRead(c.Request().Context(), userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to mark notifications read")
	}

	return c.NoContent(http.StatusNoContent)
}
