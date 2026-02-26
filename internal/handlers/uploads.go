package handlers

import (
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/auth"
	"bereaucat/internal/store"
	"bereaucat/internal/uploads"
)

// UploadHandler handles file upload endpoints.
type UploadHandler struct {
	store         store.Querier
	uploadService *uploads.Service
}

// NewUploadHandler creates a new upload handler.
func NewUploadHandler(store store.Querier, uploadService *uploads.Service) *UploadHandler {
	return &UploadHandler{
		store:         store,
		uploadService: uploadService,
	}
}

// UploadResponse represents the response after a successful upload.
type UploadResponse struct {
	ID       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	MimeType string    `json:"mime_type"`
	Size     int64     `json:"size_bytes"`
	URL      string    `json:"url"`
}

// Upload handles file uploads.
//
//	@Summary		Upload file
//	@Description	Upload an image file (jpeg, png, gif, webp). Max 5MB.
//	@Tags			Uploads
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"File to upload"
//	@Success		201		{object}	UploadResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/uploads [post]
func (h *UploadHandler) Upload(c *echo.Context) error {
	// Get current user ID from auth middleware
	userIDStr := c.Request().Header.Get(auth.HeaderUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	// Get file from request
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "file is required")
	}

	// Save file to disk
	result, err := h.uploadService.SaveFile(file)
	if err != nil {
		if err == uploads.ErrFileTooLarge {
			return echo.NewHTTPError(http.StatusBadRequest, "file exceeds maximum size (5MB)")
		}
		if err == uploads.ErrInvalidMimeType {
			return echo.NewHTTPError(http.StatusBadRequest, "file type not allowed (only images: jpeg, png, gif, webp)")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save file")
	}

	ctx := c.Request().Context()

	// Save upload record to database
	upload, err := h.store.CreateUpload(ctx, store.CreateUploadParams{
		Filename:   file.Filename,
		StoredName: result.StoredName,
		MimeType:   result.MimeType,
		SizeBytes:  result.SizeBytes,
		UploadedBy: userID,
	})
	if err != nil {
		// Clean up the file on database error
		h.uploadService.DeleteFile(result.StoredName)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save upload record")
	}

	return c.JSON(http.StatusCreated, UploadResponse{
		ID:       upload.ID,
		Filename: upload.Filename,
		MimeType: upload.MimeType,
		Size:     upload.SizeBytes,
		URL:      "/api/v1/uploads/" + upload.ID.String(),
	})
}

// Serve serves an uploaded file.
//
//	@Summary		Serve uploaded file
//	@Description	Serve a previously uploaded file by its ID.
//	@Tags			Uploads
//	@Produce		octet-stream
//	@Param			id	path	string	true	"Upload ID"
//	@Success		200
//	@Failure		400	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/uploads/{id} [get]
func (h *UploadHandler) Serve(c *echo.Context) error {
	uploadIDStr := c.Param("id")
	uploadID, err := uuid.Parse(uploadIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid upload ID")
	}

	ctx := c.Request().Context()

	// Get upload record
	upload, err := h.store.GetUploadByID(ctx, uploadID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "upload not found")
	}

	// Get file path
	filePath := h.uploadService.GetFilePath(upload.StoredName)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return echo.NewHTTPError(http.StatusNotFound, "file not found")
	}

	// Set content type header
	c.Response().Header().Set("Content-Type", upload.MimeType)

	// Set cache headers (1 hour)
	c.Response().Header().Set("Cache-Control", "public, max-age=3600")

	return c.File(filePath)
}
