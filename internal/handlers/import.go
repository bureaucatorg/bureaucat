package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/auth"
	"bereaucat/internal/planeimport"
)

// ImportHandler handles data import endpoints.
type ImportHandler struct {
	pool *pgxpool.Pool
}

// NewImportHandler creates a new import handler.
func NewImportHandler(pool *pgxpool.Pool) *ImportHandler {
	return &ImportHandler{pool: pool}
}

// ImportPlane handles uploading and importing a Plane.so SQL dump.
//
//	@Summary		Import from Plane.so
//	@Description	Upload and import a Plane.so SQL dump file. Max 500MB. Requires admin role.
//	@Tags			Admin - Import
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"SQL dump file"
//	@Success		200		{object}	MessageResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/import/plane [post]
func (h *ImportHandler) ImportPlane(c *echo.Context) error {
	// Get admin user ID.
	adminUserIDStr := c.Request().Header.Get(auth.HeaderUserID)
	adminUserID, err := uuid.Parse(adminUserIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	// Get uploaded file.
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "file is required")
	}

	// Limit to 500MB.
	if file.Size > 500*1024*1024 {
		return echo.NewHTTPError(http.StatusBadRequest, "file too large (max 500MB)")
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to open uploaded file")
	}
	defer src.Close()

	// Parse the SQL dump.
	dump, err := planeimport.Parse(src)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse SQL dump: "+err.Error())
	}

	// Run the import.
	ctx := c.Request().Context()
	result, err := planeimport.Import(ctx, h.pool, adminUserID, dump)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "import failed: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Import completed successfully",
		"summary": result,
	})
}
