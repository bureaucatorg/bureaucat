package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/auth"
	"bereaucat/internal/store"
)

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Message string `json:"message"`
}

// MessageResponse represents a generic success message.
type MessageResponse struct {
	Message string `json:"message"`
}

// CleanupResponse represents the response from cleanup operations.
type CleanupResponse struct {
	Message string `json:"message"`
	Deleted int64  `json:"deleted"`
}

// AdminHandler handles admin-only endpoints.
type AdminHandler struct {
	store       store.Querier
	pool        *pgxpool.Pool
	authManager *auth.Manager
	devMode     bool
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(store store.Querier, pool *pgxpool.Pool, authManager *auth.Manager, devMode bool) *AdminHandler {
	return &AdminHandler{
		store:       store,
		pool:        pool,
		authManager: authManager,
		devMode:     devMode,
	}
}

// CreateUserRequest represents the admin create user request.
type CreateUserRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserType  string `json:"user_type"`
}

// PaginatedUsersResponse represents a paginated list of users.
type PaginatedUsersResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

// TokenInfo represents a refresh token with user info.
type TokenInfo struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt string    `json:"created_at"`
	ExpiresAt string    `json:"expires_at"`
}

// PaginatedTokensResponse represents a paginated list of tokens.
type PaginatedTokensResponse struct {
	Tokens     []TokenInfo `json:"tokens"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

// ListUsers returns paginated list of all users.
//
//	@Summary		List users
//	@Description	Returns a paginated list of all users. Supports search by username, email, or name.
//	@Tags			Admin - Users
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(20)
//	@Param			search		query		string	false	"Search by username, email, or name"
//	@Success		200			{object}	PaginatedUsersResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users [get]
func (h *AdminHandler) ListUsers(c *echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage
	search := c.QueryParam("search")

	ctx := c.Request().Context()

	var total int64
	var users []store.ListUsersPaginatedRow
	var err error

	if search != "" {
		// Search with filter
		searchText := pgtype.Text{String: search, Valid: true}
		total, err = h.store.CountSearchUsers(ctx, searchText)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to count users")
		}
		searchResults, searchErr := h.store.SearchUsersPaginated(ctx, store.SearchUsersPaginatedParams{
			Column1: searchText,
			Limit:   int32(perPage),
			Offset:  int32(offset),
		})
		if searchErr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to search users")
		}
		// Convert search results to same type
		users = make([]store.ListUsersPaginatedRow, len(searchResults))
		for i, u := range searchResults {
			users[i] = store.ListUsersPaginatedRow{
				ID:        u.ID,
				Username:  u.Username,
				Email:     u.Email,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				UserType:  u.UserType,
				CreatedAt: u.CreatedAt,
				UpdatedAt: u.UpdatedAt,
			}
		}
	} else {
		// No search filter
		total, err = h.store.CountUsers(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to count users")
		}
		users, err = h.store.ListUsersPaginated(ctx, store.ListUsersPaginatedParams{
			Limit:  int32(perPage),
			Offset: int32(offset),
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list users")
		}
	}

	// Convert to response format
	userResponses := make([]UserResponse, len(users))
	for i, u := range users {
		userResponses[i] = UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			UserType:  u.UserType,
			CreatedAt: u.CreatedAt.Time,
		}
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, PaginatedUsersResponse{
		Users:      userResponses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// CreateUser creates a new user (admin can create any type).
//
//	@Summary		Create user
//	@Description	Create a new user account with the specified details.
//	@Tags			Admin - Users
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateUserRequest	true	"User details"
//	@Success		201		{object}	UserResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users [post]
func (h *AdminHandler) CreateUser(c *echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" ||
		req.FirstName == "" || req.LastName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "all fields are required")
	}

	// Validate user type
	if req.UserType != "admin" && req.UserType != "user" {
		return echo.NewHTTPError(http.StatusBadRequest, "user_type must be 'admin' or 'user'")
	}

	// Validate email
	if !isValidEmail(req.Email) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid email format")
	}

	ctx := c.Request().Context()

	// Check if user exists
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

	// Create user
	user, err := h.store.CreateUser(ctx, store.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: pgtype.Text{String: passwordHash, Valid: true},
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		UserType:     req.UserType,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}

	return c.JSON(http.StatusCreated, UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		UserType:  user.UserType,
		CreatedAt: user.CreatedAt.Time,
	})
}

// DeleteUser deletes a user by ID.
//
//	@Summary		Delete user
//	@Description	Delete a user by their ID. Cannot delete yourself.
//	@Tags			Admin - Users
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	MessageResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users/{id} [delete]
func (h *AdminHandler) DeleteUser(c *echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	// Get current admin's ID from header
	currentUserIDStr := c.Request().Header.Get(auth.HeaderUserID)
	currentUserID, _ := uuid.Parse(currentUserIDStr)

	// Prevent self-deletion
	if userID == currentUserID {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot delete yourself")
	}

	ctx := c.Request().Context()

	// Check user exists
	_, err = h.store.GetUserByID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	// Delete user (cascade will delete refresh tokens)
	err = h.store.DeleteUserByID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete user")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "user deleted"})
}

// ListTokens returns paginated list of active refresh tokens.
//
//	@Summary		List active tokens
//	@Description	Returns a paginated list of active refresh tokens with user info.
//	@Tags			Admin - Tokens
//	@Produce		json
//	@Param			page		query		int	false	"Page number"		default(1)
//	@Param			per_page	query		int	false	"Items per page"	default(20)
//	@Success		200			{object}	PaginatedTokensResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/tokens [get]
func (h *AdminHandler) ListTokens(c *echo.Context) error {
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

	// Get total count
	total, err := h.store.CountActiveRefreshTokens(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count tokens")
	}

	// Get paginated tokens
	tokens, err := h.store.ListActiveRefreshTokens(ctx, store.ListActiveRefreshTokensParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list tokens")
	}

	// Convert to response format
	tokenInfos := make([]TokenInfo, len(tokens))
	for i, t := range tokens {
		tokenInfos[i] = TokenInfo{
			ID:        t.ID,
			UserID:    t.UserID,
			Username:  t.Username,
			Email:     t.Email,
			CreatedAt: t.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			ExpiresAt: t.ExpiresAt.Time.Format("2006-01-02T15:04:05Z"),
		}
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, PaginatedTokensResponse{
		Tokens:     tokenInfos,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// RevokeToken revokes a specific refresh token.
//
//	@Summary		Revoke token
//	@Description	Revoke a specific refresh token by ID.
//	@Tags			Admin - Tokens
//	@Produce		json
//	@Param			id	path		string	true	"Token ID"
//	@Success		200	{object}	MessageResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/tokens/{id} [delete]
func (h *AdminHandler) RevokeToken(c *echo.Context) error {
	tokenIDStr := c.Param("id")
	tokenID, err := uuid.Parse(tokenIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token ID")
	}

	ctx := c.Request().Context()

	// Check token exists
	_, err = h.store.GetRefreshTokenByID(ctx, tokenID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "token not found")
	}

	// Revoke token
	err = h.store.RevokeRefreshToken(ctx, tokenID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to revoke token")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "token revoked"})
}

// UpdateUserRoleRequest represents the request to update a user's role.
type UpdateUserRoleRequest struct {
	UserType string `json:"user_type"`
}

// UpdateUserRole updates a user's role (admin/user).
//
//	@Summary		Update user role
//	@Description	Change a user's role between admin and user. Cannot demote yourself.
//	@Tags			Admin - Users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"User ID"
//	@Param			body	body		UpdateUserRoleRequest	true	"New role"
//	@Success		200		{object}	UserResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users/{id}/role [put]
func (h *AdminHandler) UpdateUserRole(c *echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	var req UpdateUserRoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.UserType != "admin" && req.UserType != "user" {
		return echo.NewHTTPError(http.StatusBadRequest, "user_type must be 'admin' or 'user'")
	}

	// Prevent self-demotion
	currentUserIDStr := c.Request().Header.Get(auth.HeaderUserID)
	currentUserID, _ := uuid.Parse(currentUserIDStr)
	if userID == currentUserID && req.UserType != "admin" {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot demote yourself")
	}

	ctx := c.Request().Context()

	// Check user exists
	user, err := h.store.GetUserByID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	err = h.store.UpdateUserType(ctx, store.UpdateUserTypeParams{
		ID:       userID,
		UserType: req.UserType,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user role")
	}

	return c.JSON(http.StatusOK, UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		UserType:  req.UserType,
		CreatedAt: user.CreatedAt.Time,
	})
}

// ResetUserPasswordRequest represents the request to reset a user's password.
type ResetUserPasswordRequest struct {
	Password string `json:"password"`
}

// ResetUserPassword resets a user's password and revokes all their sessions.
//
//	@Summary		Reset user password
//	@Description	Reset a user's password and revoke all their active sessions.
//	@Tags			Admin - Users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"User ID"
//	@Param			body	body		ResetUserPasswordRequest	true	"New password"
//	@Success		200		{object}	MessageResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users/{id}/password [put]
func (h *AdminHandler) ResetUserPassword(c *echo.Context) error {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	var req ResetUserPasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if len(req.Password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "password must be at least 8 characters")
	}

	ctx := c.Request().Context()

	// Check user exists
	_, err = h.store.GetUserByID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	// Hash new password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
	}

	err = h.store.UpdateUserPassword(ctx, store.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: pgtype.Text{String: passwordHash, Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to reset password")
	}

	// Revoke all refresh tokens to force re-login
	err = h.store.RevokeAllUserRefreshTokens(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to revoke sessions")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "password reset successfully"})
}

// CleanupExpiredTokens hard-deletes all expired tokens.
//
//	@Summary		Cleanup expired tokens
//	@Description	Hard-delete all expired refresh tokens from the database.
//	@Tags			Admin - Tokens
//	@Produce		json
//	@Success		200	{object}	CleanupResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/tokens/expired [delete]
func (h *AdminHandler) CleanupExpiredTokens(c *echo.Context) error {
	ctx := c.Request().Context()

	deleted, err := h.store.DeleteExpiredRefreshTokens(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to cleanup tokens")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "expired tokens cleaned up",
		"deleted": deleted,
	})
}

// StatCount is a generic labelled count used in stat breakdowns.
type StatCount struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// ProjectStat represents a project with its task count.
type ProjectStat struct {
	ProjectKey string `json:"project_key"`
	Name       string `json:"name"`
	TaskCount  int    `json:"task_count"`
}

// WorkspaceStat represents a workspace with its project count.
type WorkspaceStat struct {
	WorkspaceKey string `json:"workspace_key"`
	Name         string `json:"name"`
	ProjectCount int    `json:"project_count"`
}

// DayCount is a single day bucket in a per-day time series.
type DayCount struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

// ViewDayCount is a single day bucket split by view visibility.
type ViewDayCount struct {
	Day     string `json:"day"`
	Private int    `json:"private"`
	Shared  int    `json:"shared"`
}

// AdminStatsResponse is the aggregate stats payload for the admin stats page.
type AdminStatsResponse struct {
	Totals struct {
		Workspaces       int64 `json:"workspaces"`
		Projects         int64 `json:"projects"`
		Tasks            int64 `json:"tasks"`
		Subtasks         int64 `json:"subtasks"`
		Pages            int64 `json:"pages"`
		Users            int64 `json:"users"`
		Attachments      int64 `json:"attachments"`
		AttachmentsBytes int64 `json:"attachments_bytes"`
	} `json:"totals"`
	TasksByState         []StatCount     `json:"tasks_by_state"`
	TasksByPriority      []StatCount     `json:"tasks_by_priority"`
	TopProjects          []ProjectStat   `json:"top_projects"`
	ProjectsPerWorkspace []WorkspaceStat `json:"projects_per_workspace"`
	Series               struct {
		From     string     `json:"from"`
		To       string     `json:"to"`
		Days     int        `json:"days"`
		Tasks       []DayCount     `json:"tasks"`
		Subtasks    []DayCount     `json:"subtasks"`
		Pages       []DayCount     `json:"pages"`
		Views       []ViewDayCount `json:"views"`
		Comments    []DayCount     `json:"comments"`
		Activity    []DayCount     `json:"activity"`
		Attachments []DayCount     `json:"attachments"`
		Cycles      []DayCount     `json:"cycles"`
		Modules     []DayCount     `json:"modules"`
	} `json:"series"`
}

// priorityLabels maps task priority integers to human-readable labels.
var priorityLabels = map[int]string{
	0: "No priority",
	1: "Low",
	2: "Medium",
	3: "High",
	4: "Urgent",
}

// GetStats returns aggregate metrics for the admin stats dashboard.
//
//	@Summary		Get admin stats
//	@Description	Returns system-wide totals, breakdowns, and per-day activity series.
//	@Tags			Admin - Stats
//	@Produce		json
//	@Param			from	query		string	false	"Start date (YYYY-MM-DD) for per-day series"
//	@Param			to		query		string	false	"End date (YYYY-MM-DD) for per-day series"
//	@Success		200		{object}	AdminStatsResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/stats [get]
func (h *AdminHandler) GetStats(c *echo.Context) error {
	ctx := c.Request().Context()

	const dateLayout = "2006-01-02"
	const maxSpanDays = 366

	// Default range: trailing 30 days ending today.
	today := time.Now().Truncate(24 * time.Hour)
	toDate := today
	fromDate := today.AddDate(0, 0, -29)

	if v := c.QueryParam("to"); v != "" {
		if parsed, err := time.Parse(dateLayout, v); err == nil {
			toDate = parsed
		}
	}
	if v := c.QueryParam("from"); v != "" {
		if parsed, err := time.Parse(dateLayout, v); err == nil {
			fromDate = parsed
		}
	}

	// Normalise: ensure from <= to and clamp the span to a sane maximum.
	if fromDate.After(toDate) {
		fromDate, toDate = toDate, fromDate
	}
	if toDate.Sub(fromDate) > maxSpanDays*24*time.Hour {
		fromDate = toDate.AddDate(0, 0, -(maxSpanDays - 1))
	}

	var resp AdminStatsResponse
	var err error

	if resp.Totals.Workspaces, err = h.store.CountWorkspaces(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count workspaces")
	}
	if resp.Totals.Projects, err = h.store.CountProjects(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count projects")
	}
	if resp.Totals.Tasks, err = h.store.CountTopLevelTasks(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count tasks")
	}
	if resp.Totals.Subtasks, err = h.store.CountSubtasks(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count subtasks")
	}
	if resp.Totals.Pages, err = h.store.CountPages(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count pages")
	}
	if resp.Totals.Users, err = h.store.CountUsers(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count users")
	}
	if resp.Totals.Attachments, err = h.store.CountAttachments(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count attachments")
	}
	if resp.Totals.AttachmentsBytes, err = h.store.AttachmentsTotalSize(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to sum attachment sizes")
	}

	states, err := h.store.TasksByStateType(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load task states")
	}
	resp.TasksByState = make([]StatCount, len(states))
	for i, s := range states {
		resp.TasksByState[i] = StatCount{Label: s.StateType, Count: int(s.Count)}
	}

	priorities, err := h.store.TasksByPriority(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load task priorities")
	}
	resp.TasksByPriority = make([]StatCount, len(priorities))
	for i, p := range priorities {
		label, ok := priorityLabels[int(p.Priority)]
		if !ok {
			label = "Unknown"
		}
		resp.TasksByPriority[i] = StatCount{Label: label, Count: int(p.Count)}
	}

	topProjects, err := h.store.TopProjectsByTaskCount(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load top projects")
	}
	resp.TopProjects = make([]ProjectStat, len(topProjects))
	for i, p := range topProjects {
		resp.TopProjects[i] = ProjectStat{ProjectKey: p.ProjectKey, Name: p.Name, TaskCount: int(p.TaskCount)}
	}

	perWorkspace, err := h.store.ProjectsPerWorkspace(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load projects per workspace")
	}
	resp.ProjectsPerWorkspace = make([]WorkspaceStat, len(perWorkspace))
	for i, w := range perWorkspace {
		resp.ProjectsPerWorkspace[i] = WorkspaceStat{WorkspaceKey: w.WorkspaceKey, Name: w.Name, ProjectCount: int(w.ProjectCount)}
	}

	from := pgtype.Date{Time: fromDate, Valid: true}
	to := pgtype.Date{Time: toDate, Valid: true}

	resp.Series.From = fromDate.Format(dateLayout)
	resp.Series.To = toDate.Format(dateLayout)
	resp.Series.Days = int(toDate.Sub(fromDate).Hours()/24) + 1

	tasksSeries, err := h.store.TasksCreatedPerDay(ctx, store.TasksCreatedPerDayParams{FromDate: from, ToDate: to})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load tasks series")
	}
	resp.Series.Tasks = make([]DayCount, len(tasksSeries))
	for i, d := range tasksSeries {
		resp.Series.Tasks[i] = DayCount{Day: d.Day.Time.Format(dateLayout), Count: int(d.Count)}
	}

	subtasksSeries, err := h.store.SubtasksCreatedPerDay(ctx, store.SubtasksCreatedPerDayParams{FromDate: from, ToDate: to})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load subtasks series")
	}
	resp.Series.Subtasks = make([]DayCount, len(subtasksSeries))
	for i, d := range subtasksSeries {
		resp.Series.Subtasks[i] = DayCount{Day: d.Day.Time.Format(dateLayout), Count: int(d.Count)}
	}

	pagesSeries, err := h.store.PagesCreatedPerDay(ctx, store.PagesCreatedPerDayParams{FromDate: from, ToDate: to})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load pages series")
	}
	resp.Series.Pages = make([]DayCount, len(pagesSeries))
	for i, d := range pagesSeries {
		resp.Series.Pages[i] = DayCount{Day: d.Day.Time.Format(dateLayout), Count: int(d.Count)}
	}

	viewsSeries, err := h.store.ViewsCreatedPerDay(ctx, store.ViewsCreatedPerDayParams{FromDate: from, ToDate: to})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load views series")
	}
	resp.Series.Views = make([]ViewDayCount, len(viewsSeries))
	for i, d := range viewsSeries {
		resp.Series.Views[i] = ViewDayCount{
			Day:     d.Day.Time.Format(dateLayout),
			Private: int(d.PrivateCount),
			Shared:  int(d.SharedCount),
		}
	}

	commentsSeries, err := h.store.CommentsCreatedPerDay(ctx, store.CommentsCreatedPerDayParams{FromDate: from, ToDate: to})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load comments series")
	}
	resp.Series.Comments = make([]DayCount, len(commentsSeries))
	for i, d := range commentsSeries {
		resp.Series.Comments[i] = DayCount{Day: d.Day.Time.Format(dateLayout), Count: int(d.Count)}
	}

	activitySeries, err := h.store.ActivityCreatedPerDay(ctx, store.ActivityCreatedPerDayParams{FromDate: from, ToDate: to})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load activity series")
	}
	resp.Series.Activity = make([]DayCount, len(activitySeries))
	for i, d := range activitySeries {
		resp.Series.Activity[i] = DayCount{Day: d.Day.Time.Format(dateLayout), Count: int(d.Count)}
	}

	attachmentsSeries, err := h.store.AttachmentsCreatedPerDay(ctx, store.AttachmentsCreatedPerDayParams{FromDate: from, ToDate: to})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load attachments series")
	}
	resp.Series.Attachments = make([]DayCount, len(attachmentsSeries))
	for i, d := range attachmentsSeries {
		resp.Series.Attachments[i] = DayCount{Day: d.Day.Time.Format(dateLayout), Count: int(d.Count)}
	}

	cyclesSeries, err := h.store.CyclesCreatedPerDay(ctx, store.CyclesCreatedPerDayParams{FromDate: from, ToDate: to})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load cycles series")
	}
	resp.Series.Cycles = make([]DayCount, len(cyclesSeries))
	for i, d := range cyclesSeries {
		resp.Series.Cycles[i] = DayCount{Day: d.Day.Time.Format(dateLayout), Count: int(d.Count)}
	}

	modulesSeries, err := h.store.ModulesCreatedPerDay(ctx, store.ModulesCreatedPerDayParams{FromDate: from, ToDate: to})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load modules series")
	}
	resp.Series.Modules = make([]DayCount, len(modulesSeries))
	for i, d := range modulesSeries {
		resp.Series.Modules[i] = DayCount{Day: d.Day.Time.Format(dateLayout), Count: int(d.Count)}
	}

	return c.JSON(http.StatusOK, resp)
}

// GraphUser is a user node in the admin task graph.
type GraphUser struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// GraphTask is a task node in the admin task graph.
type GraphTask struct {
	ID            string     `json:"id"`
	ProjectKey    string     `json:"project_key"`
	ProjectName   string     `json:"project_name"`
	TaskNumber    int32      `json:"task_number"`
	Title         string     `json:"title"`
	IsSubtask     bool       `json:"is_subtask"`
	ParentID      *uuid.UUID `json:"parent_id,omitempty"`
	WorkspaceID   string     `json:"workspace_id"`
	WorkspaceKey  string     `json:"workspace_key"`
	WorkspaceName string     `json:"workspace_name"`
	StateName     string     `json:"state_name"`
	StateType     string     `json:"state_type"`
	StateColor    *string    `json:"state_color,omitempty"`
}

// GraphEdge links a user to a task they are assigned to.
type GraphEdge struct {
	UserID string `json:"user_id"`
	TaskID string `json:"task_id"`
}

// GraphBlockerEdge links a blocker task to the task it blocks.
type GraphBlockerEdge struct {
	BlockerID string `json:"blocker_id"`
	BlockedID string `json:"blocked_id"`
}

// GraphWorkspaceOption is a workspace choice for the graph view filters.
type GraphWorkspaceOption struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// GraphProjectOption is a project choice for the graph view filters.
type GraphProjectOption struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	WorkspaceKey string `json:"workspace_key"`
}

// GraphUserOption is a user choice for the graph view filters.
type GraphUserOption struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// TaskGraphFiltersResponse lists the choices for the graph view filters.
type TaskGraphFiltersResponse struct {
	Workspaces []GraphWorkspaceOption `json:"workspaces"`
	Projects   []GraphProjectOption   `json:"projects"`
	Users      []GraphUserOption      `json:"users"`
}

// TaskGraphResponse is the payload for the admin graph view.
type TaskGraphResponse struct {
	Users        []GraphUser        `json:"users"`
	Tasks        []GraphTask        `json:"tasks"`
	Edges        []GraphEdge        `json:"edges"`
	BlockerEdges []GraphBlockerEdge `json:"blocker_edges"`
}

// splitCSVParam splits a comma-separated query value, dropping empty entries.
func splitCSVParam(v string) []string {
	out := []string{}
	for _, s := range strings.Split(v, ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// GetTaskGraphFilters returns the workspace, project and user choices for the graph view.
//
//	@Summary		Get admin task graph filters
//	@Description	Returns workspace, project and user options for the admin graph view filters.
//	@Tags			Admin - Stats
//	@Produce		json
//	@Success		200	{object}	TaskGraphFiltersResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/graph/filters [get]
func (h *AdminHandler) GetTaskGraphFilters(c *echo.Context) error {
	return h.taskGraphFilters(c, pgtype.UUID{})
}

// GetMyTaskGraphFilters returns graph view filter choices limited to the caller's member projects.
//
//	@Summary		Get task graph filters
//	@Description	Returns workspace, project and user options for the graph view, limited to projects the caller is a member of.
//	@Tags			Graph
//	@Produce		json
//	@Success		200	{object}	TaskGraphFiltersResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/graph/filters [get]
func (h *AdminHandler) GetMyTaskGraphFilters(c *echo.Context) error {
	viewer, err := graphViewer(c)
	if err != nil {
		return err
	}
	return h.taskGraphFilters(c, viewer)
}

// graphViewer returns the caller's ID for scoping graph queries to their member projects.
func graphViewer(c *echo.Context) (pgtype.UUID, error) {
	userID, err := uuid.Parse(c.Request().Header.Get(auth.HeaderUserID))
	if err != nil {
		return pgtype.UUID{}, echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}
	return pgtype.UUID{Bytes: userID, Valid: true}, nil
}

// taskGraphFilters lists filter choices; a null viewer means no membership scoping.
func (h *AdminHandler) taskGraphFilters(c *echo.Context, viewer pgtype.UUID) error {
	ctx := c.Request().Context()

	workspaces, err := h.store.ListGraphWorkspaceOptions(ctx, viewer)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load workspaces")
	}
	projects, err := h.store.ListGraphProjectOptions(ctx, viewer)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load projects")
	}
	users, err := h.store.ListGraphUserOptions(ctx, viewer)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load users")
	}

	resp := TaskGraphFiltersResponse{
		Workspaces: make([]GraphWorkspaceOption, len(workspaces)),
		Projects:   make([]GraphProjectOption, len(projects)),
		Users:      make([]GraphUserOption, len(users)),
	}
	for i, w := range workspaces {
		resp.Workspaces[i] = GraphWorkspaceOption{Key: w.WorkspaceKey, Name: w.Name}
	}
	for i, p := range projects {
		resp.Projects[i] = GraphProjectOption{Key: p.ProjectKey, Name: p.Name, WorkspaceKey: p.WorkspaceKey}
	}
	for i, u := range users {
		resp.Users[i] = GraphUserOption{Username: u.Username, Email: u.Email, FirstName: u.FirstName, LastName: u.LastName}
	}

	return c.JSON(http.StatusOK, resp)
}

// GetTaskGraph returns users and the tasks they are assigned to.
//
//	@Summary		Get admin task graph
//	@Description	Returns user and task nodes with assignment edges for the admin graph view.
//	@Tags			Admin - Stats
//	@Produce		json
//	@Param			workspace	query		string	false	"Workspace key"
//	@Param			projects	query		string	false	"Comma-separated project keys"
//	@Param			state_types	query		string	false	"Comma-separated state types"
//	@Param			users		query		string	false	"Comma-separated usernames"
//	@Success		200			{object}	TaskGraphResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/graph [get]
func (h *AdminHandler) GetTaskGraph(c *echo.Context) error {
	return h.taskGraph(c, pgtype.UUID{})
}

// GetMyTaskGraph returns the task graph limited to the caller's member projects.
//
//	@Summary		Get task graph
//	@Description	Returns user and task nodes with assignment edges, limited to projects the caller is a member of.
//	@Tags			Graph
//	@Produce		json
//	@Param			workspace	query		string	false	"Workspace key"
//	@Param			projects	query		string	false	"Comma-separated project keys"
//	@Param			state_types	query		string	false	"Comma-separated state types"
//	@Param			users		query		string	false	"Comma-separated usernames"
//	@Success		200			{object}	TaskGraphResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/graph [get]
func (h *AdminHandler) GetMyTaskGraph(c *echo.Context) error {
	viewer, err := graphViewer(c)
	if err != nil {
		return err
	}
	return h.taskGraph(c, viewer)
}

// taskGraph builds the graph payload; a null viewer means no membership scoping.
func (h *AdminHandler) taskGraph(c *echo.Context, viewer pgtype.UUID) error {
	ctx := c.Request().Context()

	params := store.ListTaskAssignmentsForGraphParams{
		ProjectKeys: splitCSVParam(c.QueryParam("projects")),
		StateTypes:  splitCSVParam(c.QueryParam("state_types")),
		Usernames:   splitCSVParam(c.QueryParam("users")),
		ViewerID:    viewer,
	}
	if ws := strings.TrimSpace(c.QueryParam("workspace")); ws != "" {
		params.WorkspaceKey = pgtype.Text{String: ws, Valid: true}
	}

	rows, err := h.store.ListTaskAssignmentsForGraph(ctx, params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load task graph")
	}

	resp := TaskGraphResponse{Users: []GraphUser{}, Tasks: []GraphTask{}, Edges: []GraphEdge{}, BlockerEdges: []GraphBlockerEdge{}}
	taskIDs := []uuid.UUID{}
	seenUsers := make(map[string]bool)
	seenTasks := make(map[string]bool)

	for _, r := range rows {
		userID := r.UserID.String()
		taskID := r.TaskID.String()

		if !seenUsers[userID] {
			seenUsers[userID] = true
			resp.Users = append(resp.Users, GraphUser{
				ID:        userID,
				Username:  r.Username,
				Email:     r.Email,
				FirstName: r.FirstName,
				LastName:  r.LastName,
				AvatarURL: textToStringPtr(r.AvatarUrl),
			})
		}
		if !seenTasks[taskID] {
			seenTasks[taskID] = true
			taskIDs = append(taskIDs, r.TaskID)
			resp.Tasks = append(resp.Tasks, GraphTask{
				ID:            taskID,
				ProjectKey:    r.ProjectKey,
				ProjectName:   r.ProjectName,
				TaskNumber:    r.TaskNumber,
				Title:         r.Title,
				IsSubtask:     r.IsSubtask,
				ParentID:      pgUUIDToUUIDPtr(r.ParentTaskID),
				WorkspaceID:   r.WorkspaceID.String(),
				WorkspaceKey:  r.WorkspaceKey,
				WorkspaceName: r.WorkspaceName,
				StateName:     r.StateName,
				StateType:     r.StateType,
				StateColor:    textToStringPtr(r.StateColor),
			})
		}
		resp.Edges = append(resp.Edges, GraphEdge{UserID: userID, TaskID: taskID})
	}

	if len(taskIDs) > 0 {
		links, err := h.store.ListBlockerLinksForGraph(ctx, taskIDs)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to load task graph")
		}
		for _, l := range links {
			resp.BlockerEdges = append(resp.BlockerEdges, GraphBlockerEdge{
				BlockerID: l.BlockerTaskID.String(),
				BlockedID: l.TaskID.String(),
			})
		}
	}

	return c.JSON(http.StatusOK, resp)
}

// DeletedProjectResponse represents a soft-deleted project in admin responses.
type DeletedProjectResponse struct {
	ID              uuid.UUID `json:"id"`
	ProjectKey      string    `json:"project_key"`
	Name            string    `json:"name"`
	Description     *string   `json:"description,omitempty"`
	WorkspaceID     uuid.UUID `json:"workspace_id"`
	WorkspaceName   string    `json:"workspace_name"`
	CreatedBy       uuid.UUID `json:"created_by"`
	CreatorName     string    `json:"creator_name"`
	CreatorUsername string    `json:"creator_username"`
	CreatedAt       time.Time `json:"created_at"`
	DeletedAt       time.Time `json:"deleted_at"`
}

// PaginatedDeletedProjectsResponse represents a paginated list of deleted projects.
type PaginatedDeletedProjectsResponse struct {
	Projects   []DeletedProjectResponse `json:"projects"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	PerPage    int                      `json:"per_page"`
	TotalPages int                      `json:"total_pages"`
}

