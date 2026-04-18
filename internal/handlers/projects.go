package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/auth"
	"bereaucat/internal/store"
)

// ProjectHandler handles project-related endpoints.
type ProjectHandler struct {
	store store.Querier
}

// NewProjectHandler creates a new project handler.
func NewProjectHandler(store store.Querier) *ProjectHandler {
	return &ProjectHandler{
		store: store,
	}
}

// ProjectResponse represents a project in API responses.
type ProjectResponse struct {
	ID          uuid.UUID  `json:"id"`
	ProjectKey  string     `json:"project_key"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	IconURL     *string    `json:"icon_url,omitempty"`
	CoverURL    *string    `json:"cover_url,omitempty"`
	Role        string     `json:"role,omitempty"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateProjectRequest represents the request to create a project.
type CreateProjectRequest struct {
	ProjectKey  string  `json:"project_key"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IconID      *string `json:"icon_id"`
	CoverID     *string `json:"cover_id"`
}

// UpdateProjectRequest represents the request to update a project.
type UpdateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IconID      *string `json:"icon_id"`
	CoverID     *string `json:"cover_id"`
}

// PaginatedProjectsResponse represents a paginated list of projects.
type PaginatedProjectsResponse struct {
	Projects   []ProjectResponse `json:"projects"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	TotalPages int               `json:"total_pages"`
}

// Default states to create when a new project is created
var defaultStates = []struct {
	StateType string
	Name      string
	Color     string
	Position  int
	IsDefault bool
}{
	{"backlog", "Backlog", "#6B7280", 0, true},
	{"unstarted", "Todo", "#3B82F6", 1, false},
	{"unstarted", "Approval Pending", "#F59E0B", 2, false},
	{"started", "In Progress", "#10B981", 3, false},
	{"started", "Blocked", "#EF4444", 4, false},
	{"started", "Testing", "#8B5CF6", 5, false},
	{"completed", "Done", "#22C55E", 6, false},
	{"cancelled", "Cancelled", "#9CA3AF", 7, false},
}

// ListProjects returns paginated list of projects with optional search.
//
//	@Summary		List projects
//	@Description	Returns a paginated list of projects. Admins see all projects; others see only their memberships.
//	@Tags			Projects
//	@Produce		json
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(20)
//	@Param			search		query		string	false	"Search by name"
//	@Success		200			{object}	PaginatedProjectsResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects [get]
func (h *ProjectHandler) ListProjects(c *echo.Context) error {
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

	search := strings.TrimSpace(c.QueryParam("search"))
	searchParam := pgtype.Text{}
	if search != "" {
		searchParam = pgtype.Text{String: search, Valid: true}
	}

	ctx := c.Request().Context()
	userType := c.Request().Header.Get(auth.HeaderUserType)

	var total int64
	var projectResponses []ProjectResponse

	if userType == "admin" {
		total, err = h.store.CountAllProjectsFiltered(ctx, searchParam)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to count projects")
		}
		projects, err := h.store.ListAllProjectsFiltered(ctx, store.ListAllProjectsFilteredParams{
			Limit:  int32(perPage),
			Offset: int32(offset),
			Search: searchParam,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list projects")
		}
		projectResponses = make([]ProjectResponse, len(projects))
		for i, p := range projects {
			projectResponses[i] = ProjectResponse{
				ID:          p.ID,
				ProjectKey:  p.ProjectKey,
				Name:        p.Name,
				Description: textToStringPtr(p.Description),
				IconURL:     pgtypeUUIDToURL(p.IconID),
				CoverURL:    pgtypeUUIDToURL(p.CoverID),
				Role:        p.Role,
				CreatedBy:   p.CreatedBy,
				CreatedAt:   p.CreatedAt.Time,
				UpdatedAt:   p.UpdatedAt.Time,
			}
		}
	} else {
		total, err = h.store.CountUserProjectsFiltered(ctx, store.CountUserProjectsFilteredParams{
			UserID: userID,
			Search: searchParam,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to count projects")
		}
		projects, err := h.store.ListUserProjectsFiltered(ctx, store.ListUserProjectsFilteredParams{
			UserID: userID,
			Limit:  int32(perPage),
			Offset: int32(offset),
			Search: searchParam,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list projects")
		}
		projectResponses = make([]ProjectResponse, len(projects))
		for i, p := range projects {
			projectResponses[i] = ProjectResponse{
				ID:          p.ID,
				ProjectKey:  p.ProjectKey,
				Name:        p.Name,
				Description: textToStringPtr(p.Description),
				IconURL:     pgtypeUUIDToURL(p.IconID),
				CoverURL:    pgtypeUUIDToURL(p.CoverID),
				Role:        p.Role,
				CreatedBy:   p.CreatedBy,
				CreatedAt:   p.CreatedAt.Time,
				UpdatedAt:   p.UpdatedAt.Time,
			}
		}
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, PaginatedProjectsResponse{
		Projects:   projectResponses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// CreateProject creates a new project.
//
//	@Summary		Create project
//	@Description	Create a new project with default workflow states. Creator is added as admin.
//	@Tags			Projects
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateProjectRequest	true	"Project details"
//	@Success		201		{object}	ProjectResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects [post]
func (h *ProjectHandler) CreateProject(c *echo.Context) error {
	userIDStr := c.Request().Header.Get(auth.HeaderUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	var req CreateProjectRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate required fields
	if req.ProjectKey == "" || req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "project_key and name are required")
	}

	// Validate project key format (alphanumeric, uppercase, 2-10 chars)
	req.ProjectKey = strings.ToUpper(req.ProjectKey)
	if len(req.ProjectKey) < 2 || len(req.ProjectKey) > 10 {
		return echo.NewHTTPError(http.StatusBadRequest, "project_key must be 2-10 characters")
	}
	for _, r := range req.ProjectKey {
		if !((r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return echo.NewHTTPError(http.StatusBadRequest, "project_key must be alphanumeric")
		}
	}

	ctx := c.Request().Context()

	// Check if project key already exists
	exists, err := h.store.ProjectKeyExists(ctx, req.ProjectKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check project key")
	}
	if exists {
		return echo.NewHTTPError(http.StatusConflict, "project key already exists")
	}

	// Parse optional UUIDs
	iconID := stringToPgtypeUUID(req.IconID)
	coverID := stringToPgtypeUUID(req.CoverID)

	// Create project
	project, err := h.store.CreateProject(ctx, store.CreateProjectParams{
		ProjectKey:  req.ProjectKey,
		Name:        req.Name,
		Description: stringToPgtypeText(req.Description),
		IconID:      iconID,
		CoverID:     coverID,
		CreatedBy:   userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create project")
	}

	// Add creator as admin member
	_, err = h.store.AddProjectMember(ctx, store.AddProjectMemberParams{
		ProjectID: project.ID,
		UserID:    userID,
		Role:      "admin",
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add creator as member")
	}

	// Create default states
	for _, state := range defaultStates {
		_, err := h.store.CreateProjectState(ctx, store.CreateProjectStateParams{
			ProjectID: project.ID,
			StateType: state.StateType,
			Name:      state.Name,
			Color:     pgtype.Text{String: state.Color, Valid: true},
			Position:  int32(state.Position),
			IsDefault: state.IsDefault,
		})
		if err != nil {
			// Log but don't fail on state creation errors
			continue
		}
	}

	return c.JSON(http.StatusCreated, ProjectResponse{
		ID:          project.ID,
		ProjectKey:  project.ProjectKey,
		Name:        project.Name,
		Description: textToStringPtr(project.Description),
		IconURL:     pgtypeUUIDToURL(project.IconID),
		CoverURL:    pgtypeUUIDToURL(project.CoverID),
		Role:        "admin",
		CreatedBy:   project.CreatedBy,
		CreatedAt:   project.CreatedAt.Time,
		UpdatedAt:   project.UpdatedAt.Time,
	})
}

// GetProject returns project details.
//
//	@Summary		Get project
//	@Description	Returns project details by project key.
//	@Tags			Projects
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Success		200			{object}	ProjectResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey} [get]
func (h *ProjectHandler) GetProject(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	role := c.Request().Header.Get(auth.HeaderProjectRole)

	ctx := c.Request().Context()

	project, err := h.store.GetProjectByID(ctx, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}

	return c.JSON(http.StatusOK, ProjectResponse{
		ID:          project.ID,
		ProjectKey:  project.ProjectKey,
		Name:        project.Name,
		Description: textToStringPtr(project.Description),
		IconURL:     pgtypeUUIDToURL(project.IconID),
		CoverURL:    pgtypeUUIDToURL(project.CoverID),
		Role:        role,
		CreatedBy:   project.CreatedBy,
		CreatedAt:   project.CreatedAt.Time,
		UpdatedAt:   project.UpdatedAt.Time,
	})
}

// UpdateProject updates a project.
//
//	@Summary		Update project
//	@Description	Update project details. Requires project admin role.
//	@Tags			Projects
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string					true	"Project key"
//	@Param			body		body		UpdateProjectRequest	true	"Fields to update"
//	@Success		200			{object}	ProjectResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey} [patch]
func (h *ProjectHandler) UpdateProject(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	var req UpdateProjectRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ctx := c.Request().Context()

	project, err := h.store.UpdateProject(ctx, store.UpdateProjectParams{
		ID:          projectID,
		Name:        stringToPgtypeText(req.Name),
		Description: stringToPgtypeText(req.Description),
		IconID:      stringToPgtypeUUID(req.IconID),
		CoverID:     stringToPgtypeUUID(req.CoverID),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update project")
	}

	return c.JSON(http.StatusOK, ProjectResponse{
		ID:          project.ID,
		ProjectKey:  project.ProjectKey,
		Name:        project.Name,
		Description: textToStringPtr(project.Description),
		IconURL:     pgtypeUUIDToURL(project.IconID),
		CoverURL:    pgtypeUUIDToURL(project.CoverID),
		CreatedBy:   project.CreatedBy,
		CreatedAt:   project.CreatedAt.Time,
		UpdatedAt:   project.UpdatedAt.Time,
	})
}

// DeleteProject soft deletes a project.
//
//	@Summary		Delete project
//	@Description	Soft-delete a project. Requires project admin role.
//	@Tags			Projects
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Success		200			{object}	MessageResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey} [delete]
func (h *ProjectHandler) DeleteProject(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()

	err = h.store.SoftDeleteProject(ctx, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete project")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "project deleted"})
}

// MemberResponse represents a project member in API responses.
type MemberResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

// AddMemberRequest represents the request to add a member.
type AddMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// UpdateMemberRequest represents the request to update a member's role.
type UpdateMemberRequest struct {
	Role string `json:"role"`
}

// ListMembers returns project members.
//
//	@Summary		List members
//	@Description	Returns all members of a project.
//	@Tags			Project Members
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Success		200			{array}		MemberResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/members [get]
func (h *ProjectHandler) ListMembers(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()

	members, err := h.store.ListProjectMembers(ctx, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list members")
	}

	memberResponses := make([]MemberResponse, len(members))
	for i, m := range members {
		memberResponses[i] = MemberResponse{
			ID:        m.ID,
			UserID:    m.UserID,
			Username:  m.Username,
			Email:     m.Email,
			FirstName: m.FirstName,
			LastName:  m.LastName,
			AvatarURL: textToStringPtr(m.AvatarUrl),
			Role:      m.Role,
			JoinedAt:  m.JoinedAt.Time,
		}
	}

	return c.JSON(http.StatusOK, memberResponses)
}

// AddMember adds a member to a project.
//
//	@Summary		Add member
//	@Description	Add a user as a member of the project. Requires project admin role.
//	@Tags			Project Members
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string			true	"Project key"
//	@Param			body		body		AddMemberRequest	true	"Member details"
//	@Success		201			{object}	MemberResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		409			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/members [post]
func (h *ProjectHandler) AddMember(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	var req AddMemberRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}

	// Validate role
	if req.Role != "guest" && req.Role != "member" && req.Role != "admin" {
		return echo.NewHTTPError(http.StatusBadRequest, "role must be 'guest', 'member', or 'admin'")
	}

	ctx := c.Request().Context()

	// Check if user is already a member
	isMember, err := h.store.IsProjectMember(ctx, store.IsProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check membership")
	}
	if isMember {
		return echo.NewHTTPError(http.StatusConflict, "user is already a member")
	}

	member, err := h.store.AddProjectMember(ctx, store.AddProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
		Role:      req.Role,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add member")
	}

	// Get full member info
	fullMember, err := h.store.GetProjectMember(ctx, store.GetProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get member info")
	}

	return c.JSON(http.StatusCreated, MemberResponse{
		ID:        member.ID,
		UserID:    fullMember.UserID,
		Username:  fullMember.Username,
		Email:     fullMember.Email,
		FirstName: fullMember.FirstName,
		LastName:  fullMember.LastName,
		AvatarURL: textToStringPtr(fullMember.AvatarUrl),
		Role:      fullMember.Role,
		JoinedAt:  member.JoinedAt.Time,
	})
}

// UpdateMemberRole updates a member's role.
//
//	@Summary		Update member role
//	@Description	Update a project member's role. Requires project admin role.
//	@Tags			Project Members
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string				true	"Project key"
//	@Param			userId		path		string				true	"User ID"
//	@Param			body		body		UpdateMemberRequest	true	"New role"
//	@Success		200			{object}	MessageResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/members/{userId} [patch]
func (h *ProjectHandler) UpdateMemberRole(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	var req UpdateMemberRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate role
	if req.Role != "guest" && req.Role != "member" && req.Role != "admin" {
		return echo.NewHTTPError(http.StatusBadRequest, "role must be 'guest', 'member', or 'admin'")
	}

	ctx := c.Request().Context()

	// Check if user is a member
	isMember, err := h.store.IsProjectMember(ctx, store.IsProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check membership")
	}
	if !isMember {
		return echo.NewHTTPError(http.StatusNotFound, "member not found")
	}

	err = h.store.UpdateProjectMemberRole(ctx, store.UpdateProjectMemberRoleParams{
		ProjectID: projectID,
		UserID:    userID,
		Role:      req.Role,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update role")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "role updated"})
}

// RemoveMember removes a member from a project.
//
//	@Summary		Remove member
//	@Description	Remove a member from the project. Cannot remove yourself. Requires project admin role.
//	@Tags			Project Members
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			userId		path		string	true	"User ID"
//	@Success		200			{object}	MessageResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/members/{userId} [delete]
func (h *ProjectHandler) RemoveMember(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	// Prevent removing yourself
	currentUserIDStr := c.Request().Header.Get(auth.HeaderUserID)
	currentUserID, _ := uuid.Parse(currentUserIDStr)
	if userID == currentUserID {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot remove yourself")
	}

	ctx := c.Request().Context()

	err = h.store.RemoveProjectMember(ctx, store.RemoveProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove member")
	}

	// Reassign the departing member's shared views to the acting admin so they
	// don't get orphaned, and soft-delete their private views.
	_ = h.store.TransferOwnedSharedViews(ctx, store.TransferOwnedSharedViewsParams{
		ProjectID: projectID,
		OwnerID:   userID,
		OwnerID_2: currentUserID,
	})
	_ = h.store.SoftDeleteOwnedPrivateViews(ctx, store.SoftDeleteOwnedPrivateViewsParams{
		ProjectID: projectID,
		OwnerID:   userID,
	})

	return c.JSON(http.StatusOK, map[string]string{"message": "member removed"})
}

// StateResponse represents a project state in API responses.
type StateResponse struct {
	ID        uuid.UUID `json:"id"`
	StateType string    `json:"state_type"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	Position  int       `json:"position"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateStateRequest represents the request to create a state.
type CreateStateRequest struct {
	StateType string `json:"state_type"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	Position  *int   `json:"position"`
}

// UpdateStateRequest represents the request to update a state.
type UpdateStateRequest struct {
	Name     *string `json:"name"`
	Color    *string `json:"color"`
	Position *int    `json:"position"`
}

// ListStates returns project states.
//
//	@Summary		List states
//	@Description	Returns all workflow states for a project.
//	@Tags			Project States
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Success		200			{array}		StateResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/states [get]
func (h *ProjectHandler) ListStates(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()

	states, err := h.store.ListProjectStates(ctx, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list states")
	}

	stateResponses := make([]StateResponse, len(states))
	for i, s := range states {
		stateResponses[i] = StateResponse{
			ID:        s.ID,
			StateType: s.StateType,
			Name:      s.Name,
			Color:     textToString(s.Color, "#6B7280"),
			Position:  int(s.Position),
			IsDefault: s.IsDefault,
			CreatedAt: s.CreatedAt.Time,
		}
	}

	return c.JSON(http.StatusOK, stateResponses)
}

// CreateState creates a new state.
//
//	@Summary		Create state
//	@Description	Create a new workflow state. Requires project admin role.
//	@Tags			Project States
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string				true	"Project key"
//	@Param			body		body		CreateStateRequest	true	"State details"
//	@Success		201			{object}	StateResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/states [post]
func (h *ProjectHandler) CreateState(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	var req CreateStateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" || req.StateType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name and state_type are required")
	}

	// Validate state type
	validTypes := map[string]bool{"backlog": true, "unstarted": true, "started": true, "completed": true, "cancelled": true}
	if !validTypes[req.StateType] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid state_type")
	}

	if req.Color == "" {
		req.Color = "#6B7280"
	}

	position := int32(0)
	if req.Position != nil {
		position = int32(*req.Position)
	}

	ctx := c.Request().Context()

	state, err := h.store.CreateProjectState(ctx, store.CreateProjectStateParams{
		ProjectID: projectID,
		StateType: req.StateType,
		Name:      req.Name,
		Color:     pgtype.Text{String: req.Color, Valid: true},
		Position:  position,
		IsDefault: false,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create state")
	}

	return c.JSON(http.StatusCreated, StateResponse{
		ID:        state.ID,
		StateType: state.StateType,
		Name:      state.Name,
		Color:     textToString(state.Color, "#6B7280"),
		Position:  int(state.Position),
		IsDefault: state.IsDefault,
		CreatedAt: state.CreatedAt.Time,
	})
}

// UpdateState updates a state.
//
//	@Summary		Update state
//	@Description	Update a workflow state. Requires project admin role.
//	@Tags			Project States
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string				true	"Project key"
//	@Param			stateId		path		string				true	"State ID"
//	@Param			body		body		UpdateStateRequest	true	"Fields to update"
//	@Success		200			{object}	StateResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/states/{stateId} [patch]
func (h *ProjectHandler) UpdateState(c *echo.Context) error {
	stateIDStr := c.Param("stateId")
	stateID, err := uuid.Parse(stateIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid state ID")
	}

	var req UpdateStateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ctx := c.Request().Context()

	state, err := h.store.UpdateProjectState(ctx, store.UpdateProjectStateParams{
		ID:       stateID,
		Name:     stringToPgtypeText(req.Name),
		Color:    stringToPgtypeText(req.Color),
		Position: intToPgtypeInt4(req.Position),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update state")
	}

	return c.JSON(http.StatusOK, StateResponse{
		ID:        state.ID,
		StateType: state.StateType,
		Name:      state.Name,
		Color:     textToString(state.Color, "#6B7280"),
		Position:  int(state.Position),
		IsDefault: state.IsDefault,
		CreatedAt: state.CreatedAt.Time,
	})
}

// DeleteState deletes a state.
//
//	@Summary		Delete state
//	@Description	Delete a workflow state. Cannot delete default states or states with tasks. Requires project admin role.
//	@Tags			Project States
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			stateId		path		string	true	"State ID"
//	@Success		200			{object}	MessageResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/states/{stateId} [delete]
func (h *ProjectHandler) DeleteState(c *echo.Context) error {
	stateIDStr := c.Param("stateId")
	stateID, err := uuid.Parse(stateIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid state ID")
	}

	ctx := c.Request().Context()

	// Check if state is default
	state, err := h.store.GetProjectStateByID(ctx, stateID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "state not found")
	}
	if state.IsDefault {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot delete default state")
	}

	// Check if any tasks use this state
	count, err := h.store.CountTasksInState(ctx, stateID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check state usage")
	}
	if count > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot delete state with tasks")
	}

	err = h.store.DeleteProjectState(ctx, stateID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete state")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "state deleted"})
}

// LabelResponse represents a project label in API responses.
type LabelResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateLabelRequest represents the request to create a label.
type CreateLabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// UpdateLabelRequest represents the request to update a label.
type UpdateLabelRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

// ListLabels returns project labels.
//
//	@Summary		List labels
//	@Description	Returns all labels for a project.
//	@Tags			Project Labels
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Success		200			{array}		LabelResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/labels [get]
func (h *ProjectHandler) ListLabels(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()

	labels, err := h.store.ListProjectLabels(ctx, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list labels")
	}

	labelResponses := make([]LabelResponse, len(labels))
	for i, l := range labels {
		labelResponses[i] = LabelResponse{
			ID:        l.ID,
			Name:      l.Name,
			Color:     textToString(l.Color, "#3B82F6"),
			CreatedAt: l.CreatedAt.Time,
		}
	}

	return c.JSON(http.StatusOK, labelResponses)
}

// CreateLabel creates a new label.
//
//	@Summary		Create label
//	@Description	Create a new project label.
//	@Tags			Project Labels
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string				true	"Project key"
//	@Param			body		body		CreateLabelRequest	true	"Label details"
//	@Success		201			{object}	LabelResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/labels [post]
func (h *ProjectHandler) CreateLabel(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	var req CreateLabelRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	if req.Color == "" {
		req.Color = "#3B82F6"
	}

	ctx := c.Request().Context()

	label, err := h.store.CreateProjectLabel(ctx, store.CreateProjectLabelParams{
		ProjectID: projectID,
		Name:      req.Name,
		Color:     pgtype.Text{String: req.Color, Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create label")
	}

	return c.JSON(http.StatusCreated, LabelResponse{
		ID:        label.ID,
		Name:      label.Name,
		Color:     textToString(label.Color, "#3B82F6"),
		CreatedAt: label.CreatedAt.Time,
	})
}

// UpdateLabel updates a label.
//
//	@Summary		Update label
//	@Description	Update a project label. Requires project admin role.
//	@Tags			Project Labels
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string				true	"Project key"
//	@Param			labelId		path		string				true	"Label ID"
//	@Param			body		body		UpdateLabelRequest	true	"Fields to update"
//	@Success		200			{object}	LabelResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/labels/{labelId} [patch]
func (h *ProjectHandler) UpdateLabel(c *echo.Context) error {
	labelIDStr := c.Param("labelId")
	labelID, err := uuid.Parse(labelIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid label ID")
	}

	var req UpdateLabelRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ctx := c.Request().Context()

	label, err := h.store.UpdateProjectLabel(ctx, store.UpdateProjectLabelParams{
		ID:    labelID,
		Name:  stringToPgtypeText(req.Name),
		Color: stringToPgtypeText(req.Color),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update label")
	}

	return c.JSON(http.StatusOK, LabelResponse{
		ID:        label.ID,
		Name:      label.Name,
		Color:     textToString(label.Color, "#3B82F6"),
		CreatedAt: label.CreatedAt.Time,
	})
}

// DeleteLabel deletes a label.
//
//	@Summary		Delete label
//	@Description	Delete a project label.
//	@Tags			Project Labels
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			labelId		path		string	true	"Label ID"
//	@Success		200			{object}	MessageResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/labels/{labelId} [delete]
func (h *ProjectHandler) DeleteLabel(c *echo.Context) error {
	labelIDStr := c.Param("labelId")
	labelID, err := uuid.Parse(labelIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid label ID")
	}

	ctx := c.Request().Context()

	err = h.store.DeleteProjectLabel(ctx, labelID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete label")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "label deleted"})
}

// TemplateResponse represents a task template in API responses.
type TemplateResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTemplateRequest represents the request to create a template.
type CreateTemplateRequest struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateTemplateRequest represents the request to update a template.
type UpdateTemplateRequest struct {
	Name        *string `json:"name"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

// ListTemplates returns project task templates.
//
//	@Summary		List templates
//	@Description	Returns all task templates for a project.
//	@Tags			Project Templates
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Success		200			{array}		TemplateResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/templates [get]
func (h *ProjectHandler) ListTemplates(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()

	templates, err := h.store.ListTaskTemplates(ctx, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list templates")
	}

	templateResponses := make([]TemplateResponse, len(templates))
	for i, t := range templates {
		templateResponses[i] = TemplateResponse{
			ID:          t.ID,
			Name:        t.Name,
			Title:       t.Title,
			Description: t.Description,
			CreatedBy:   t.CreatedBy,
			CreatedAt:   t.CreatedAt.Time,
			UpdatedAt:   t.UpdatedAt.Time,
		}
	}

	return c.JSON(http.StatusOK, templateResponses)
}

// CreateTemplate creates a new task template.
//
//	@Summary		Create template
//	@Description	Create a new task template. Requires project admin role.
//	@Tags			Project Templates
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string					true	"Project key"
//	@Param			body		body		CreateTemplateRequest	true	"Template details"
//	@Success		201			{object}	TemplateResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/templates [post]
func (h *ProjectHandler) CreateTemplate(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	userIDStr := c.Request().Header.Get(auth.HeaderUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	var req CreateTemplateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	ctx := c.Request().Context()

	template, err := h.store.CreateTaskTemplate(ctx, store.CreateTaskTemplateParams{
		ProjectID:   projectID,
		Name:        req.Name,
		Title:       req.Title,
		Description: req.Description,
		CreatedBy:   userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create template")
	}

	return c.JSON(http.StatusCreated, TemplateResponse{
		ID:          template.ID,
		Name:        template.Name,
		Title:       template.Title,
		Description: template.Description,
		CreatedBy:   template.CreatedBy,
		CreatedAt:   template.CreatedAt.Time,
		UpdatedAt:   template.UpdatedAt.Time,
	})
}

// UpdateTemplate updates a task template.
//
//	@Summary		Update template
//	@Description	Update a task template. Requires project admin role.
//	@Tags			Project Templates
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string					true	"Project key"
//	@Param			templateId	path		string					true	"Template ID"
//	@Param			body		body		UpdateTemplateRequest	true	"Fields to update"
//	@Success		200			{object}	TemplateResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/templates/{templateId} [patch]
func (h *ProjectHandler) UpdateTemplate(c *echo.Context) error {
	templateIDStr := c.Param("templateId")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid template ID")
	}

	var req UpdateTemplateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ctx := c.Request().Context()

	template, err := h.store.UpdateTaskTemplate(ctx, store.UpdateTaskTemplateParams{
		ID:          templateID,
		Name:        stringToPgtypeText(req.Name),
		Title:       stringToPgtypeText(req.Title),
		Description: stringToPgtypeText(req.Description),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update template")
	}

	return c.JSON(http.StatusOK, TemplateResponse{
		ID:          template.ID,
		Name:        template.Name,
		Title:       template.Title,
		Description: template.Description,
		CreatedBy:   template.CreatedBy,
		CreatedAt:   template.CreatedAt.Time,
		UpdatedAt:   template.UpdatedAt.Time,
	})
}

// DeleteTemplate deletes a task template.
//
//	@Summary		Delete template
//	@Description	Delete a task template. Requires project admin role.
//	@Tags			Project Templates
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			templateId	path		string	true	"Template ID"
//	@Success		200			{object}	MessageResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/templates/{templateId} [delete]
func (h *ProjectHandler) DeleteTemplate(c *echo.Context) error {
	templateIDStr := c.Param("templateId")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid template ID")
	}

	ctx := c.Request().Context()

	err = h.store.DeleteTaskTemplate(ctx, templateID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete template")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "template deleted"})
}

// Helper functions for pgtype conversions

func textToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func textToString(t pgtype.Text, def string) string {
	if !t.Valid {
		return def
	}
	return t.String
}

func stringToPgtypeText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func stringToPgtypeUUID(s *string) pgtype.UUID {
	if s == nil {
		return pgtype.UUID{Valid: false}
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgtypeUUIDToURL(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	url := "/api/v1/uploads/" + id.String()
	return &url
}

func intToPgtypeInt4(i *int) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(*i), Valid: true}
}
