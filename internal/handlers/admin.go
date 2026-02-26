package handlers

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/auth"
	"bereaucat/internal/store"
)

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
