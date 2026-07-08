package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	authManager *auth.Manager
	devMode     bool
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(store store.Querier, authManager *auth.Manager, devMode bool) *AdminHandler {
	return &AdminHandler{
		store:       store,
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

// AdminStatsResponse is the aggregate stats payload for the admin stats page.
type AdminStatsResponse struct {
	Totals struct {
		Workspaces int64 `json:"workspaces"`
		Projects   int64 `json:"projects"`
		Tasks      int64 `json:"tasks"`
		Subtasks   int64 `json:"subtasks"`
		Pages      int64 `json:"pages"`
		Users      int64 `json:"users"`
	} `json:"totals"`
	TasksByState         []StatCount     `json:"tasks_by_state"`
	TasksByPriority      []StatCount     `json:"tasks_by_priority"`
	TopProjects          []ProjectStat   `json:"top_projects"`
	ProjectsPerWorkspace []WorkspaceStat `json:"projects_per_workspace"`
	Series               struct {
		From     string     `json:"from"`
		To       string     `json:"to"`
		Days     int        `json:"days"`
		Tasks    []DayCount `json:"tasks"`
		Subtasks []DayCount `json:"subtasks"`
		Pages    []DayCount `json:"pages"`
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

	return c.JSON(http.StatusOK, resp)
}
