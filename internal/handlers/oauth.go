package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"bereaucat/internal/auth"
	"bereaucat/internal/store"
)

// OAuthHandler handles SSO authentication flows.
type OAuthHandler struct {
	store       store.Querier
	authManager *auth.Manager
	authHandler *AuthHandler
	devMode     bool
}

// NewOAuthHandler creates a new OAuth handler.
func NewOAuthHandler(store store.Querier, authManager *auth.Manager, authHandler *AuthHandler, devMode bool) *OAuthHandler {
	return &OAuthHandler{
		store:       store,
		authManager: authManager,
		authHandler: authHandler,
		devMode:     devMode,
	}
}

// oidcClaims represents the claims extracted from an OIDC ID token.
type oidcClaims struct {
	Sub        string `json:"sub"`
	Email      string `json:"email"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Name       string `json:"name"`
}

// StartSSO initiates the OAuth2/OIDC flow for the given provider.
//
//	@Summary		Start SSO
//	@Description	Initiates the OAuth2/OIDC flow. Redirects to the identity provider.
//	@Tags			SSO
//	@Param			provider	path	string	true	"SSO provider (google or zitadel)"
//	@Success		307
//	@Failure		400	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/auth/sso/{provider} [get]
func (h *OAuthHandler) StartSSO(c *echo.Context) error {
	provider := c.Param("provider")
	if provider != "google" && provider != "zitadel" {
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported provider")
	}

	// Load SSO settings from DB
	ssoConfig, err := h.loadSSOConfig(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "SSO not configured")
	}

	providerConfig := h.getProviderConfig(ssoConfig, provider)
	if !providerConfig.Enabled {
		return echo.NewHTTPError(http.StatusBadRequest, "provider is not enabled")
	}

	// Build OAuth2 config
	oauth2Config, err := h.buildOAuth2Config(c.Request().Context(), provider, providerConfig)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to configure OAuth")
	}

	// Generate and sign state
	state, err := h.generateState()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate state")
	}

	// Set state cookie (SameSite=Lax required for cross-site redirect from IdP)
	h.setStateCookie(c, state, provider)

	// Redirect to provider's authorization URL
	authURL := oauth2Config.AuthCodeURL(state)
	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// CallbackSSO handles the OAuth2/OIDC callback from the provider.
//
//	@Summary		SSO callback
//	@Description	Handles the OAuth2/OIDC callback. Redirects to frontend with auth status.
//	@Tags			SSO
//	@Param			provider	path	string	true	"SSO provider (google or zitadel)"
//	@Param			code		query	string	false	"Authorization code"
//	@Param			state		query	string	false	"OAuth state"
//	@Success		307
//	@Router			/auth/sso/{provider}/callback [get]
func (h *OAuthHandler) CallbackSSO(c *echo.Context) error {
	provider := c.Param("provider")
	if provider != "google" && provider != "zitadel" {
		return h.redirectError(c, "unsupported provider")
	}

	// Validate state
	queryState := c.QueryParam("state")
	stateCookie, err := c.Cookie("oauth_state")
	if err != nil || stateCookie.Value == "" {
		return h.redirectError(c, "invalid state")
	}

	// Verify state cookie contains the correct signed state
	if !h.verifyState(stateCookie.Value, queryState) {
		return h.redirectError(c, "state mismatch")
	}

	// Get provider from cookie
	providerCookie, err := c.Cookie("oauth_provider")
	if err != nil || providerCookie.Value != provider {
		return h.redirectError(c, "provider mismatch")
	}

	// Clear state cookies
	h.clearStateCookies(c)

	// Check for error from provider
	if errParam := c.QueryParam("error"); errParam != "" {
		errDesc := c.QueryParam("error_description")
		if errDesc == "" {
			errDesc = errParam
		}
		return h.redirectError(c, errDesc)
	}

	code := c.QueryParam("code")
	if code == "" {
		return h.redirectError(c, "no authorization code received")
	}

	// Load SSO settings
	ssoConfig, err := h.loadSSOConfig(c)
	if err != nil {
		return h.redirectError(c, "SSO not configured")
	}

	providerConfig := h.getProviderConfig(ssoConfig, provider)
	if !providerConfig.Enabled {
		return h.redirectError(c, "provider is not enabled")
	}

	ctx := c.Request().Context()

	// Build OAuth2 config
	oauth2Config, err := h.buildOAuth2Config(ctx, provider, providerConfig)
	if err != nil {
		return h.redirectError(c, "failed to configure OAuth")
	}

	// Exchange code for tokens
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return h.redirectError(c, "failed to exchange authorization code")
	}

	// Extract user info from ID token
	claims, err := h.extractClaims(ctx, provider, providerConfig, token)
	if err != nil {
		return h.redirectError(c, "failed to verify identity")
	}

	if claims.Email == "" {
		return h.redirectError(c, "email not provided by identity provider")
	}

	// Find or create user
	user, err := h.findOrCreateUser(ctx, provider, claims)
	if err != nil {
		return h.redirectError(c, "failed to process user account")
	}

	// Generate tokens and set cookies
	_, err = h.authHandler.GenerateTokensAndSetCookies(c, ctx, user.ID, user.Username, user.UserType, *user)
	if err != nil {
		return h.redirectError(c, "failed to create session")
	}

	return c.Redirect(http.StatusTemporaryRedirect, "/auth/callback?status=success")
}

// loadSSOConfig loads the SSO settings from the database.
func (h *OAuthHandler) loadSSOConfig(c *echo.Context) (*SSOSettings, error) {
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

// getProviderConfig returns the config for the given provider.
func (h *OAuthHandler) getProviderConfig(sso *SSOSettings, provider string) SSOProviderConfig {
	switch provider {
	case "google":
		return sso.Google
	case "zitadel":
		return sso.Zitadel
	default:
		return SSOProviderConfig{}
	}
}

// buildOAuth2Config builds the OAuth2 configuration for the given provider.
func (h *OAuthHandler) buildOAuth2Config(ctx context.Context, provider string, config SSOProviderConfig) (*oauth2.Config, error) {
	scopes := []string{oidc.ScopeOpenID, "email", "profile"}

	switch provider {
	case "google":
		return &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURI,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		}, nil

	case "zitadel":
		oidcProvider, err := oidc.NewProvider(ctx, config.IssuerURL)
		if err != nil {
			return nil, fmt.Errorf("failed to discover OIDC provider: %w", err)
		}
		return &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURI,
			Scopes:       scopes,
			Endpoint:     oidcProvider.Endpoint(),
		}, nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// extractClaims verifies the ID token and extracts user claims.
func (h *OAuthHandler) extractClaims(ctx context.Context, provider string, config SSOProviderConfig, token *oauth2.Token) (*oidcClaims, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in response")
	}

	var issuer string
	switch provider {
	case "google":
		issuer = "https://accounts.google.com"
	case "zitadel":
		issuer = config.IssuerURL
	default:
		return nil, fmt.Errorf("unsupported provider")
	}

	oidcProvider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	verifier := oidcProvider.Verifier(&oidc.Config{ClientID: config.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	var claims oidcClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return &claims, nil
}

// findOrCreateUser finds an existing user or creates a new one for SSO.
// Returns the userInfo needed for token generation.
func (h *OAuthHandler) findOrCreateUser(ctx context.Context, provider string, claims *oidcClaims) (*userInfo, error) {
	// 1. Try to find by provider + subject ID
	user, err := h.store.GetUserByProviderID(ctx, store.GetUserByProviderIDParams{
		AuthProvider:   pgtype.Text{String: provider, Valid: true},
		ProviderUserID: pgtype.Text{String: claims.Sub, Valid: true},
	})
	if err == nil {
		return &userInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			UserType:  user.UserType,
			CreatedAt: user.CreatedAt.Time,
		}, nil
	}

	// 2. Try to find by email
	userByEmail, err := h.store.GetUserByEmail(ctx, claims.Email)
	if err == nil {
		// Link provider to existing user
		_ = h.store.LinkProviderToUser(ctx, store.LinkProviderToUserParams{
			ID:             userByEmail.ID,
			AuthProvider:   pgtype.Text{String: provider, Valid: true},
			ProviderUserID: pgtype.Text{String: claims.Sub, Valid: true},
		})
		return &userInfo{
			ID:        userByEmail.ID,
			Username:  userByEmail.Username,
			Email:     userByEmail.Email,
			FirstName: userByEmail.FirstName,
			LastName:  userByEmail.LastName,
			UserType:  userByEmail.UserType,
			CreatedAt: userByEmail.CreatedAt.Time,
		}, nil
	}

	// 3. Auto-create new user
	firstName := claims.GivenName
	lastName := claims.FamilyName
	if firstName == "" && lastName == "" && claims.Name != "" {
		parts := strings.SplitN(claims.Name, " ", 2)
		firstName = parts[0]
		if len(parts) > 1 {
			lastName = parts[1]
		}
	}
	if firstName == "" {
		firstName = strings.Split(claims.Email, "@")[0]
	}
	if lastName == "" {
		lastName = "-"
	}

	// Generate username from email
	username := h.generateUsername(ctx, claims.Email)

	newUser, err := h.store.CreateSSOUser(ctx, store.CreateSSOUserParams{
		Username:       username,
		Email:          claims.Email,
		FirstName:      firstName,
		LastName:       lastName,
		UserType:       "user",
		AuthProvider:   pgtype.Text{String: provider, Valid: true},
		ProviderUserID: pgtype.Text{String: claims.Sub, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &userInfo{
		ID:        newUser.ID,
		Username:  newUser.Username,
		Email:     newUser.Email,
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		UserType:  newUser.UserType,
		CreatedAt: newUser.CreatedAt.Time,
	}, nil
}

// generateUsername creates a unique username from an email address.
func (h *OAuthHandler) generateUsername(ctx context.Context, email string) string {
	base := strings.Split(email, "@")[0]
	// Clean the base: only alphanumeric and underscores
	clean := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, base)

	if len(clean) > 40 {
		clean = clean[:40]
	}

	// Check if username is taken
	exists, _ := h.store.UserExistsByEmailOrUsername(ctx, store.UserExistsByEmailOrUsernameParams{
		Email:    "", // won't match
		Username: clean,
	})
	if !exists {
		return clean
	}

	// Append random suffix
	for i := 0; i < 10; i++ {
		suffix, _ := rand.Int(rand.Reader, big.NewInt(10000))
		candidate := fmt.Sprintf("%s_%04d", clean, suffix.Int64())
		exists, _ = h.store.UserExistsByEmailOrUsername(ctx, store.UserExistsByEmailOrUsernameParams{
			Email:    "",
			Username: candidate,
		})
		if !exists {
			return candidate
		}
	}

	// Fallback: use timestamp
	return fmt.Sprintf("%s_%d", clean, time.Now().UnixNano()%100000)
}

// generateState creates a random state value.
func (h *OAuthHandler) generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// signState creates an HMAC signature for the state value.
func (h *OAuthHandler) signState(state string) string {
	mac := hmac.New(sha256.New, []byte(h.authManager.GetJWTSecret()))
	mac.Write([]byte(state))
	return base64.URLEncoding.EncodeToString(mac.Sum(nil))
}

// setStateCookie sets the OAuth state and provider cookies.
func (h *OAuthHandler) setStateCookie(c *echo.Context, state, provider string) {
	sig := h.signState(state)
	cookieValue := state + "." + sig

	stateCookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    cookieValue,
		Path:     "/api/v1/auth/sso/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   !h.devMode,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(stateCookie)

	providerCookie := &http.Cookie{
		Name:     "oauth_provider",
		Value:    provider,
		Path:     "/api/v1/auth/sso/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   !h.devMode,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(providerCookie)
}

// verifyState validates the state cookie value against the query state.
func (h *OAuthHandler) verifyState(cookieValue, queryState string) bool {
	parts := strings.SplitN(cookieValue, ".", 2)
	if len(parts) != 2 {
		return false
	}

	originalState := parts[0]
	sig := parts[1]

	// Verify the state matches
	if originalState != queryState {
		return false
	}

	// Verify the HMAC signature
	expectedSig := h.signState(originalState)
	return hmac.Equal([]byte(sig), []byte(expectedSig))
}

// clearStateCookies removes the OAuth state cookies.
func (h *OAuthHandler) clearStateCookies(c *echo.Context) {
	stateCookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/api/v1/auth/sso/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   !h.devMode,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(stateCookie)

	providerCookie := &http.Cookie{
		Name:     "oauth_provider",
		Value:    "",
		Path:     "/api/v1/auth/sso/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   !h.devMode,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(providerCookie)
}

// redirectError redirects to the frontend callback page with an error message.
func (h *OAuthHandler) redirectError(c *echo.Context, message string) error {
	return c.Redirect(http.StatusTemporaryRedirect, "/auth/callback?status=error&message="+url.QueryEscape(message))
}