// ListDeletedProjects returns a paginated list of soft-deleted projects so an
// admin can review and restore them.
//
//	@Summary		List deleted projects
//	@Description	Returns a paginated list of soft-deleted projects across all workspaces.
//	@Tags			Admin - Projects
//	@Produce		json
//	@Param			page		query		int	false	"Page number"		default(1)
//	@Param			per_page	query		int	false	"Items per page"	default(20)
//	@Success		200			{object}	PaginatedDeletedProjectsResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/projects/deleted [get]
func (h *AdminHandler) ListDeletedProjects(c *echo.Context) error {
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

	total, err := h.store.CountDeletedProjects(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count deleted projects")
	}

	rows, err := h.store.ListDeletedProjects(ctx, store.ListDeletedProjectsParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list deleted projects")
	}

	projects := make([]DeletedProjectResponse, len(rows))
	for i, p := range rows {
		projects[i] = DeletedProjectResponse{
			ID:              p.ID,
			ProjectKey:      p.ProjectKey,
			Name:            p.Name,
			Description:     textToStringPtr(p.Description),
			WorkspaceID:     p.WorkspaceID,
			WorkspaceName:   p.WorkspaceName,
			CreatedBy:       p.CreatedBy,
			CreatorName:     strings.TrimSpace(p.CreatorFirstName + " " + p.CreatorLastName),
			CreatorUsername: p.CreatorUsername,
			CreatedAt:       p.CreatedAt.Time,
			DeletedAt:       p.DeletedAt.Time,
		}
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	return c.JSON(http.StatusOK, PaginatedDeletedProjectsResponse{
		Projects:   projects,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// RestoreProject restores a soft-deleted project.
//
//	@Summary		Restore deleted project
//	@Description	Clears the soft-delete flag on a project, making it accessible again.
//	@Tags			Admin - Projects
//	@Produce		json
//	@Param			id	path		string	true	"Project ID"
//	@Success		200	{object}	MessageResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/projects/{id}/restore [post]
func (h *AdminHandler) RestoreProject(c *echo.Context) error {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project ID")
	}

	ctx := c.Request().Context()

	if err := h.store.RestoreProject(ctx, projectID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to restore project")
	}

	return c.JSON(http.StatusOK, MessageResponse{Message: "project restored"})
}

// MergeUserRequest is the body for POST /admin/users/:id/merge.
type MergeUserRequest struct {
	TargetUserID uuid.UUID `json:"target_user_id"`
}

// MergeUserResponse reports how many rows were copied to the target user.
type MergeUserResponse struct {
	Workspaces   int64 `json:"workspaces"`
	Projects     int64 `json:"projects"`
	Assignments  int64 `json:"assignments"`
	Modules      int64 `json:"modules"`
	FollowedTasks int64 `json:"followed_tasks"`
	Views        int64 `json:"views"`
}

// MergeUser grants the target user everything the source user (:id) has.
//
//	@Summary		Merge user
//	@Description	Additively copies the source user's workspace/project/module memberships, task assignments and private views to the target user, and makes the target follow tasks the source created, follows, or commented on. The source user is left unchanged.
//	@Tags			Admin - Users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Source user ID"
//	@Param			body	body		MergeUserRequest	true	"Target user"
//	@Success		200		{object}	MergeUserResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users/{id}/merge [post]
func (h *AdminHandler) MergeUser(c *echo.Context) error {
	sourceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	var req MergeUserRequest
	if err := c.Bind(&req); err != nil || req.TargetUserID == uuid.Nil {
		return echo.NewHTTPError(http.StatusBadRequest, "target_user_id is required")
	}
	if req.TargetUserID == sourceID {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot merge a user into themselves")
	}

	ctx := c.Request().Context()

	if _, err := h.store.GetUserByID(ctx, sourceID); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "source user not found")
	}
	if _, err := h.store.GetUserByID(ctx, req.TargetUserID); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "target user not found")
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to merge user")
	}
	defer tx.Rollback(ctx)
	q := store.New(tx)

	var resp MergeUserResponse
	steps := []struct {
		count *int64
		run   func() (int64, error)
	}{
		{&resp.Workspaces, func() (int64, error) {
			return q.MergeCopyWorkspaceMembers(ctx, store.MergeCopyWorkspaceMembersParams{TargetID: req.TargetUserID, SourceID: sourceID})
		}},
		{&resp.Projects, func() (int64, error) {
			return q.MergeCopyProjectMembers(ctx, store.MergeCopyProjectMembersParams{TargetID: req.TargetUserID, SourceID: sourceID})
		}},
		{&resp.Assignments, func() (int64, error) {
			return q.MergeCopyTaskAssignees(ctx, store.MergeCopyTaskAssigneesParams{TargetID: req.TargetUserID, SourceID: sourceID})
		}},
		{&resp.Modules, func() (int64, error) {
			return q.MergeCopyModuleMembers(ctx, store.MergeCopyModuleMembersParams{TargetID: req.TargetUserID, SourceID: sourceID})
		}},
		{&resp.FollowedTasks, func() (int64, error) {
			return q.MergeFollowSourceTasks(ctx, store.MergeFollowSourceTasksParams{TargetID: req.TargetUserID, SourceID: sourceID})
		}},
		{&resp.Views, func() (int64, error) {
			return q.MergeCopyPrivateViews(ctx, store.MergeCopyPrivateViewsParams{TargetID: req.TargetUserID, SourceID: sourceID})
		}},
	}
	for _, step := range steps {
		n, err := step.run()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to merge user")
		}
		*step.count = n
	}

	if err := tx.Commit(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to merge user")
	}

	return c.JSON(http.StatusOK, resp)
}
