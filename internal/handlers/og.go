package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"bereaucat/internal/branding"
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
//
//	@Summary		Generate OG image
//	@Description	Generates a dynamic PNG Open Graph image with the app's branding.
//	@Tags			Misc
//	@Produce		image/png
//	@Success		200
//	@Failure		500	{object}	ErrorResponse
//	@Router			/og-image [get]
func (h *OGHandler) OGImage(c *echo.Context) error {
	brand := branding.Load(c.Request().Context(), h.store)

	data, err := ogimage.Render(brand.Name, brand.Custom)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to render og image"})
	}

	c.Response().Header().Set("Cache-Control", "public, max-age=3600")
	return c.Blob(http.StatusOK, "image/png", data)
}
