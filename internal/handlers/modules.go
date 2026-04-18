package handlers

import (
	"context"
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

// ModuleHandler handles module-related endpoints.
type ModuleHandler struct {
	store store.Querier
}

// NewModuleHandler creates a new module handler.
func NewModuleHandler(s store.Querier) *ModuleHandler {
	return &ModuleHandler{store: s}
}

const moduleDateLayout = "2006-01-02"

var validModuleStatuses = map[string]struct{}{
	"backlog":     {},
	"planned":     {},
	"in_progress": {},
	"paused":      {},
	"completed":   {},
	"cancelled":   {},
}

// ModuleUserBrief is the embedded user shape used for lead and members.
type ModuleUserBrief struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

// ModuleResponse is the shared shape for module API responses.
type ModuleResponse struct {
	ID             uuid.UUID         `json:"id"`
	ProjectID      uuid.UUID         `json:"project_id"`
	Title          string            `json:"title"`
	Description    *string           `json:"description,omitempty"`
	Status         string            `json:"status"`
	StartDate      *string           `json:"start_date,omitempty"`
	EndDate        *string           `json:"end_date,omitempty"`
	Lead           *ModuleUserBrief  `json:"lead,omitempty"`
	Members        []ModuleUserBrief `json:"members"`
	CreatedBy      uuid.UUID         `json:"created_by"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	TotalTasks     int               `json:"total_tasks"`
	CompletedTasks int               `json:"completed_tasks"`
	ProjectKey     string            `json:"project_key,omitempty"`
	ProjectName    string            `json:"project_name,omitempty"`
}

// CreateModuleRequest is the request body for creating a module.
type CreateModuleRequest struct {
	Title       string   `json:"title"`
	Description *string  `json:"description"`
	Status      *string  `json:"status"`
	StartDate   *string  `json:"start_date"`
	EndDate     *string  `json:"end_date"`
	LeadID      *string  `json:"lead_id"`
	MemberIDs   []string `json:"member_ids"`
}

// UpdateModuleRequest allows partial updates. Nil = leave unchanged. For the
// nullable fields (dates, lead), an explicit JSON null is represented by the
// Clear* flags below, because a nil pointer already means "don't touch."
type UpdateModuleRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
	LeadID      *string `json:"lead_id"`

	ClearStartDate bool `json:"clear_start_date"`
	ClearEndDate   bool `json:"clear_end_date"`
	ClearLead      bool `json:"clear_lead"`
}

// DuplicateModuleRequest drives the duplicate-module wizard.
type DuplicateModuleRequest struct {
	Title     string   `json:"title"`
	StartDate *string  `json:"start_date"`
	EndDate   *string  `json:"end_date"`
	TaskIDs   []string `json:"task_ids"`
}

// AddModuleTasksRequest adds existing tasks to a module in bulk.
type AddModuleTasksRequest struct {
	TaskIDs []string `json:"task_ids"`
}

// AddModuleMemberRequest adds a single member to a module.
type AddModuleMemberRequest struct {
	UserID string `json:"user_id"`
}

// PaginatedModulesResponse wraps a list of modules.
type PaginatedModulesResponse struct {
	Modules    []ModuleResponse `json:"modules"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PerPage    int              `json:"per_page"`
	TotalPages int              `json:"total_pages"`
}

// ModuleTaskResponse is a task as it appears in a module context.
type ModuleTaskResponse struct {
	ID         uuid.UUID          `json:"id"`
	ProjectKey string             `json:"project_key"`
	TaskNumber int                `json:"task_number"`
	TaskID     string             `json:"task_id"`
	Title      string             `json:"title"`
	StateID    uuid.UUID          `json:"state_id"`
	StateName  string             `json:"state_name"`
	StateType  string             `json:"state_type"`
	StateColor string             `json:"state_color"`
	Priority   int                `json:"priority"`
	StartDate  *time.Time         `json:"start_date,omitempty"`
	DueDate    *time.Time         `json:"due_date,omitempty"`
	Assignees  []AssigneeResponse `json:"assignees"`
}

// ModuleMetricsResponse mirrors CycleMetricsResponse for modules.
type ModuleMetricsResponse struct {
	Total          int                     `json:"total"`
	Completed      int                     `json:"completed"`
	InProgress     int                     `json:"in_progress"`
	Todo           int                     `json:"todo"`
	Cancelled      int                     `json:"cancelled"`
	StateBreakdown []ModuleStateBucket     `json:"state_breakdown"`
}

// ModuleStateBucket is the per-state task count breakdown.
type ModuleStateBucket struct {
	StateID    uuid.UUID `json:"state_id"`
	StateName  string    `json:"state_name"`
	StateColor string    `json:"state_color"`
	StateType  string    `json:"state_type"`
	Count      int       `json:"count"`
}

// ====================== Helpers ======================

func parseModuleDate(s string) (pgtype.Date, error) {
	t, err := time.Parse(moduleDateLayout, s)
	if err != nil {
		return pgtype.Date{}, err
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

func moduleDateToPtr(d pgtype.Date) *string {
	if !d.Valid {
		return nil
	}
	s := d.Time.Format(moduleDateLayout)
	return &s
}

func parseOptionalModuleDate(s *string) (pgtype.Date, bool, error) {
	if s == nil {
		return pgtype.Date{Valid: false}, false, nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return pgtype.Date{Valid: false}, false, nil
	}
	d, err := parseModuleDate(trimmed)
	if err != nil {
		return pgtype.Date{Valid: false}, false, err
	}
	return d, true, nil
}

func pgUUIDFromPtr(s *string) (pgtype.UUID, error) {
	if s == nil {
		return pgtype.UUID{Valid: false}, nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return pgtype.UUID{Valid: false}, nil
	}
	id, err := uuid.Parse(trimmed)
	if err != nil {
		return pgtype.UUID{Valid: false}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

func pgUUIDToUserBrief(
	id pgtype.UUID,
	username, firstName, lastName, email string,
	avatar pgtype.Text,
) *ModuleUserBrief {
	if !id.Valid {
		return nil
	}
	return &ModuleUserBrief{
		UserID:    uuid.UUID(id.Bytes),
		Username:  username,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		AvatarURL: textToStringPtr(avatar),
	}
}

// seedMembersFromTasks looks up the distinct assignees of the given task IDs
// and bulk-adds them as module members. No-op if the list is empty.
func (h *ModuleHandler) seedMembersFromTasks(
	ctx context.Context,
	moduleID uuid.UUID,
	taskIDs []uuid.UUID,
	addedBy uuid.UUID,
) error {
	if len(taskIDs) == 0 {
		return nil
	}
	userIDs, err := h.store.GetTaskAssigneesForSeeding(ctx, taskIDs)
	if err != nil {
		return err
	}
	if len(userIDs) == 0 {
		return nil
	}
	return h.store.AddModuleMembersBulk(ctx, store.AddModuleMembersBulkParams{
		ModuleID: moduleID,
		UserIds:  userIDs,
		AddedBy:  addedBy,
	})
}

// loadModuleMembers pulls the members for a module and shapes them for API output.
func (h *ModuleHandler) loadModuleMembers(
	ctx context.Context,
	moduleID uuid.UUID,
) ([]ModuleUserBrief, error) {
	rows, err := h.store.ListModuleMembers(ctx, moduleID)
	if err != nil {
		return nil, err
	}
	out := make([]ModuleUserBrief, len(rows))
	for i, r := range rows {
		out[i] = ModuleUserBrief{
			UserID:    r.UserID,
			Username:  r.Username,
			Email:     r.Email,
			FirstName: r.FirstName,
			LastName:  r.LastName,
			AvatarURL: textToStringPtr(r.AvatarUrl),
		}
	}
	return out, nil
}

// ensureProjectMember returns a 409 error if the user is not a project member.
func (h *ModuleHandler) ensureProjectMember(
	ctx context.Context,
	projectID, userID uuid.UUID,
) error {
	ok, err := h.store.IsProjectMember(ctx, store.IsProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check project membership")
	}
	if !ok {
		return echo.NewHTTPError(http.StatusConflict, "user must be a project member")
	}
	return nil
}

// ====================== Endpoints ======================

// CreateModule creates a new module on a project.
func (h *ModuleHandler) CreateModule(c *echo.Context) error {
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

	var req CreateModuleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	// Status defaults to "backlog" when the caller doesn't supply one.
	statusStr := "backlog"
	if req.Status != nil {
		s := strings.TrimSpace(*req.Status)
		if s != "" {
			if _, ok := validModuleStatuses[s]; !ok {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid status")
			}
			statusStr = s
		}
	}

	start, _, err := parseOptionalModuleDate(req.StartDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid start_date (expected YYYY-MM-DD)")
	}
	end, _, err := parseOptionalModuleDate(req.EndDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid end_date (expected YYYY-MM-DD)")
	}
	if start.Valid && end.Valid && end.Time.Before(start.Time) {
		return echo.NewHTTPError(http.StatusBadRequest, "end_date must be on or after start_date")
	}

	leadParam, err := pgUUIDFromPtr(req.LeadID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid lead_id")
	}

	ctx := c.Request().Context()

	// Validate lead is a project member.
	if leadParam.Valid {
		if err := h.ensureProjectMember(ctx, projectID, uuid.UUID(leadParam.Bytes)); err != nil {
			return err
		}
	}

	created, err := h.store.CreateModule(ctx, store.CreateModuleParams{
		ProjectID:   projectID,
		Title:       req.Title,
		Description: stringToPgtypeText(req.Description),
		Status:      statusStr,
		StartDate:   start,
		EndDate:     end,
		LeadID:      leadParam,
		CreatedBy:   userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create module")
	}

	// Seed members: initial list + lead (auto-added).
	initialMembers := map[uuid.UUID]struct{}{}
	for _, s := range req.MemberIDs {
		id, err := uuid.Parse(strings.TrimSpace(s))
		if err != nil {
			continue
		}
		if err := h.ensureProjectMember(ctx, projectID, id); err != nil {
			return err
		}
		initialMembers[id] = struct{}{}
	}
	if leadParam.Valid {
		initialMembers[uuid.UUID(leadParam.Bytes)] = struct{}{}
	}
	if len(initialMembers) > 0 {
		ids := make([]uuid.UUID, 0, len(initialMembers))
		for id := range initialMembers {
			ids = append(ids, id)
		}
		if err := h.store.AddModuleMembersBulk(ctx, store.AddModuleMembersBulkParams{
			ModuleID: created.ID,
			UserIds:  ids,
			AddedBy:  userID,
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to add initial members")
		}
	}

	members, err := h.loadModuleMembers(ctx, created.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load members")
	}

	return c.JSON(http.StatusCreated, ModuleResponse{
		ID:          created.ID,
		ProjectID:   created.ProjectID,
		Title:       created.Title,
		Description: textToStringPtr(created.Description),
		Status:      created.Status,
		StartDate:   moduleDateToPtr(created.StartDate),
		EndDate:     moduleDateToPtr(created.EndDate),
		Lead:        nil, // loaded fresh via GetModule on next read
		Members:     members,
		CreatedBy:   created.CreatedBy,
		CreatedAt:   created.CreatedAt.Time,
		UpdatedAt:   created.UpdatedAt.Time,
	})
}

// GetModule returns a single module by id with lead + members + progress counts.
func (h *ModuleHandler) GetModule(c *echo.Context) error {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid module ID")
	}
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()
	m, err := h.store.GetModuleByID(ctx, moduleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "module not found")
	}
	if m.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "module not found")
	}

	members, err := h.loadModuleMembers(ctx, moduleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load members")
	}

	return c.JSON(http.StatusOK, ModuleResponse{
		ID:             m.ID,
		ProjectID:      m.ProjectID,
		Title:          m.Title,
		Description:    textToStringPtr(m.Description),
		Status:         m.Status,
		StartDate:      moduleDateToPtr(m.StartDate),
		EndDate:        moduleDateToPtr(m.EndDate),
		Lead:           pgUUIDToUserBrief(m.LeadID, m.LeadUsername, m.LeadFirstName, m.LeadLastName, m.LeadEmail, m.LeadAvatarUrl),
		Members:        members,
		CreatedBy:      m.CreatedBy,
		CreatedAt:      m.CreatedAt.Time,
		UpdatedAt:      m.UpdatedAt.Time,
		TotalTasks:     int(m.TotalTasks),
		CompletedTasks: int(m.CompletedTasks),
		ProjectKey:     m.ProjectKey,
		ProjectName:    m.ProjectName,
	})
}

// ListModules returns a filtered, sorted, paginated list of a project's modules.
func (h *ModuleHandler) ListModules(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}
	projectKey := c.Request().Header.Get(auth.HeaderProjectKey)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	// Filters. Empty string in the SQL means "no filter".
	statusFilter := ""
	if s := strings.TrimSpace(c.QueryParam("status")); s != "" {
		if _, ok := validModuleStatuses[s]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid status filter")
		}
		statusFilter = s
	}
	leadParam := pgtype.UUID{Valid: false}
	if l := strings.TrimSpace(c.QueryParam("lead_id")); l != "" {
		id, err := uuid.Parse(l)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid lead_id filter")
		}
		leadParam = pgtype.UUID{Bytes: id, Valid: true}
	}
	startAfter := pgtype.Date{Valid: false}
	if s := strings.TrimSpace(c.QueryParam("start_after")); s != "" {
		d, err := parseModuleDate(s)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid start_after filter")
		}
		startAfter = d
	}
	endBefore := pgtype.Date{Valid: false}
	if s := strings.TrimSpace(c.QueryParam("end_before")); s != "" {
		d, err := parseModuleDate(s)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid end_before filter")
		}
		endBefore = d
	}

	// Sort
	sortBy := strings.TrimSpace(c.QueryParam("sort_by"))
	switch sortBy {
	case "created_at", "end_date", "progress":
		// ok
	default:
		sortBy = "created_at"
	}
	sortDir := strings.ToLower(strings.TrimSpace(c.QueryParam("sort_dir")))
	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "desc"
	}

	ctx := c.Request().Context()

	total, err := h.store.CountProjectModules(ctx, store.CountProjectModulesParams{
		ProjectID:  projectID,
		Status:     statusFilter,
		LeadID:     leadParam,
		StartAfter: startAfter,
		EndBefore:  endBefore,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count modules")
	}

	rows, err := h.store.ListProjectModules(ctx, store.ListProjectModulesParams{
		ProjectID:  projectID,
		Limit:      int32(perPage),
		Offset:     int32(offset),
		Status:     statusFilter,
		LeadID:     leadParam,
		StartAfter: startAfter,
		EndBefore:  endBefore,
		SortBy:     sortBy,
		SortDir:    sortDir,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list modules")
	}

	// Hydrate members for the card preview in one shot.
	ids := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	membersByModule := map[uuid.UUID][]ModuleUserBrief{}
	if len(ids) > 0 {
		mm, err := h.store.ListModuleMembersForModules(ctx, ids)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to load members")
		}
		for _, r := range mm {
			membersByModule[r.ModuleID] = append(membersByModule[r.ModuleID], ModuleUserBrief{
				UserID:    r.UserID,
				Username:  r.Username,
				FirstName: r.FirstName,
				LastName:  r.LastName,
				AvatarURL: textToStringPtr(r.AvatarUrl),
			})
		}
	}

	modules := make([]ModuleResponse, len(rows))
	for i, r := range rows {
		modules[i] = ModuleResponse{
			ID:             r.ID,
			ProjectID:      r.ProjectID,
			Title:          r.Title,
			Description:    textToStringPtr(r.Description),
			Status:         r.Status,
			StartDate:      moduleDateToPtr(r.StartDate),
			EndDate:        moduleDateToPtr(r.EndDate),
			Lead:           pgUUIDToUserBrief(r.LeadID, r.LeadUsername, r.LeadFirstName, r.LeadLastName, r.LeadEmail, r.LeadAvatarUrl),
			Members:        membersByModule[r.ID],
			CreatedBy:      r.CreatedBy,
			CreatedAt:      r.CreatedAt.Time,
			UpdatedAt:      r.UpdatedAt.Time,
			TotalTasks:     int(r.TotalTasks),
			CompletedTasks: int(r.CompletedTasks),
			ProjectKey:     projectKey,
		}
		if modules[i].Members == nil {
			modules[i].Members = []ModuleUserBrief{}
		}
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, PaginatedModulesResponse{
		Modules:    modules,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// UpdateModule updates fields on a module.
func (h *ModuleHandler) UpdateModule(c *echo.Context) error {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid module ID")
	}
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

	var req UpdateModuleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ctx := c.Request().Context()
	existing, err := h.store.GetModuleByID(ctx, moduleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "module not found")
	}
	if existing.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "module not found")
	}

	// Title
	var titleParam pgtype.Text
	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "title cannot be empty")
		}
		titleParam = pgtype.Text{String: t, Valid: true}
	}

	// Status — empty string means "don't change" (matches SQL CASE).
	statusStr := ""
	if req.Status != nil {
		s := strings.TrimSpace(*req.Status)
		if _, ok := validModuleStatuses[s]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid status")
		}
		statusStr = s
	}

	// Dates
	startParam, _, err := parseOptionalModuleDate(req.StartDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid start_date")
	}
	endParam, _, err := parseOptionalModuleDate(req.EndDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid end_date")
	}

	// Effective dates after the update — check the CHECK constraint before DB
	// does so we can return a clean error.
	effStart := existing.StartDate
	if req.ClearStartDate {
		effStart = pgtype.Date{Valid: false}
	} else if startParam.Valid {
		effStart = startParam
	}
	effEnd := existing.EndDate
	if req.ClearEndDate {
		effEnd = pgtype.Date{Valid: false}
	} else if endParam.Valid {
		effEnd = endParam
	}
	if effStart.Valid && effEnd.Valid && effEnd.Time.Before(effStart.Time) {
		return echo.NewHTTPError(http.StatusBadRequest, "end_date must be on or after start_date")
	}

	// Lead
	leadParam, err := pgUUIDFromPtr(req.LeadID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid lead_id")
	}
	if leadParam.Valid {
		if err := h.ensureProjectMember(ctx, projectID, uuid.UUID(leadParam.Bytes)); err != nil {
			return err
		}
	}

	updated, err := h.store.UpdateModule(ctx, store.UpdateModuleParams{
		ID:             moduleID,
		Title:          titleParam,
		Description:    stringToPgtypeText(req.Description),
		Status:         statusStr,
		StartDate:      startParam,
		EndDate:        endParam,
		LeadID:         leadParam,
		ClearStartDate: req.ClearStartDate,
		ClearEndDate:   req.ClearEndDate,
		ClearLead:      req.ClearLead,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update module")
	}

	// If a (new) lead was set, auto-add as member.
	if updated.LeadID.Valid {
		if err := h.store.AddModuleMember(ctx, store.AddModuleMemberParams{
			ModuleID: moduleID,
			UserID:   uuid.UUID(updated.LeadID.Bytes),
			AddedBy:  userID,
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to sync lead membership")
		}
	}

	members, err := h.loadModuleMembers(ctx, moduleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load members")
	}

	// Reload with joined lead columns to keep response shape stable.
	m, err := h.store.GetModuleByID(ctx, moduleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to reload module")
	}

	return c.JSON(http.StatusOK, ModuleResponse{
		ID:             m.ID,
		ProjectID:      m.ProjectID,
		Title:          m.Title,
		Description:    textToStringPtr(m.Description),
		Status:         m.Status,
		StartDate:      moduleDateToPtr(m.StartDate),
		EndDate:        moduleDateToPtr(m.EndDate),
		Lead:           pgUUIDToUserBrief(m.LeadID, m.LeadUsername, m.LeadFirstName, m.LeadLastName, m.LeadEmail, m.LeadAvatarUrl),
		Members:        members,
		CreatedBy:      m.CreatedBy,
		CreatedAt:      m.CreatedAt.Time,
		UpdatedAt:      m.UpdatedAt.Time,
		TotalTasks:     int(m.TotalTasks),
		CompletedTasks: int(m.CompletedTasks),
	})
}

// DeleteModule soft-deletes a module. Does not touch the linked tasks.
func (h *ModuleHandler) DeleteModule(c *echo.Context) error {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid module ID")
	}
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()
	existing, err := h.store.GetModuleByID(ctx, moduleID)
	if err != nil || existing.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "module not found")
	}

	if err := h.store.SoftDeleteModule(ctx, moduleID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete module")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "module deleted"})
}

