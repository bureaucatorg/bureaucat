package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"bereaucat/internal/notifier"
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
//
//	@Summary		Get branding
//	@Description	Returns the current branding settings. Public endpoint.
//	@Tags			Settings
//	@Produce		json
//	@Success		200	{object}	BrandingResponse
//	@Router			/settings/branding [get]
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
//
//	@Summary		Update branding
//	@Description	Update branding settings. Requires admin role.
//	@Tags			Admin - Settings
//	@Accept			json
//	@Produce		json
//	@Param			body	body		UpdateBrandingRequest	true	"Branding settings"
//	@Success		200		{object}	BrandingResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/settings/branding [put]
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

// --- Signup Settings ---

// SignupSettings represents the signup configuration.
type SignupSettings struct {
	Enabled bool `json:"enabled"`
}

// GetSignupSettings returns whether public signups are enabled.
// This endpoint is public so the frontend can hide the signup page when disabled.
func (h *SettingsHandler) GetSignupSettings(c *echo.Context) error {
	ctx := c.Request().Context()

	setting, err := h.store.GetSetting(ctx, "signup")
	if err != nil {
		// Default: signups are enabled
		return c.JSON(http.StatusOK, SignupSettings{Enabled: true})
	}

	var signup SignupSettings
	if err := json.Unmarshal(setting.Value, &signup); err != nil {
		return c.JSON(http.StatusOK, SignupSettings{Enabled: true})
	}

	return c.JSON(http.StatusOK, signup)
}

