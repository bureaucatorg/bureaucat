package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
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
	AvatarURL *string   `json:"avatar_url,omitempty"`
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
//
//	@Summary		Sign up
//	@Description	Register a new user account. The first user created becomes an admin.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		SignupRequest	true	"Registration details"
//	@Success		200		{object}	AuthResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/signup [post]
func (h *AuthHandler) Signup(c *echo.Context) error {
	// Check if signups are enabled
	if !h.isSignupEnabled(c.Request().Context()) {
		return echo.NewHTTPError(http.StatusForbidden, "new signups are currently disabled")
	}

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
		PasswordHash: pgtype.Text{String: passwordHash, Valid: true},
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
//
//	@Summary		Sign in
//	@Description	Authenticate with email/username and password. Returns access token and sets refresh token cookie.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		SigninRequest	true	"Login credentials"
//	@Success		200		{object}	AuthResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/signin [post]
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

	// Verify password (SSO-only users have no password)
	if !user.PasswordHash.Valid || user.PasswordHash.String == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if !auth.CheckPassword(req.Password, user.PasswordHash.String) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	// Generate tokens
	return h.generateAndSetTokens(c, ctx, user.ID, user.Username, user.UserType, userFromFullUser(user))
}

// TokenRefresh handles token refresh.
//
//	@Summary		Refresh token
//	@Description	Refresh the access token using the refresh token cookie. Performs token rotation.
//	@Tags			Auth
//	@Produce		json
//	@Success		200	{object}	AuthResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/token_refresh [post]
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
//
//	@Summary		Logout
//	@Description	Revoke all refresh tokens for the current user and clear cookies.
//	@Tags			Auth
//	@Produce		json
//	@Success		200	{object}	MessageResponse
//	@Router			/logout [post]
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
//
//	@Summary		Get current user
//	@Description	Returns the authenticated user's profile information.
//	@Tags			Auth
//	@Produce		json
//	@Success		200	{object}	UserResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/me [get]
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
		AvatarURL: textToStringPtr(user.AvatarUrl),
		CreatedAt: user.CreatedAt.Time,
	})
}

// GetUserProfile returns a user's public profile by ID.
//
//	@Summary		Get user profile
//	@Description	Returns a user's public profile by their ID.
//	@Tags			Auth
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	UserResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/users/{id} [get]
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
		AvatarURL: textToStringPtr(user.AvatarUrl),
		CreatedAt: user.CreatedAt.Time,
	})
}

// MyTaskAssignee represents a task assignee in the my-tasks response.
type MyTaskAssignee struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

// MyTaskItem represents a task assigned to the current user.
type MyTaskItem struct {
	ID           uuid.UUID        `json:"id"`
	ProjectKey   string           `json:"project_key"`
	TaskNumber   int32            `json:"task_number"`
	TaskID       string           `json:"task_id"`
	Title        string           `json:"title"`
	StateName    string           `json:"state_name"`
	StateType    string           `json:"state_type"`
	StateColor   string           `json:"state_color"`
	Priority     int32            `json:"priority"`
	Assignees    []MyTaskAssignee `json:"assignees"`
	CommentCount int              `json:"comment_count"`
}