// DuplicateModule creates a copy of an existing module with a chosen subset of
// linked tasks. Status resets to backlog; dates come from the request (may be
// empty); members are copied; lead is copied.
func (h *ModuleHandler) DuplicateModule(c *echo.Context) error {
	srcID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid module ID")
	}
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

	var req DuplicateModuleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	start, _, err := parseOptionalModuleDate(req.StartDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid start_date")
	}
	end, _, err := parseOptionalModuleDate(req.EndDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid end_date")
	}
	if start.Valid && end.Valid && end.Time.Before(start.Time) {
		return echo.NewHTTPError(http.StatusBadRequest, "end_date must be on or after start_date")
	}

	ctx := c.Request().Context()
	src, err := h.store.GetModuleByID(ctx, srcID)
	if err != nil || src.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "module not found")
	}

	// Create the new module. Status forced to backlog per spec.
	created, err := h.store.CreateModule(ctx, store.CreateModuleParams{
		ProjectID:   projectID,
		Title:       req.Title,
		Description: src.Description,
		Status:      "backlog",
		StartDate:   start,
		EndDate:     end,
		LeadID:      src.LeadID,
		CreatedBy:   userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create duplicate")
	}

	// Copy members from source.
	srcMembers, err := h.store.ListModuleMembers(ctx, srcID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to read source members")
	}
	if len(srcMembers) > 0 {
		ids := make([]uuid.UUID, len(srcMembers))
		for i, m := range srcMembers {
			ids[i] = m.UserID
		}
		if err := h.store.AddModuleMembersBulk(ctx, store.AddModuleMembersBulkParams{
			ModuleID: created.ID,
			UserIds:  ids,
			AddedBy:  userID,
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to copy members")
		}
	}

	// Link the chosen subset of tasks (may be empty).
	if len(req.TaskIDs) > 0 {
		taskIDs := make([]uuid.UUID, 0, len(req.TaskIDs))
		for _, s := range req.TaskIDs {
			id, err := uuid.Parse(strings.TrimSpace(s))
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid task id: "+s)
			}
			taskIDs = append(taskIDs, id)
		}
		if err := h.store.AddModuleTasksBulk(ctx, store.AddModuleTasksBulkParams{
			ModuleID: created.ID,
			TaskIds:  taskIDs,
			AddedBy:  userID,
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to link tasks")
		}
		if err := h.seedMembersFromTasks(ctx, created.ID, taskIDs, userID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to seed members from tasks")
		}
	}

	m, err := h.store.GetModuleByID(ctx, created.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to reload module")
	}
	members, err := h.loadModuleMembers(ctx, created.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load members")
	}

	return c.JSON(http.StatusCreated, ModuleResponse{
		ID:             m.ID,
		ProjectID:      m.ProjectID,
		Title:          m.Title,
		Description:    textToStringPtr(m.Description),
		Status:         m.Status,
		StartDate:      moduleDateToPtr(m.StartDate),
		EndDate:        moduleDateToPtr(m.EndDate),
		Lead:           pgUUIDToUserBrief(m.LeadID, m.LeadUsername, m.LeadFirstName, m.LeadLastName, m.LeadEmail, m.LeadAvatarUrl),
		Members:        members,
		CreatedBy:      m.CreatedBy,
		CreatedAt:      m.CreatedAt.Time,
		UpdatedAt:      m.UpdatedAt.Time,
		TotalTasks:     int(m.TotalTasks),
		CompletedTasks: int(m.CompletedTasks),
		ProjectKey:     m.ProjectKey,
		ProjectName:    m.ProjectName,
	})
}

// ====================== Tasks ======================

// ListModuleTasks returns the tasks in a module, optionally filtered by assignee.
func (h *ModuleHandler) ListModuleTasks(c *echo.Context) error {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid module ID")
	}
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()
	existing, err := h.store.GetModuleByID(ctx, moduleID)
	if err != nil || existing.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "module not found")
	}

	assigneeParam := pgtype.UUID{Valid: false}
	if a := c.QueryParam("assignee"); a != "" {
		id, err := uuid.Parse(a)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid assignee ID")
		}
		assigneeParam = pgtype.UUID{Bytes: id, Valid: true}
	}

	rows, err := h.store.ListModuleTasks(ctx, store.ListModuleTasksParams{
		ModuleID:   moduleID,
		AssigneeID: assigneeParam,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list module tasks")
	}

	ids := make([]uuid.UUID, len(rows))
	for i, t := range rows {
		ids[i] = t.ID
	}
	assigneesByTask := map[uuid.UUID][]AssigneeResponse{}
	if len(ids) > 0 {
		assignees, err := h.store.ListAssigneesForTasks(ctx, ids)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to load assignees")
		}
		for _, a := range assignees {
			assigneesByTask[a.TaskID] = append(assigneesByTask[a.TaskID], AssigneeResponse{
				ID:        a.ID,
				UserID:    a.UserID,
				Username:  a.Username,
				Email:     a.Email,
				FirstName: a.FirstName,
				LastName:  a.LastName,
				AvatarURL: textToStringPtr(a.AvatarUrl),
			})
		}
	}

	out := make([]ModuleTaskResponse, len(rows))
	for i, t := range rows {
		out[i] = ModuleTaskResponse{
			ID:         t.ID,
			ProjectKey: t.ProjectKey,
			TaskNumber: int(t.TaskNumber),
			TaskID:     t.ProjectKey + "-" + strconv.Itoa(int(t.TaskNumber)),
			Title:      t.Title,
			StateID:    t.StateID,
			StateName:  t.StateName,
			StateType:  t.StateType,
			StateColor: textToString(t.StateColor, "#6B7280"),
			Priority:   int(t.Priority),
			StartDate:  timestamptzToTimePtr(t.StartDate),
			DueDate:    timestamptzToTimePtr(t.DueDate),
			Assignees:  assigneesByTask[t.ID],
		}
		if out[i].Assignees == nil {
			out[i].Assignees = []AssigneeResponse{}
		}
	}
	return c.JSON(http.StatusOK, out)
}