// UpdateSignupSettings updates the signup settings (admin only).
func (h *SettingsHandler) UpdateSignupSettings(c *echo.Context) error {
	var req SignupSettings
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ctx := c.Request().Context()

	value, err := json.Marshal(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to marshal settings")
	}

	_, err = h.store.UpsertSetting(ctx, store.UpsertSettingParams{
		Key:   "signup",
		Value: value,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update settings")
	}

	return c.JSON(http.StatusOK, req)
}

// IsSignupEnabled checks if public signups are enabled (used by auth handler).
func (h *SettingsHandler) IsSignupEnabled(ctx context.Context) bool {
	setting, err := h.store.GetSetting(ctx, "signup")
	if err != nil {
		return true // Default: enabled
	}

	var signup SignupSettings
	if err := json.Unmarshal(setting.Value, &signup); err != nil {
		return true
	}

	return signup.Enabled
}

// --- SSO Settings ---

// SSOProviderConfig represents configuration for a single SSO provider.
type SSOProviderConfig struct {
	Enabled      bool   `json:"enabled"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	IssuerURL    string `json:"issuer_url,omitempty"` // Zitadel only
	RedirectURI  string `json:"redirect_uri"`
}

// SSOSettings holds all SSO provider configs.
type SSOSettings struct {
	Google  SSOProviderConfig `json:"google"`
	Zitadel SSOProviderConfig `json:"zitadel"`
}

// SSOProvidersPublicResponse is the public response showing only which providers are enabled.
type SSOProvidersPublicResponse struct {
	Google  bool `json:"google"`
	Zitadel bool `json:"zitadel"`
}

// GetSSOProviders returns which SSO providers are enabled (public, no secrets).
//
//	@Summary		Get SSO providers
//	@Description	Returns which SSO providers are enabled. Public endpoint, no secrets exposed.
//	@Tags			Settings
//	@Produce		json
//	@Success		200	{object}	SSOProvidersPublicResponse
//	@Router			/settings/sso [get]
func (h *SettingsHandler) GetSSOProviders(c *echo.Context) error {
	ctx := c.Request().Context()

	setting, err := h.store.GetSetting(ctx, "sso")
	if err != nil {
		return c.JSON(http.StatusOK, SSOProvidersPublicResponse{
			Google:  false,
			Zitadel: false,
		})
	}

	var sso SSOSettings
	if err := json.Unmarshal(setting.Value, &sso); err != nil {
		return c.JSON(http.StatusOK, SSOProvidersPublicResponse{
			Google:  false,
			Zitadel: false,
		})
	}

	return c.JSON(http.StatusOK, SSOProvidersPublicResponse{
		Google:  sso.Google.Enabled,
		Zitadel: sso.Zitadel.Enabled,
	})
}

// GetSSOSettings returns the full SSO config with secrets masked (admin only).
//
//	@Summary		Get SSO settings
//	@Description	Returns full SSO configuration with secrets masked. Requires admin role.
//	@Tags			Admin - Settings
//	@Produce		json
//	@Success		200	{object}	SSOSettings
//	@Security		BearerAuth
//	@Router			/admin/settings/sso [get]
func (h *SettingsHandler) GetSSOSettings(c *echo.Context) error {
	ctx := c.Request().Context()

	setting, err := h.store.GetSetting(ctx, "sso")
	if err != nil {
		return c.JSON(http.StatusOK, SSOSettings{})
	}

	var sso SSOSettings
	if err := json.Unmarshal(setting.Value, &sso); err != nil {
		return c.JSON(http.StatusOK, SSOSettings{})
	}

	// Mask secrets before returning
	sso.Google.ClientSecret = maskSecret(sso.Google.ClientSecret)
	sso.Zitadel.ClientSecret = maskSecret(sso.Zitadel.ClientSecret)

	return c.JSON(http.StatusOK, sso)
}

// UpdateSSOSettings saves the SSO configuration (admin only).
//
//	@Summary		Update SSO settings
//	@Description	Update SSO provider configuration. Masked secrets are preserved. Requires admin role.
//	@Tags			Admin - Settings
//	@Accept			json
//	@Produce		json
//	@Param			body	body		SSOSettings	true	"SSO configuration"
//	@Success		200		{object}	SSOSettings
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/settings/sso [put]
func (h *SettingsHandler) UpdateSSOSettings(c *echo.Context) error {
	var req SSOSettings
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ctx := c.Request().Context()

	// Load existing settings to preserve secrets if masked/empty
	var existing SSOSettings
	setting, err := h.store.GetSetting(ctx, "sso")
	if err == nil {
		_ = json.Unmarshal(setting.Value, &existing)
	}

	// Preserve existing secrets if the new value is masked or empty
	if isSecretMasked(req.Google.ClientSecret) {
		req.Google.ClientSecret = existing.Google.ClientSecret
	}
	if isSecretMasked(req.Zitadel.ClientSecret) {
		req.Zitadel.ClientSecret = existing.Zitadel.ClientSecret
	}

	// Validate: if enabled, required fields must be set
	if req.Google.Enabled {
		if req.Google.ClientID == "" || req.Google.ClientSecret == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Google SSO requires client_id and client_secret")
		}
	}
	if req.Zitadel.Enabled {
		if req.Zitadel.ClientID == "" || req.Zitadel.ClientSecret == "" || req.Zitadel.IssuerURL == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Zitadel SSO requires client_id, client_secret, and issuer_url")
		}
	}

	// Auto-generate redirect URIs from request host
	scheme := "https"
	if c.Request().TLS == nil {
		// Check X-Forwarded-Proto header
		if proto := c.Request().Header.Get("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		} else {
			scheme = "http"
		}
	}
	host := c.Request().Host
	baseURL := scheme + "://" + host

	req.Google.RedirectURI = baseURL + "/api/v1/auth/sso/google/callback"
	req.Zitadel.RedirectURI = baseURL + "/api/v1/auth/sso/zitadel/callback"

	value, err := json.Marshal(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to marshal settings")
	}

	_, err = h.store.UpsertSetting(ctx, store.UpsertSettingParams{
		Key:   "sso",
		Value: value,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update SSO settings")
	}

	// Return with secrets masked
	req.Google.ClientSecret = maskSecret(req.Google.ClientSecret)
	req.Zitadel.ClientSecret = maskSecret(req.Zitadel.ClientSecret)

	return c.JSON(http.StatusOK, req)
}

// LoadSSOSettings loads the SSO config from the database (used by OAuth handler).
func (h *SettingsHandler) LoadSSOSettings(c *echo.Context) (*SSOSettings, error) {
	setting, err := h.store.GetSetting(c.Request().Context(), "sso")
	if err != nil {
		return nil, err
	}

	var sso SSOSettings
	if err := json.Unmarshal(setting.Value, &sso); err != nil {
		return nil, err
	}
	return &sso, nil
}

// --- Mattermost Settings ---

// MattermostSettings represents the Mattermost integration configuration.
type MattermostSettings struct {
	Enabled   bool   `json:"enabled"`
	ServerURL string `json:"server_url"`
	BotToken  string `json:"bot_token"`
}

// MattermostResponse is the API response for Mattermost settings (token masked).
type MattermostResponse struct {
	Enabled   bool   `json:"enabled"`
	ServerURL string `json:"server_url"`
	BotToken  string `json:"bot_token"`
}

// GetMattermostSettings returns the Mattermost integration config with token masked (admin only).
//
//	@Summary		Get Mattermost settings
//	@Description	Returns Mattermost integration configuration with bot token masked. Requires admin role.
//	@Tags			Admin - Settings
//	@Produce		json
//	@Success		200	{object}	MattermostResponse
//	@Security		BearerAuth
//	@Router			/admin/settings/mattermost [get]
func (h *SettingsHandler) GetMattermostSettings(c *echo.Context) error {
	ctx := c.Request().Context()

	setting, err := h.store.GetSetting(ctx, "mattermost")
	if err != nil {
		return c.JSON(http.StatusOK, MattermostResponse{})
	}

	var mm MattermostSettings
	if err := json.Unmarshal(setting.Value, &mm); err != nil {
		return c.JSON(http.StatusOK, MattermostResponse{})
	}

	return c.JSON(http.StatusOK, MattermostResponse{
		Enabled:   mm.Enabled,
		ServerURL: mm.ServerURL,
		BotToken:  maskSecret(mm.BotToken),
	})
}

// UpdateMattermostSettings saves the Mattermost integration configuration (admin only).
//
//	@Summary		Update Mattermost settings
//	@Description	Update Mattermost integration configuration. Masked tokens are preserved. Requires admin role.
//	@Tags			Admin - Settings
//	@Accept			json
//	@Produce		json
//	@Param			body	body		MattermostSettings	true	"Mattermost configuration"
//	@Success		200		{object}	MattermostResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/settings/mattermost [put]
func (h *SettingsHandler) UpdateMattermostSettings(c *echo.Context) error {
	var req MattermostSettings
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ctx := c.Request().Context()

	// Load existing settings to preserve token if masked
	var existing MattermostSettings
	setting, err := h.store.GetSetting(ctx, "mattermost")
	if err == nil {
		_ = json.Unmarshal(setting.Value, &existing)
	}

	if isSecretMasked(req.BotToken) {
		req.BotToken = existing.BotToken
	}

	// Validate: if enabled, required fields must be set
	if req.Enabled {
		if req.ServerURL == "" || req.BotToken == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Mattermost requires server_url and bot_token")
		}
	}

	value, err := json.Marshal(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to marshal settings")
	}

	_, err = h.store.UpsertSetting(ctx, store.UpsertSettingParams{
		Key:   "mattermost",
		Value: value,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update Mattermost settings")
	}

	return c.JSON(http.StatusOK, MattermostResponse{
		Enabled:   req.Enabled,
		ServerURL: req.ServerURL,
		BotToken:  maskSecret(req.BotToken),
	})
}

// TestMattermostConnection tests the Mattermost bot connection (admin only).
//
//	@Summary		Test Mattermost connection
//	@Description	Tests the Mattermost bot connection using the saved settings. Requires admin role.
//	@Tags			Admin - Settings
//	@Produce		json
//	@Success		200	{object}	MessageResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/settings/mattermost/test [post]
func (h *SettingsHandler) TestMattermostConnection(c *echo.Context) error {
	ctx := c.Request().Context()

	setting, err := h.store.GetSetting(ctx, "mattermost")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Mattermost not configured")
	}

	var mm MattermostSettings
	if err := json.Unmarshal(setting.Value, &mm); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to parse Mattermost settings")
	}

	if !mm.Enabled || mm.ServerURL == "" || mm.BotToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Mattermost is not configured or not enabled")
	}

	notif := notifierFromSettings(mm)
	if err := notif.TestConnection(ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "connection failed: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "connection successful"})
}

// LoadMattermostSettings loads the Mattermost config from the database.
func (h *SettingsHandler) LoadMattermostSettings(ctx context.Context) (*MattermostSettings, error) {
	setting, err := h.store.GetSetting(ctx, "mattermost")
	if err != nil {
		return nil, err
	}

	var mm MattermostSettings
	if err := json.Unmarshal(setting.Value, &mm); err != nil {
		return nil, err
	}
	return &mm, nil
}

// maskSecret replaces all but the last 4 characters with asterisks.
func maskSecret(secret string) string {
	if secret == "" {
		return ""
	}
	if len(secret) <= 4 {
		return "****"
	}
	return strings.Repeat("*", len(secret)-4) + secret[len(secret)-4:]
}

// isSecretMasked returns true if the secret is empty or contains the mask pattern.
func isSecretMasked(secret string) bool {
	return secret == "" || strings.HasPrefix(secret, "****")
}

// notifierFromSettings creates a MattermostNotifier from settings.
func notifierFromSettings(mm MattermostSettings) *notifier.MattermostNotifier {
	return notifier.NewMattermostNotifier(notifier.MattermostConfig{
		ServerURL: mm.ServerURL,
		BotToken:  mm.BotToken,
	})
}