// MyTasksResponse represents the paginated response for user's assigned tasks.
type MyTasksResponse struct {
	Tasks      []MyTaskItem `json:"tasks"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PerPage    int          `json:"per_page"`
	TotalPages int          `json:"total_pages"`
}

// MyTasks returns tasks assigned to the current user across all projects.
func (h *AuthHandler) MyTasks(c *echo.Context) error {
	userIDStr := c.Request().Header.Get(auth.HeaderUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	ctx := c.Request().Context()

	tasks, err := h.store.ListTasksByAssignee(ctx, store.ListTasksByAssigneeParams{
		UserID: userID,
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list tasks")
	}

	total, err := h.store.CountTasksByAssignee(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count tasks")
	}

	items := make([]MyTaskItem, len(tasks))
	for i, t := range tasks {
		stateColor := ""
		if t.StateColor.Valid {
			stateColor = t.StateColor.String
		}

		// Fetch assignees
		assigneeRows, _ := h.store.ListTaskAssignees(ctx, t.ID)
		assignees := make([]MyTaskAssignee, len(assigneeRows))
		for j, a := range assigneeRows {
			assignees[j] = MyTaskAssignee{
				ID:        a.ID,
				UserID:    a.UserID,
				Username:  a.Username,
				Email:     a.Email,
				FirstName: a.FirstName,
				LastName:  a.LastName,
				AvatarURL: textToStringPtr(a.AvatarUrl),
			}
		}

		// Fetch comment count
		commentCount, _ := h.store.CountTaskComments(ctx, t.ID)

		items[i] = MyTaskItem{
			ID:           t.ID,
			ProjectKey:   t.ProjectKey,
			TaskNumber:   t.TaskNumber,
			TaskID:       t.ProjectKey + "-" + strconv.Itoa(int(t.TaskNumber)),
			Title:        t.Title,
			StateName:    t.StateName,
			StateType:    t.StateType,
			StateColor:   stateColor,
			Priority:     t.Priority,
			Assignees:    assignees,
			CommentCount: int(commentCount),
		}
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, MyTasksResponse{
		Tasks:      items,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
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
	AvatarURL *string
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

func userFromFullUser(u store.GetUserByEmailOrUsernameRow) userInfo {
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

// GenerateTokensAndSetCookies creates access + refresh tokens, stores refresh in DB, and sets cookies.
// Returns the AuthResponse data without sending it. Used by both normal auth and SSO callback.
func (h *AuthHandler) GenerateTokensAndSetCookies(c *echo.Context, ctx context.Context, userID uuid.UUID, username, userType string, user userInfo) (*AuthResponse, error) {
	// Generate access token
	accessToken, expiresAt, err := h.authManager.GenerateAccessToken(userID, username, userType)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, refreshExpiresAt, err := h.authManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store refresh token hash in database
	tokenHash := auth.HashToken(refreshToken)
	_, err = h.store.CreateRefreshToken(ctx, store.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: refreshExpiresAt, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// Set refresh token as httpOnly cookie
	h.setRefreshTokenCookie(c, refreshToken, refreshExpiresAt)

	// Set access token as httpOnly cookie (backup for page reload)
	h.setAccessTokenCookie(c, accessToken, expiresAt)

	return &AuthResponse{
		User: UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			UserType:  user.UserType,
			AvatarURL: user.AvatarURL,
			CreatedAt: user.CreatedAt,
		},
		AccessToken: accessToken,
		ExpiresAt:   expiresAt.Unix(),
	}, nil
}

// generateAndSetTokens generates access and refresh tokens, sets cookies, and returns JSON response.
func (h *AuthHandler) generateAndSetTokens(c *echo.Context, ctx context.Context, userID uuid.UUID, username, userType string, user userInfo) error {
	resp, err := h.GenerateTokensAndSetCookies(c, ctx, userID, username, userType, user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate tokens")
	}
	return c.JSON(http.StatusOK, resp)
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

// isSignupEnabled checks the "signup" setting from the database.
// Returns true (enabled) by default if the setting doesn't exist.
func (h *AuthHandler) isSignupEnabled(ctx context.Context) bool {
	setting, err := h.store.GetSetting(ctx, "signup")
	if err != nil {
		return true // Default: enabled
	}

	var signup struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.Unmarshal(setting.Value, &signup); err != nil {
		return true
	}

	return signup.Enabled
}

// UserActivityResponse represents a user activity entry with task/project context.
type UserActivityResponse struct {
	ID           uuid.UUID   `json:"id"`
	TaskID       uuid.UUID   `json:"task_id"`
	ActivityType string      `json:"activity_type"`
	ActorID      uuid.UUID   `json:"actor_id"`
	Username     string      `json:"username"`
	FirstName    string      `json:"first_name"`
	LastName     string      `json:"last_name"`
	FieldName    *string     `json:"field_name,omitempty"`
	OldValue     interface{} `json:"old_value,omitempty"`
	NewValue     interface{} `json:"new_value,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	TaskNumber   int32       `json:"task_number"`
	ProjectKey   string      `json:"project_key"`
	TaskTitle    string      `json:"task_title"`
}

