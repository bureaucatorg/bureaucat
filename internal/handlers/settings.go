package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v5"

	"bereaucat/internal/store"
)

// SettingsHandler handles application settings endpoints.
type SettingsHandler struct {
	store store.Querier
}

// NewSettingsHandler creates a new settings handler.
func NewSettingsHandler(store store.Querier) *SettingsHandler {
	return &SettingsHandler{store: store}
}

// BrandingSettings represents the branding configuration.
type BrandingSettings struct {
	Enabled bool   `json:"enabled"`
	AppName string `json:"app_name"`
}

// BrandingResponse is the API response for branding settings.
type BrandingResponse struct {
	Enabled bool   `json:"enabled"`
	AppName string `json:"app_name"`
}

// UpdateBrandingRequest is the request to update branding settings.
type UpdateBrandingRequest struct {
	Enabled bool   `json:"enabled"`
	AppName string `json:"app_name"`
}

// GetBranding returns the current branding settings.
// This endpoint is public so the frontend can display the correct app name.
func (h *SettingsHandler) GetBranding(c *echo.Context) error {
	ctx := c.Request().Context()

	setting, err := h.store.GetSetting(ctx, "branding")
	if err != nil {
		// Return defaults if not found
		return c.JSON(http.StatusOK, BrandingResponse{
			Enabled: false,
			AppName: "Bureaucat",
		})
	}

	var branding BrandingSettings
	if err := json.Unmarshal(setting.Value, &branding); err != nil {
		return c.JSON(http.StatusOK, BrandingResponse{
			Enabled: false,
			AppName: "Bureaucat",
		})
	}

	return c.JSON(http.StatusOK, BrandingResponse{
		Enabled: branding.Enabled,
		AppName: branding.AppName,
	})
}

// UpdateBranding updates the branding settings (admin only).
func (h *SettingsHandler) UpdateBranding(c *echo.Context) error {
	var req UpdateBrandingRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate app name
	if req.AppName == "" {
		req.AppName = "Bureaucat"
	}

	ctx := c.Request().Context()

	branding := BrandingSettings{
		Enabled: req.Enabled,
		AppName: req.AppName,
	}

	value, err := json.Marshal(branding)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to marshal settings")
	}

	_, err = h.store.UpsertSetting(ctx, store.UpsertSettingParams{
		Key:   "branding",
		Value: value,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update settings")
	}

	return c.JSON(http.StatusOK, BrandingResponse{
		Enabled: branding.Enabled,
		AppName: branding.AppName,
	})
}
