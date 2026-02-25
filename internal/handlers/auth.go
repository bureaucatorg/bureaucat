package handlers

import (
	"context"
	"net/http"
	"regexp"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/auth"
	"bereaucat/internal/store"
)

// SignupRequest represents the signup request body.
type SignupRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// SigninRequest represents the signin request body.
type SigninRequest struct {
	Identifier string `json:"identifier"` // email or username
	Password   string `json:"password"`
}

// AuthResponse represents the authentication response.
type AuthResponse struct {
	User        UserResponse `json:"user"`
	AccessToken string       `json:"access_token"`
	ExpiresAt   int64        `json:"expires_at"` // Unix timestamp
}

// UserResponse represents the user data in responses.
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	UserType  string    `json:"user_type"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	store       store.Querier
	authManager *auth.Manager
	devMode     bool
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(store store.Querier, authManager *auth.Manager, devMode bool) *AuthHandler {
	return &AuthHandler{
		store:       store,
		authManager: authManager,
		devMode:     devMode,
	}
}

// Signup handles user registration.
func (h *AuthHandler) Signup(c *echo.Context) error {
	var req SignupRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "all fields are required")
	}

	// Validate email format
	if !isValidEmail(req.Email) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid email format")
	}

	// Validate password (strict in production, any in dev)
	if !h.devMode {
		if errors := validatePassword(req.Password); len(errors) > 0 {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "password does not meet requirements",
				"errors":  errors,
			})
		}
	}

	ctx := c.Request().Context()

	// Check if user already exists
	exists, err := h.store.UserExistsByEmailOrUsername(ctx, store.UserExistsByEmailOrUsernameParams{
		Email:    req.Email,
		Username: req.Username,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check user existence")
	}
	if exists {
		return echo.NewHTTPError(http.StatusConflict, "user with this email or username already exists")
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
	}

	// Check if this is the first user (make them admin)
	count, err := h.store.CountUsers(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check user count")
	}

	userType := "user"
	if count == 0 {
		userType = "admin"
	}

	// Create user
	user, err := h.store.CreateUser(ctx, store.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		UserType:     userType,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}

	// Generate tokens
	return h.generateAndSetTokens(c, ctx, user.ID, user.Username, user.UserType, userFromCreateRow(user))
}

// Signin handles user login.
func (h *AuthHandler) Signin(c *echo.Context) error {
	var req SigninRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Identifier == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "identifier and password are required")
	}

	ctx := c.Request().Context()

	// Find user by email or username
	user, err := h.store.GetUserByEmailOrUsername(ctx, req.Identifier)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	// Verify password
	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	// Generate tokens
	return h.generateAndSetTokens(c, ctx, user.ID, user.Username, user.UserType, userFromFullUser(user))
}

// TokenRefresh handles token refresh.
func (h *AuthHandler) TokenRefresh(c *echo.Context) error {
	// Get refresh token from cookie
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "refresh token not found")
	}

	ctx := c.Request().Context()

	// Hash the token to look it up
	tokenHash := auth.HashToken(cookie.Value)

	// Find the refresh token
	refreshToken, err := h.store.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid refresh token")
	}

	// Revoke the old token (rotation)
	if err := h.store.RevokeRefreshToken(ctx, refreshToken.ID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to revoke token")
	}

	// Get the user
	user, err := h.store.GetUserByID(ctx, refreshToken.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	}

	// Generate new tokens
	return h.generateAndSetTokens(c, ctx, user.ID, user.Username, user.UserType, userFromGetByIDRow(user))
}

// Logout handles user logout by revoking all refresh tokens.
func (h *AuthHandler) Logout(c *echo.Context) error {
	// Get refresh token from cookie
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		// Even if no cookie, clear it and return success
		h.clearRefreshTokenCookie(c)
		return c.JSON(http.StatusOK, map[string]string{"message": "logged out"})
	}

	ctx := c.Request().Context()

	// Hash the token to look it up
	tokenHash := auth.HashToken(cookie.Value)

	// Find the refresh token to get the user ID
	refreshToken, err := h.store.GetRefreshTokenByHash(ctx, tokenHash)
	if err == nil {
		// Revoke all user's refresh tokens
		_ = h.store.RevokeAllUserRefreshTokens(ctx, refreshToken.UserID)
	}

	// Clear the cookie
	h.clearRefreshTokenCookie(c)

	return c.JSON(http.StatusOK, map[string]string{"message": "logged out"})
}

// Me returns the current user's info.
func (h *AuthHandler) Me(c *echo.Context) error {
	userIDStr := c.Request().Header.Get(auth.HeaderUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	ctx := c.Request().Context()

	user, err := h.store.GetUserByID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	return c.JSON(http.StatusOK, UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		UserType:  user.UserType,
		CreatedAt: user.CreatedAt.Time,
	})
}

// GetUserProfile returns a user's public profile by ID.
func (h *AuthHandler) GetUserProfile(c *echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	user, err := h.store.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	return c.JSON(http.StatusOK, UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		UserType:  user.UserType,
		CreatedAt: user.CreatedAt.Time,
	})
}

// userInfo holds common user fields for token generation.
type userInfo struct {
	ID        uuid.UUID
	Username  string
	Email     string
	FirstName string
	LastName  string
	UserType  string
	CreatedAt time.Time
}

func userFromCreateRow(u store.CreateUserRow) userInfo {
	return userInfo{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		UserType:  u.UserType,
		CreatedAt: u.CreatedAt.Time,
	}
}

func userFromFullUser(u store.User) userInfo {
	return userInfo{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		UserType:  u.UserType,
		CreatedAt: u.CreatedAt.Time,
	}
}

func userFromGetByIDRow(u store.GetUserByIDRow) userInfo {
	return userInfo{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		UserType:  u.UserType,
		CreatedAt: u.CreatedAt.Time,
	}
}

// generateAndSetTokens generates access and refresh tokens and sets cookies.
func (h *AuthHandler) generateAndSetTokens(c *echo.Context, ctx context.Context, userID uuid.UUID, username, userType string, user userInfo) error {
	// Generate access token
	accessToken, expiresAt, err := h.authManager.GenerateAccessToken(userID, username, userType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate access token")
	}

	// Generate refresh token
	refreshToken, refreshExpiresAt, err := h.authManager.GenerateRefreshToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate refresh token")
	}

	// Store refresh token hash in database
	tokenHash := auth.HashToken(refreshToken)
	_, err = h.store.CreateRefreshToken(ctx, store.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: refreshExpiresAt, Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to store refresh token")
	}

	// Set refresh token as httpOnly cookie
	h.setRefreshTokenCookie(c, refreshToken, refreshExpiresAt)

	// Set access token as httpOnly cookie (backup for page reload)
	h.setAccessTokenCookie(c, accessToken, expiresAt)

	return c.JSON(http.StatusOK, AuthResponse{
		User: UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			UserType:  user.UserType,
			CreatedAt: user.CreatedAt,
		},
		AccessToken: accessToken,
		ExpiresAt:   expiresAt.Unix(),
	})
}

func (h *AuthHandler) setRefreshTokenCookie(c *echo.Context, token string, expiresAt time.Time) {
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/token_refresh",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   !h.devMode,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)
}

func (h *AuthHandler) setAccessTokenCookie(c *echo.Context, token string, expiresAt time.Time) {
	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   !h.devMode,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)
}

func (h *AuthHandler) clearRefreshTokenCookie(c *echo.Context) {
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/token_refresh",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   !h.devMode,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)

	// Also clear access token cookie
	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   !h.devMode,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(accessCookie)
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func validatePassword(password string) []string {
	var errors []string

	if len(password) < 8 {
		errors = append(errors, "password must be at least 8 characters")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		errors = append(errors, "password must contain at least one uppercase letter")
	}
	if !hasLower {
		errors = append(errors, "password must contain at least one lowercase letter")
	}
	if !hasNumber {
		errors = append(errors, "password must contain at least one number")
	}
	if !hasSpecial {
		errors = append(errors, "password must contain at least one special character")
	}

	return errors
}