// UserActivityDateCount represents the activity count on a single date.
type UserActivityDateCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// GetUserActivity returns paginated activity for a user across all projects.
func (h *AuthHandler) GetUserActivity(c *echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 30
	}
	offset := (page - 1) * perPage

	ctx := c.Request().Context()

	activities, err := h.store.ListUserActivity(ctx, store.ListUserActivityParams{
		ActorID: userID,
		Limit:   int32(perPage),
		Offset:  int32(offset),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list user activity")
	}

	total, err := h.store.CountUserActivity(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count user activity")
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	items := make([]UserActivityResponse, len(activities))
	for i, a := range activities {
		var fieldName *string
		if a.FieldName.Valid {
			fieldName = &a.FieldName.String
		}
		items[i] = UserActivityResponse{
			ID:           a.ID,
			TaskID:       a.TaskID,
			ActivityType: a.ActivityType,
			ActorID:      a.ActorID,
			Username:     a.Username,
			FirstName:    a.FirstName,
			LastName:     a.LastName,
			FieldName:    fieldName,
			OldValue:     parseJSONBAuth(a.OldValue),
			NewValue:     parseJSONBAuth(a.NewValue),
			CreatedAt:    a.CreatedAt.Time,
			TaskNumber:   a.TaskNumber,
			ProjectKey:   a.ProjectKey,
			TaskTitle:    a.TaskTitle,
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"activities":  items,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
		"total_pages": totalPages,
	})
}

// GetUserActivityGraph returns daily activity counts for the contribution graph.
func (h *AuthHandler) GetUserActivityGraph(c *echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	// Default to last 365 days
	since := time.Now().AddDate(-1, 0, 0)

	ctx := c.Request().Context()

	dates, err := h.store.ListUserActivityDates(ctx, store.ListUserActivityDatesParams{
		ActorID:   userID,
		CreatedAt: pgtype.Timestamptz{Time: since, Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list activity dates")
	}

	items := make([]UserActivityDateCount, len(dates))
	for i, d := range dates {
		items[i] = UserActivityDateCount{
			Date:  d.ActivityDate.Time.Format("2006-01-02"),
			Count: int(d.ActivityCount),
		}
	}

	return c.JSON(http.StatusOK, items)
}

// GetMyNotifications returns activity on tasks the current user is involved with.
func (h *AuthHandler) GetMyNotifications(c *echo.Context) error {
	userIDStr := c.Request().Header.Get("X-User-ID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	ctx := c.Request().Context()

	activities, err := h.store.ListUserNotifications(ctx, store.ListUserNotificationsParams{
		ActorID: userID,
		Limit:   int32(perPage),
		Offset:  int32(offset),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list notifications")
	}

	total, err := h.store.CountUserNotifications(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count notifications")
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	items := make([]UserActivityResponse, len(activities))
	for i, a := range activities {
		var fieldName *string
		if a.FieldName.Valid {
			fieldName = &a.FieldName.String
		}
		items[i] = UserActivityResponse{
			ID:           a.ID,
			TaskID:       a.TaskID,
			ActivityType: a.ActivityType,
			ActorID:      a.ActorID,
			Username:     a.Username,
			FirstName:    a.FirstName,
			LastName:     a.LastName,
			FieldName:    fieldName,
			OldValue:     parseJSONBAuth(a.OldValue),
			NewValue:     parseJSONBAuth(a.NewValue),
			CreatedAt:    a.CreatedAt.Time,
			TaskNumber:   a.TaskNumber,
			ProjectKey:   a.ProjectKey,
			TaskTitle:    a.TaskTitle,
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"activities":  items,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
		"total_pages": totalPages,
	})
}

// parseJSONBAuth safely parses JSONB data.
func parseJSONBAuth(data []byte) interface{} {
	if data == nil {
		return nil
	}
	return string(data)
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