// AddModuleTasks attaches existing project tasks to a module and auto-seeds the
// task assignees as module members.
func (h *ModuleHandler) AddModuleTasks(c *echo.Context) error {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid module ID")
	}
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

	var req AddModuleTasksRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if len(req.TaskIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "task_ids is required")
	}

	ctx := c.Request().Context()
	existing, err := h.store.GetModuleByID(ctx, moduleID)
	if err != nil || existing.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "module not found")
	}

	taskIDs := make([]uuid.UUID, 0, len(req.TaskIDs))
	for _, s := range req.TaskIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid task id: "+s)
		}
		taskIDs = append(taskIDs, id)
	}

	if err := h.store.AddModuleTasksBulk(ctx, store.AddModuleTasksBulkParams{
		ModuleID: moduleID,
		TaskIds:  taskIDs,
		AddedBy:  userID,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to link tasks")
	}

	if err := h.seedMembersFromTasks(ctx, moduleID, taskIDs, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to seed members from tasks")
	}

	return c.JSON(http.StatusOK, map[string]any{"added": len(taskIDs)})
}

// RemoveModuleTask detaches a single task from a module.
func (h *ModuleHandler) RemoveModuleTask(c *echo.Context) error {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid module ID")
	}
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task ID")
	}
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()
	existing, err := h.store.GetModuleByID(ctx, moduleID)
	if err != nil || existing.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "module not found")
	}

	if err := h.store.RemoveModuleTask(ctx, store.RemoveModuleTaskParams{
		ModuleID: moduleID,
		TaskID:   taskID,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove task from module")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "task removed from module"})
}

// ListProjectTasksNotInModule serves the "Existing" tab of the Add Task picker.
// Excludes only tasks already in the target module (a task can be in many).
func (h *ModuleHandler) ListProjectTasksNotInModule(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	moduleIDStr := strings.TrimSpace(c.QueryParam("module_id"))
	if moduleIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "module_id is required")
	}
	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid module_id")
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 200 {
		limit = 50
	}
	searchParam := pgtype.Text{}
	if s := strings.TrimSpace(c.QueryParam("search")); s != "" {
		searchParam = pgtype.Text{String: s, Valid: true}
	}

	ctx := c.Request().Context()
	rows, err := h.store.ListProjectTasksNotInModule(ctx, store.ListProjectTasksNotInModuleParams{
		ProjectID: projectID,
		ModuleID:  moduleID,
		Limit:     int32(limit),
		Search:    searchParam,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list tasks")
	}

	out := make([]ModuleTaskResponse, len(rows))
	for i, t := range rows {
		out[i] = ModuleTaskResponse{
			ID:         t.ID,
			ProjectKey: t.ProjectKey,
			TaskNumber: int(t.TaskNumber),
			TaskID:     t.ProjectKey + "-" + strconv.Itoa(int(t.TaskNumber)),
			Title:      t.Title,
			StateID:    t.StateID,
			StateName:  t.StateName,
			StateType:  t.StateType,
			StateColor: textToString(t.StateColor, "#6B7280"),
			Priority:   int(t.Priority),
			Assignees:  []AssigneeResponse{},
		}
	}
	return c.JSON(http.StatusOK, out)
}

