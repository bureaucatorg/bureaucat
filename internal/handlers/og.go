package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v5"

	"bereaucat/internal/ogimage"
	"bereaucat/internal/store"
)

// OGHandler handles Open Graph image generation.
type OGHandler struct {
	store store.Querier
}

// NewOGHandler creates a new OG image handler.
func NewOGHandler(store store.Querier) *OGHandler {
	return &OGHandler{store: store}
}

// OGImage generates a dynamic PNG Open Graph image with branding.
func (h *OGHandler) OGImage(c *echo.Context) error {
	ctx := c.Request().Context()

	appName := "Bureaucat"
	setting, err := h.store.GetSetting(ctx, "branding")
	if err == nil {
		var branding BrandingSettings
		if err := json.Unmarshal(setting.Value, &branding); err == nil {
			if branding.Enabled && branding.AppName != "" {
				appName = branding.AppName
			}
		}
	}

	data, err := ogimage.Render(appName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to render og image"})
	}

	c.Response().Header().Set("Cache-Control", "public, max-age=3600")
	return c.Blob(http.StatusOK, "image/png", data)
}
