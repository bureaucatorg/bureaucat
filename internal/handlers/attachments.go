package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/auth"
	"bereaucat/internal/store"
	"bereaucat/internal/uploads"
)

// AttachmentHandler handles file attachment endpoints.
type AttachmentHandler struct {
	store         store.Querier
	uploadService *uploads.Service
}

// NewAttachmentHandler creates a new attachment handler.
func NewAttachmentHandler(store store.Querier, uploadService *uploads.Service) *AttachmentHandler {
	return &AttachmentHandler{
		store:         store,
		uploadService: uploadService,
	}
}

// AttachmentResponse represents an attachment in API responses.
type AttachmentResponse struct {
	ID        uuid.UUID `json:"id"`
	UploadID  uuid.UUID `json:"upload_id"`
	Filename  string    `json:"filename"`
	MimeType  string    `json:"mime_type"`
	SizeBytes int64     `json:"size_bytes"`
	URL       string    `json:"url"`
	CreatedBy uuid.UUID `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// AttachRequest represents the request to attach a file.
type AttachRequest struct {
	UploadID uuid.UUID `json:"upload_id"`
}

// AttachToTask attaches an uploaded file to a task.
func (h *AttachmentHandler) AttachToTask(c *echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	task, err := h.resolveTask(c)
	if err != nil {
		return err
	}

	var req AttachRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.UploadID == uuid.Nil {
		return echo.NewHTTPError(http.StatusBadRequest, "upload_id is required")
	}

	ctx := c.Request().Context()

	attachment, err := h.store.CreateAttachment(ctx, store.CreateAttachmentParams{
		UploadID:   req.UploadID,
		EntityType: "task",
		EntityID:   task.ID,
		CreatedBy:  userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create attachment")
	}

	// Fetch the upload info for response
	upload, err := h.store.GetUploadByID(ctx, req.UploadID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get upload info")
	}

	return c.JSON(http.StatusCreated, AttachmentResponse{
		ID:        attachment.ID,
		UploadID:  attachment.UploadID,
		Filename:  upload.Filename,
		MimeType:  upload.MimeType,
		SizeBytes: upload.SizeBytes,
		URL:       "/api/v1/uploads/" + upload.ID.String(),
		CreatedBy: attachment.CreatedBy,
		CreatedAt: attachment.CreatedAt.Time,
	})
}

// ListTaskAttachments lists all attachments for a task.
func (h *AttachmentHandler) ListTaskAttachments(c *echo.Context) error {
	task, err := h.resolveTask(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	attachments, err := h.store.ListAttachmentsByEntity(ctx, store.ListAttachmentsByEntityParams{
		EntityType: "task",
		EntityID:   task.ID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list attachments")
	}

	resp := make([]AttachmentResponse, len(attachments))
	for i, a := range attachments {
		resp[i] = AttachmentResponse{
			ID:        a.ID,
			UploadID:  a.UploadID,
			Filename:  a.Filename,
			MimeType:  a.MimeType,
			SizeBytes: a.SizeBytes,
			URL:       "/api/v1/uploads/" + a.UploadID.String(),
			CreatedBy: a.CreatedBy,
			CreatedAt: a.CreatedAt.Time,
		}
	}

	return c.JSON(http.StatusOK, resp)
}

// DeleteTaskAttachment deletes an attachment from a task.
func (h *AttachmentHandler) DeleteTaskAttachment(c *echo.Context) error {
	attachmentIDStr := c.Param("attachmentId")
	attachmentID, err := uuid.Parse(attachmentIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid attachment ID")
	}

	ctx := c.Request().Context()

	if err := h.store.DeleteAttachment(ctx, attachmentID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete attachment")
	}

	return c.NoContent(http.StatusNoContent)
}

// AttachToComment attaches an uploaded file to a comment.
func (h *AttachmentHandler) AttachToComment(c *echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	commentIDStr := c.Param("commentId")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid comment ID")
	}

	var req AttachRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.UploadID == uuid.Nil {
		return echo.NewHTTPError(http.StatusBadRequest, "upload_id is required")
	}

	ctx := c.Request().Context()

	attachment, err := h.store.CreateAttachment(ctx, store.CreateAttachmentParams{
		UploadID:   req.UploadID,
		EntityType: "comment",
		EntityID:   commentID,
		CreatedBy:  userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create attachment")
	}

	upload, err := h.store.GetUploadByID(ctx, req.UploadID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get upload info")
	}

	return c.JSON(http.StatusCreated, AttachmentResponse{
		ID:        attachment.ID,
		UploadID:  attachment.UploadID,
		Filename:  upload.Filename,
		MimeType:  upload.MimeType,
		SizeBytes: upload.SizeBytes,
		URL:       "/api/v1/uploads/" + upload.ID.String(),
		CreatedBy: attachment.CreatedBy,
		CreatedAt: attachment.CreatedAt.Time,
	})
}

// ListCommentAttachments lists all attachments for a comment.
func (h *AttachmentHandler) ListCommentAttachments(c *echo.Context) error {
	commentIDStr := c.Param("commentId")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid comment ID")
	}

	ctx := c.Request().Context()

	attachments, err := h.store.ListAttachmentsByEntity(ctx, store.ListAttachmentsByEntityParams{
		EntityType: "comment",
		EntityID:   commentID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list attachments")
	}

	resp := make([]AttachmentResponse, len(attachments))
	for i, a := range attachments {
		resp[i] = AttachmentResponse{
			ID:        a.ID,
			UploadID:  a.UploadID,
			Filename:  a.Filename,
			MimeType:  a.MimeType,
			SizeBytes: a.SizeBytes,
			URL:       "/api/v1/uploads/" + a.UploadID.String(),
			CreatedBy: a.CreatedBy,
			CreatedAt: a.CreatedAt.Time,
		}
	}

	return c.JSON(http.StatusOK, resp)
}

// DeleteCommentAttachment deletes an attachment from a comment.
func (h *AttachmentHandler) DeleteCommentAttachment(c *echo.Context) error {
	attachmentIDStr := c.Param("attachmentId")
	attachmentID, err := uuid.Parse(attachmentIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid attachment ID")
	}

	ctx := c.Request().Context()

	if err := h.store.DeleteAttachment(ctx, attachmentID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete attachment")
	}

	return c.NoContent(http.StatusNoContent)
}

// resolveTask looks up a task by project key and task number from route params.
func (h *AttachmentHandler) resolveTask(c *echo.Context) (store.GetTaskByProjectAndNumberRow, error) {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return store.GetTaskByProjectAndNumberRow{}, echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	taskNumStr := c.Param("taskNum")
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		return store.GetTaskByProjectAndNumberRow{}, echo.NewHTTPError(http.StatusBadRequest, "invalid task number")
	}

	ctx := c.Request().Context()
	task, err := h.store.GetTaskByProjectAndNumber(ctx, store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return store.GetTaskByProjectAndNumberRow{}, echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	return task, nil
}

func getUserID(c *echo.Context) (uuid.UUID, error) {
	userIDStr := c.Request().Header.Get(auth.HeaderUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}
	return userID, nil
}