// ====================== Members ======================

// ListModuleMembers returns a module's members.
func (h *ModuleHandler) ListModuleMembers(c *echo.Context) error {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid module ID")
	}
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()
	existing, err := h.store.GetModuleByID(ctx, moduleID)
	if err != nil || existing.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "module not found")
	}

	members, err := h.loadModuleMembers(ctx, moduleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load members")
	}
	return c.JSON(http.StatusOK, members)
}

// AddModuleMember adds a single user as a module member. User must be a project member.
func (h *ModuleHandler) AddModuleMember(c *echo.Context) error {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid module ID")
	}
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}
	addedByStr := c.Request().Header.Get(auth.HeaderUserID)
	addedBy, err := uuid.Parse(addedByStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	var req AddModuleMemberRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	userID, err := uuid.Parse(strings.TrimSpace(req.UserID))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}

	ctx := c.Request().Context()
	existing, err := h.store.GetModuleByID(ctx, moduleID)
	if err != nil || existing.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "module not found")
	}

	if err := h.ensureProjectMember(ctx, projectID, userID); err != nil {
		return err
	}

	if err := h.store.AddModuleMember(ctx, store.AddModuleMemberParams{
		ModuleID: moduleID,
		UserID:   userID,
		AddedBy:  addedBy,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add member")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "member added"})
}

// RemoveModuleMember removes a member from a module.
func (h *ModuleHandler) RemoveModuleMember(c *echo.Context) error {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid module ID")
	}
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()
	existing, err := h.store.GetModuleByID(ctx, moduleID)
	if err != nil || existing.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "module not found")
	}

	if err := h.store.RemoveModuleMember(ctx, store.RemoveModuleMemberParams{
		ModuleID: moduleID,
		UserID:   userID,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove member")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "member removed"})
}

// ====================== Workspace-level ======================

// ListActiveModules returns all in_progress modules across projects the caller
// is a member of. Drives the "/modules/active" sidebar dashboard.
func (h *ModuleHandler) ListActiveModules(c *echo.Context) error {
	userIDStr := c.Request().Header.Get(auth.HeaderUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	ctx := c.Request().Context()
	rows, err := h.store.ListActiveModulesForUser(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list active modules")
	}

	out := make([]ModuleResponse, len(rows))
	for i, r := range rows {
		out[i] = ModuleResponse{
			ID:             r.ID,
			ProjectID:      r.ProjectID,
			Title:          r.Title,
			Description:    textToStringPtr(r.Description),
			Status:         r.Status,
			StartDate:      moduleDateToPtr(r.StartDate),
			EndDate:        moduleDateToPtr(r.EndDate),
			Lead:           pgUUIDToUserBrief(r.LeadID, r.LeadUsername, r.LeadFirstName, r.LeadLastName, r.LeadEmail, r.LeadAvatarUrl),
			Members:        []ModuleUserBrief{},
			CreatedBy:      r.CreatedBy,
			CreatedAt:      r.CreatedAt.Time,
			UpdatedAt:      r.UpdatedAt.Time,
			TotalTasks:     int(r.TotalTasks),
			CompletedTasks: int(r.CompletedTasks),
			ProjectKey:     r.ProjectKey,
			ProjectName:    r.ProjectName,
		}
	}
	return c.JSON(http.StatusOK, out)
}

// ====================== Metrics ======================

// GetModuleMetrics returns progress counts + per-state breakdown.
func (h *ModuleHandler) GetModuleMetrics(c *echo.Context) error {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid module ID")
	}
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()
	existing, err := h.store.GetModuleByID(ctx, moduleID)
	if err != nil || existing.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "module not found")
	}

	m, err := h.store.GetModuleMetrics(ctx, moduleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load metrics")
	}
	breakdown, err := h.store.GetModuleStateBreakdown(ctx, moduleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load state breakdown")
	}
	buckets := make([]ModuleStateBucket, len(breakdown))
	for i, b := range breakdown {
		buckets[i] = ModuleStateBucket{
			StateID:    b.StateID,
			StateName:  b.StateName,
			StateColor: textToString(b.StateColor, "#6B7280"),
			StateType:  b.StateType,
			Count:      int(b.TaskCount),
		}
	}
	return c.JSON(http.StatusOK, ModuleMetricsResponse{
		Total:          int(m.Total),
		Completed:      int(m.Completed),
		InProgress:     int(m.InProgress),
		Todo:           int(m.Todo),
		Cancelled:      int(m.Cancelled),
		StateBreakdown: buckets,
	})
}
