package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/activity"
	"bereaucat/internal/auth"
	"bereaucat/internal/notifier"
	"bereaucat/internal/store"
)

// CommentHandler handles comment and activity log endpoints.
type CommentHandler struct {
	store               store.Querier
	activityService     *activity.Service
	notificationService *notifier.Service
}

// NewCommentHandler creates a new comment handler.
func NewCommentHandler(store store.Querier, activityService *activity.Service, notificationService *notifier.Service) *CommentHandler {
	return &CommentHandler{
		store:               store,
		activityService:     activityService,
		notificationService: notificationService,
	}
}

// CommentResponse represents a comment in API responses.
type CommentResponse struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"task_id"`
	Content   string    `json:"content"`
	Version   int       `json:"version"`
	CreatedBy uuid.UUID `json:"created_by"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCommentRequest represents the request to create a comment.
type CreateCommentRequest struct {
	Content string `json:"content"`
}

// UpdateCommentRequest represents the request to update a comment.
type UpdateCommentRequest struct {
	Content string `json:"content"`
}

// ActivityLogResponse represents an activity log entry in API responses.
type ActivityLogResponse struct {
	ID           uuid.UUID   `json:"id"`
	TaskID       uuid.UUID   `json:"task_id"`
	ActivityType string      `json:"activity_type"`
	ActorID      uuid.UUID   `json:"actor_id"`
	Username     string      `json:"username"`
	FirstName    string      `json:"first_name"`
	LastName     string      `json:"last_name"`
	FieldName    *string     `json:"field_name,omitempty"`
	OldValue     interface{} `json:"old_value,omitempty"`
	NewValue     interface{} `json:"new_value,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

// ListComments returns comments for a task.
//
//	@Summary		List comments
//	@Description	Returns all comments for a task.
//	@Tags			Comments
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			taskNum		path		int		true	"Task number"
//	@Success		200			{array}		CommentResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum}/comments [get]
func (h *CommentHandler) ListComments(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	taskNumStr := c.Param("taskNum")
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task number")
	}

	ctx := c.Request().Context()

	// Get task
	task, err := h.store.GetTaskByProjectAndNumber(ctx, store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	// Get comments
	comments, err := h.store.ListTaskComments(ctx, task.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list comments")
	}

	commentResponses := make([]CommentResponse, len(comments))
	for i, cm := range comments {
		commentResponses[i] = CommentResponse{
			ID:        cm.ID,
			TaskID:    cm.TaskID,
			Content:   cm.Content,
			Version:   int(cm.Version),
			CreatedBy: cm.CreatedBy,
			Username:  cm.Username,
			FirstName: cm.FirstName,
			LastName:  cm.LastName,
			CreatedAt: cm.CreatedAt.Time,
			UpdatedAt: cm.UpdatedAt.Time,
		}
	}

	return c.JSON(http.StatusOK, commentResponses)
}

// CreateComment creates a new comment.
//
//	@Summary		Create comment
//	@Description	Add a comment to a task.
//	@Tags			Comments
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string					true	"Project key"
//	@Param			taskNum		path		int						true	"Task number"
//	@Param			body		body		CreateCommentRequest	true	"Comment content"
//	@Success		201			{object}	CommentResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum}/comments [post]
func (h *CommentHandler) CreateComment(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	userIDStr := c.Request().Header.Get(auth.HeaderUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	taskNumStr := c.Param("taskNum")
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task number")
	}

	var req CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Content == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "content is required")
	}

	ctx := c.Request().Context()

	// Get task
	task, err := h.store.GetTaskByProjectAndNumber(ctx, store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	// Create comment
	comment, err := h.store.CreateComment(ctx, store.CreateCommentParams{
		TaskID:    task.ID,
		Content:   req.Content,
		CreatedBy: userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create comment")
	}

	// Log activity
	h.activityService.LogActivity(ctx, activity.LogActivityParams{
		TaskID:       task.ID,
		ActivityType: activity.CommentCreated,
		ActorID:      userID,
		NewValue: map[string]interface{}{
			"comment_id": comment.ID.String(),
			"content":    comment.Content,
		},
	})

	// Send mention notifications
	if h.notificationService != nil {
		mentionedIDs := notifier.ParseMentions(req.Content)
		if len(mentionedIDs) > 0 {
			actorUser, _ := h.store.GetUserByID(ctx, userID)
			actorName := actorUser.FirstName + " " + actorUser.LastName
			if actorName == " " {
				actorName = actorUser.Username
			}
			projectKey := c.Request().Header.Get(auth.HeaderProjectKey)
			for _, mentionedID := range mentionedIDs {
				if mentionedID == userID {
					continue
				}
				h.notificationService.Notify(ctx, notifier.Notification{
					Event:       notifier.EventMentioned,
					RecipientID: mentionedID,
					ActorName:   actorName,
					ProjectKey:  projectKey,
					TaskNumber:  taskNum,
					TaskTitle:   task.Title,
				})
			}
		}
	}

	// Get full comment with user info
	fullComment, err := h.store.GetCommentByID(ctx, comment.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get comment")
	}

	return c.JSON(http.StatusCreated, CommentResponse{
		ID:        fullComment.ID,
		TaskID:    fullComment.TaskID,
		Content:   fullComment.Content,
		Version:   int(fullComment.Version),
		CreatedBy: fullComment.CreatedBy,
		Username:  fullComment.Username,
		FirstName: fullComment.FirstName,
		LastName:  fullComment.LastName,
		CreatedAt: fullComment.CreatedAt.Time,
		UpdatedAt: fullComment.UpdatedAt.Time,
	})
}

// UpdateComment updates a comment.
//
//	@Summary		Update comment
//	@Description	Update a comment. Can only edit your own comments.
//	@Tags			Comments
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string					true	"Project key"
//	@Param			taskNum		path		int						true	"Task number"
//	@Param			commentId	path		string					true	"Comment ID"
//	@Param			body		body		UpdateCommentRequest	true	"Updated content"
//	@Success		200			{object}	CommentResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		403			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum}/comments/{commentId} [patch]
func (h *CommentHandler) UpdateComment(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	userIDStr := c.Request().Header.Get(auth.HeaderUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	taskNumStr := c.Param("taskNum")
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task number")
	}

	commentIDStr := c.Param("commentId")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid comment ID")
	}

	var req UpdateCommentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Content == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "content is required")
	}

	ctx := c.Request().Context()

	// Get task
	task, err := h.store.GetTaskByProjectAndNumber(ctx, store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	// Get comment
	oldComment, err := h.store.GetCommentByID(ctx, commentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "comment not found")
	}

	// Check ownership
	if oldComment.CreatedBy != userID {
		return echo.NewHTTPError(http.StatusForbidden, "can only edit your own comments")
	}

	// Update comment
	comment, err := h.store.UpdateComment(ctx, store.UpdateCommentParams{
		ID:      commentID,
		Content: req.Content,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update comment")
	}

	// Log activity
	h.activityService.LogActivity(ctx, activity.LogActivityParams{
		TaskID:       task.ID,
		ActivityType: activity.CommentUpdated,
		ActorID:      userID,
		FieldName:    activity.StringPtr("content"),
		OldValue: map[string]interface{}{
			"comment_id": commentID.String(),
			"content":    oldComment.Content,
			"version":    oldComment.Version,
		},
		NewValue: map[string]interface{}{
			"comment_id": commentID.String(),
			"content":    comment.Content,
			"version":    comment.Version,
		},
	})

	// Get updated comment with user info
	fullComment, err := h.store.GetCommentByID(ctx, comment.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get comment")
	}

	return c.JSON(http.StatusOK, CommentResponse{
		ID:        fullComment.ID,
		TaskID:    fullComment.TaskID,
		Content:   fullComment.Content,
		Version:   int(fullComment.Version),
		CreatedBy: fullComment.CreatedBy,
		Username:  fullComment.Username,
		FirstName: fullComment.FirstName,
		LastName:  fullComment.LastName,
		CreatedAt: fullComment.CreatedAt.Time,
		UpdatedAt: fullComment.UpdatedAt.Time,
	})
}

// DeleteComment soft deletes a comment.
//
//	@Summary		Delete comment
//	@Description	Soft-delete a comment. Can delete own comments or any comment as project admin.
//	@Tags			Comments
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			taskNum		path		int		true	"Task number"
//	@Param			commentId	path		string	true	"Comment ID"
//	@Success		200			{object}	MessageResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		403			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum}/comments/{commentId} [delete]
func (h *CommentHandler) DeleteComment(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	userIDStr := c.Request().Header.Get(auth.HeaderUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	projectRole := c.Request().Header.Get(auth.HeaderProjectRole)

	taskNumStr := c.Param("taskNum")
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task number")
	}

	commentIDStr := c.Param("commentId")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid comment ID")
	}

	ctx := c.Request().Context()

	// Get task
	task, err := h.store.GetTaskByProjectAndNumber(ctx, store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	// Get comment
	comment, err := h.store.GetCommentByID(ctx, commentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "comment not found")
	}

	// Check ownership or admin
	if comment.CreatedBy != userID && projectRole != "admin" {
		return echo.NewHTTPError(http.StatusForbidden, "can only delete your own comments")
	}

	// Log activity
	h.activityService.LogActivity(ctx, activity.LogActivityParams{
		TaskID:       task.ID,
		ActivityType: activity.CommentDeleted,
		ActorID:      userID,
		OldValue: map[string]interface{}{
			"comment_id": commentID.String(),
			"content":    comment.Content,
			"version":    comment.Version,
		},
	})

	// Soft delete
	err = h.store.SoftDeleteComment(ctx, commentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete comment")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "comment deleted"})
}

// GetActivity returns activity log for a task.
//
//	@Summary		Get activity log
//	@Description	Returns the activity log for a task including all changes and comments.
//	@Tags			Activity
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			taskNum		path		int		true	"Task number"
//	@Success		200			{array}		ActivityLogResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum}/activity [get]
func (h *CommentHandler) GetActivity(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	taskNumStr := c.Param("taskNum")
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task number")
	}

	ctx := c.Request().Context()

	// Get task
	task, err := h.store.GetTaskByProjectAndNumber(ctx, store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	// Get activity log
	activities, err := h.store.ListTaskActivity(ctx, task.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list activity")
	}

	activityResponses := make([]ActivityLogResponse, len(activities))
	for i, a := range activities {
		var fieldName *string
		if a.FieldName.Valid {
			fieldName = &a.FieldName.String
		}
		activityResponses[i] = ActivityLogResponse{
			ID:           a.ID,
			TaskID:       a.TaskID,
			ActivityType: a.ActivityType,
			ActorID:      a.ActorID,
			Username:     a.Username,
			FirstName:    a.FirstName,
			LastName:     a.LastName,
			FieldName:    fieldName,
			OldValue:     parseJSONB(a.OldValue),
			NewValue:     parseJSONB(a.NewValue),
			CreatedAt:    a.CreatedAt.Time,
		}
	}

	return c.JSON(http.StatusOK, activityResponses)
}

// VerifyActivityResponse represents the response for activity chain verification.
type VerifyActivityResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// VerifyActivity verifies the integrity of the activity log chain.
//
//	@Summary		Verify activity log
//	@Description	Verify the integrity of the activity log chain for tampering detection.
//	@Tags			Activity
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			taskNum		path		int		true	"Task number"
//	@Success		200			{object}	VerifyActivityResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum}/activity/verify [get]
func (h *CommentHandler) VerifyActivity(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	taskNumStr := c.Param("taskNum")
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task number")
	}

	ctx := c.Request().Context()

	// Get task
	task, err := h.store.GetTaskByProjectAndNumber(ctx, store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	// Verify chain
	valid, err := h.activityService.VerifyChain(ctx, task.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify activity chain")
	}

	message := "Activity log chain is valid"
	if !valid {
		message = "Activity log chain integrity compromised - tampering detected"
	}

	return c.JSON(http.StatusOK, VerifyActivityResponse{
		Valid:   valid,
		Message: message,
	})
}

// parseJSONB safely parses JSONB data
func parseJSONB(data []byte) interface{} {
	if data == nil {
		return nil
	}
	// Return as raw JSON for the client to parse
	return string(data)
}

// Helper to get task by project and number (used internally)
func (h *CommentHandler) getTask(ctx context.Context, projectID uuid.UUID, taskNum int) (*store.GetTaskByProjectAndNumberRow, error) {
	task, err := h.store.GetTaskByProjectAndNumber(ctx, store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return nil, err
	}
	return &task, nil
}
