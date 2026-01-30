package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"

	"github.com/labstack/echo/v5"

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

// OGImage generates a dynamic SVG Open Graph image with branding.
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

	safeName := html.EscapeString(appName)
	safeLetter := html.EscapeString(string([]rune(appName)[0]))

	svg := fmt.Sprintf(`<svg width="1200" height="630" viewBox="0 0 1200 630" fill="none" xmlns="http://www.w3.org/2000/svg">
  <rect width="1200" height="630" fill="#09090B"/>
  <defs>
    <pattern id="grid" width="80" height="80" patternUnits="userSpaceOnUse">
      <path d="M 80 0 L 0 0 0 80" fill="none" stroke="#ffffff" stroke-opacity="0.04" stroke-width="1"/>
    </pattern>
  </defs>
  <rect width="1200" height="630" fill="url(#grid)"/>
  <circle cx="1050" cy="150" r="350" fill="#F59E0B" fill-opacity="0.06"/>
  <circle cx="200" cy="550" r="250" fill="#F59E0B" fill-opacity="0.06"/>
  <rect x="80" y="140" width="56" height="56" rx="12" fill="#FAFAFA"/>
  <text x="108" y="168" font-family="ui-monospace, monospace" font-size="28" font-weight="700" fill="#09090B" text-anchor="middle" dominant-baseline="central">%s</text>
  <text x="152" y="168" font-family="system-ui, -apple-system, sans-serif" font-size="28" font-weight="600" fill="#FAFAFA" dominant-baseline="central">%s</text>
  <text x="80" y="300" font-family="system-ui, -apple-system, sans-serif" font-size="72" font-weight="700" fill="#FAFAFA">Bureaucracy</text>
  <text x="80" y="390" font-family="system-ui, -apple-system, sans-serif" font-size="72" font-weight="700" fill="#FAFAFA">That Actually <tspan fill="#F59E0B">Moves</tspan></text>
  <text x="80" y="470" font-family="system-ui, -apple-system, sans-serif" font-size="26" fill="#A1A1AA">A no-nonsense task manager for approval workflows.</text>
</svg>`, safeLetter, safeName)

	c.Response().Header().Set("Content-Type", "image/svg+xml")
	c.Response().Header().Set("Cache-Control", "public, max-age=3600")
	return c.String(http.StatusOK, svg)
}
