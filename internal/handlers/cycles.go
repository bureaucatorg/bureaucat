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

// CycleHandler handles cycle-related endpoints.
type CycleHandler struct {
	store store.Querier
}

// NewCycleHandler creates a new cycle handler.
func NewCycleHandler(s store.Querier) *CycleHandler {
	return &CycleHandler{store: s}
}

const cycleDateLayout = "2006-01-02"

// CycleResponse is the shared shape for cycle API responses.
type CycleResponse struct {
	ID             uuid.UUID `json:"id"`
	ProjectID      uuid.UUID `json:"project_id"`
	Title          string    `json:"title"`
	Description    *string   `json:"description,omitempty"`
	StartDate      string    `json:"start_date"`
	EndDate        string    `json:"end_date"`
	Status         string    `json:"status"` // upcoming | active | completed
	CreatedBy      uuid.UUID `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	TotalTasks     int       `json:"total_tasks"`
	CompletedTasks int       `json:"completed_tasks"`
	ProjectKey     string    `json:"project_key,omitempty"`
	ProjectName    string    `json:"project_name,omitempty"`
}

// CreateCycleRequest is the request body for creating a cycle.
type CreateCycleRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	StartDate   string  `json:"start_date"` // YYYY-MM-DD
	EndDate     string  `json:"end_date"`   // YYYY-MM-DD
}

// UpdateCycleRequest is the request body for updating a cycle.
type UpdateCycleRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
}

// PaginatedCyclesResponse wraps a list of cycles.
type PaginatedCyclesResponse struct {
	Cycles     []CycleResponse `json:"cycles"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	TotalPages int             `json:"total_pages"`
}

// CycleTaskResponse is a task as it appears in a cycle context.
type CycleTaskResponse struct {
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

// CycleMetricsResponse combines top-level counts, state breakdown, and assignees.
type CycleMetricsResponse struct {
	Total          int                    `json:"total"`
	Completed      int                    `json:"completed"`
	InProgress     int                    `json:"in_progress"`
	Todo           int                    `json:"todo"`
	Cancelled      int                    `json:"cancelled"`
	StateBreakdown []CycleStateBucket     `json:"state_breakdown"`
	Assignees      []CycleAssigneeSummary `json:"assignees"`
}

// CycleStateBucket is the per-state task count breakdown.
type CycleStateBucket struct {
	StateID    uuid.UUID `json:"state_id"`
	StateName  string    `json:"state_name"`
	StateColor string    `json:"state_color"`
	StateType  string    `json:"state_type"`
	Count      int       `json:"count"`
}

// CycleAssigneeSummary is an assignee with their task count in a cycle.
type CycleAssigneeSummary struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	TaskCount int       `json:"task_count"`
}

// AddCycleTasksRequest adds existing tasks to a cycle in bulk.
type AddCycleTasksRequest struct {
	TaskIDs []string `json:"task_ids"`
}

// ====================== Helpers ======================

func parseCycleDate(s string) (pgtype.Date, error) {
	t, err := time.Parse(cycleDateLayout, s)
	if err != nil {
		return pgtype.Date{}, err
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

func dateToString(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format(cycleDateLayout)
}

// derivedCycleStatus returns "upcoming", "active", or "completed" based on the
// current UTC date and the cycle's date range.
func derivedCycleStatus(start, end pgtype.Date) string {
	if !start.Valid || !end.Valid {
		return "upcoming"
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	s := start.Time.Truncate(24 * time.Hour)
	e := end.Time.Truncate(24 * time.Hour)
	switch {
	case today.Before(s):
		return "upcoming"
	case today.After(e):
		return "completed"
	default:
		return "active"
	}
}

// ====================== Endpoints ======================

// CreateCycle creates a new cycle on a project.
func (h *CycleHandler) CreateCycle(c *echo.Context) error {
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

	var req CreateCycleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	startDate, err := parseCycleDate(req.StartDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid start_date (expected YYYY-MM-DD)")
	}
	endDate, err := parseCycleDate(req.EndDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid end_date (expected YYYY-MM-DD)")
	}
	if endDate.Time.Before(startDate.Time) {
		return echo.NewHTTPError(http.StatusBadRequest, "end_date must be on or after start_date")
	}

	ctx := c.Request().Context()

	overlap, err := h.store.CheckCycleOverlap(ctx, store.CheckCycleOverlapParams{
		ProjectID: projectID,
		StartDate: startDate,
		EndDate:   endDate,
		ExcludeID: pgtype.UUID{Valid: false},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check cycle overlap")
	}
	if overlap > 0 {
		return echo.NewHTTPError(http.StatusConflict, "another cycle exists in this date range")
	}

	cycle, err := h.store.CreateCycle(ctx, store.CreateCycleParams{
		ProjectID:   projectID,
		Title:       req.Title,
		Description: stringToPgtypeText(req.Description),
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedBy:   userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create cycle")
	}

	return c.JSON(http.StatusCreated, CycleResponse{
		ID:          cycle.ID,
		ProjectID:   cycle.ProjectID,
		Title:       cycle.Title,
		Description: textToStringPtr(cycle.Description),
		StartDate:   dateToString(cycle.StartDate),
		EndDate:     dateToString(cycle.EndDate),
		Status:      derivedCycleStatus(cycle.StartDate, cycle.EndDate),
		CreatedBy:   cycle.CreatedBy,
		CreatedAt:   cycle.CreatedAt.Time,
		UpdatedAt:   cycle.UpdatedAt.Time,
	})
}

// ListCycles returns a paginated list of a project's cycles.
func (h *CycleHandler) ListCycles(c *echo.Context) error {
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

	ctx := c.Request().Context()

	total, err := h.store.CountProjectCycles(ctx, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count cycles")
	}

	rows, err := h.store.ListProjectCycles(ctx, store.ListProjectCyclesParams{
		ProjectID: projectID,
		Limit:     int32(perPage),
		Offset:    int32(offset),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list cycles")
	}

	cycles := make([]CycleResponse, len(rows))
	for i, r := range rows {
		cycles[i] = CycleResponse{
			ID:             r.ID,
			ProjectID:      r.ProjectID,
			Title:          r.Title,
			Description:    textToStringPtr(r.Description),
			StartDate:      dateToString(r.StartDate),
			EndDate:        dateToString(r.EndDate),
			Status:         derivedCycleStatus(r.StartDate, r.EndDate),
			CreatedBy:      r.CreatedBy,
			CreatedAt:      r.CreatedAt.Time,
			UpdatedAt:      r.UpdatedAt.Time,
			TotalTasks:     int(r.TotalTasks),
			CompletedTasks: int(r.CompletedTasks),
			ProjectKey:     projectKey,
		}
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, PaginatedCyclesResponse{
		Cycles:     cycles,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// GetCycle returns a single cycle by id.
func (h *CycleHandler) GetCycle(c *echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("cycleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cycle ID")
	}
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()
	cycle, err := h.store.GetCycleByID(ctx, cycleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "cycle not found")
	}
	if cycle.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "cycle not found")
	}

	return c.JSON(http.StatusOK, CycleResponse{
		ID:          cycle.ID,
		ProjectID:   cycle.ProjectID,
		Title:       cycle.Title,
		Description: textToStringPtr(cycle.Description),
		StartDate:   dateToString(cycle.StartDate),
		EndDate:     dateToString(cycle.EndDate),
		Status:      derivedCycleStatus(cycle.StartDate, cycle.EndDate),
		CreatedBy:   cycle.CreatedBy,
		CreatedAt:   cycle.CreatedAt.Time,
		UpdatedAt:   cycle.UpdatedAt.Time,
		ProjectKey:  cycle.ProjectKey,
		ProjectName: cycle.ProjectName,
	})
}

// UpdateCycle updates fields on a cycle.
func (h *CycleHandler) UpdateCycle(c *echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("cycleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cycle ID")
	}
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	var req UpdateCycleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ctx := c.Request().Context()
	existing, err := h.store.GetCycleByID(ctx, cycleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "cycle not found")
	}
	if existing.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "cycle not found")
	}

	var titleParam pgtype.Text
	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "title cannot be empty")
		}
		titleParam = pgtype.Text{String: t, Valid: true}
	}

	descParam := stringToPgtypeText(req.Description)

	startParam := pgtype.Date{Valid: false}
	if req.StartDate != nil {
		d, err := parseCycleDate(*req.StartDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid start_date")
		}
		startParam = d
	}
	endParam := pgtype.Date{Valid: false}
	if req.EndDate != nil {
		d, err := parseCycleDate(*req.EndDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid end_date")
		}
		endParam = d
	}

	// Compute the effective date range after the update to check overlap
	effStart := existing.StartDate
	if startParam.Valid {
		effStart = startParam
	}
	effEnd := existing.EndDate
	if endParam.Valid {
		effEnd = endParam
	}
	if effEnd.Time.Before(effStart.Time) {
		return echo.NewHTTPError(http.StatusBadRequest, "end_date must be on or after start_date")
	}

	if startParam.Valid || endParam.Valid {
		overlap, err := h.store.CheckCycleOverlap(ctx, store.CheckCycleOverlapParams{
			ProjectID: projectID,
			StartDate: effStart,
			EndDate:   effEnd,
			ExcludeID: pgtype.UUID{Bytes: cycleID, Valid: true},
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to check cycle overlap")
		}
		if overlap > 0 {
			return echo.NewHTTPError(http.StatusConflict, "another cycle exists in this date range")
		}
	}

	updated, err := h.store.UpdateCycle(ctx, store.UpdateCycleParams{
		ID:          cycleID,
		Title:       titleParam,
		Description: descParam,
		StartDate:   startParam,
		EndDate:     endParam,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update cycle")
	}

	return c.JSON(http.StatusOK, CycleResponse{
		ID:          updated.ID,
		ProjectID:   updated.ProjectID,
		Title:       updated.Title,
		Description: textToStringPtr(updated.Description),
		StartDate:   dateToString(updated.StartDate),
		EndDate:     dateToString(updated.EndDate),
		Status:      derivedCycleStatus(updated.StartDate, updated.EndDate),
		CreatedBy:   updated.CreatedBy,
		CreatedAt:   updated.CreatedAt.Time,
		UpdatedAt:   updated.UpdatedAt.Time,
	})
}

// DeleteCycle soft-deletes a cycle.
func (h *CycleHandler) DeleteCycle(c *echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("cycleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cycle ID")
	}
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()
	existing, err := h.store.GetCycleByID(ctx, cycleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "cycle not found")
	}
	if existing.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "cycle not found")
	}

	if err := h.store.SoftDeleteCycle(ctx, cycleID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete cycle")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "cycle deleted"})
}

// ListCycleTasks returns the tasks in a cycle, optionally filtered by assignee.
func (h *CycleHandler) ListCycleTasks(c *echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("cycleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cycle ID")
	}
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()
	existing, err := h.store.GetCycleByID(ctx, cycleID)
	if err != nil || existing.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "cycle not found")
	}

	assigneeParam := pgtype.UUID{Valid: false}
	if a := c.QueryParam("assignee"); a != "" {
		id, err := uuid.Parse(a)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid assignee ID")
		}
		assigneeParam = pgtype.UUID{Bytes: id, Valid: true}
	}

	rows, err := h.store.ListCycleTasks(ctx, store.ListCycleTasksParams{
		CycleID:    cycleID,
		AssigneeID: assigneeParam,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list cycle tasks")
	}

	// Decorate with assignees
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

	out := make([]CycleTaskResponse, len(rows))
	for i, t := range rows {
		out[i] = CycleTaskResponse{
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

// GetCycleMetrics returns progress, state breakdown, and assignee summary.
func (h *CycleHandler) GetCycleMetrics(c *echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("cycleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cycle ID")
	}
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	ctx := c.Request().Context()
	existing, err := h.store.GetCycleByID(ctx, cycleID)
	if err != nil || existing.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "cycle not found")
	}

	m, err := h.store.GetCycleMetrics(ctx, cycleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load metrics")
	}

	breakdown, err := h.store.GetCycleStateBreakdown(ctx, cycleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load state breakdown")
	}
	buckets := make([]CycleStateBucket, len(breakdown))
	for i, b := range breakdown {
		buckets[i] = CycleStateBucket{
			StateID:    b.StateID,
			StateName:  b.StateName,
			StateColor: textToString(b.StateColor, "#6B7280"),
			StateType:  b.StateType,
			Count:      int(b.TaskCount),
		}
	}

	assignees, err := h.store.ListCycleAssignees(ctx, cycleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load assignees")
	}
	summaries := make([]CycleAssigneeSummary, len(assignees))
	for i, a := range assignees {
		summaries[i] = CycleAssigneeSummary{
			UserID:    a.UserID,
			Username:  a.Username,
			FirstName: a.FirstName,
			LastName:  a.LastName,
			AvatarURL: textToStringPtr(a.AvatarUrl),
			TaskCount: int(a.TaskCount),
		}
	}

	return c.JSON(http.StatusOK, CycleMetricsResponse{
		Total:          int(m.Total),
		Completed:      int(m.Completed),
		InProgress:     int(m.InProgress),
		Todo:           int(m.Todo),
		Cancelled:      int(m.Cancelled),
		StateBreakdown: buckets,
		Assignees:      summaries,
	})
}

// AddCycleTasks attaches existing project tasks to a cycle.
func (h *CycleHandler) AddCycleTasks(c *echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("cycleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cycle ID")
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

	var req AddCycleTasksRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if len(req.TaskIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "task_ids is required")
	}

	ctx := c.Request().Context()
	existing, err := h.store.GetCycleByID(ctx, cycleID)
	if err != nil || existing.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "cycle not found")
	}

	ids := make([]uuid.UUID, 0, len(req.TaskIDs))
	for _, s := range req.TaskIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid task id: "+s)
		}
		ids = append(ids, id)
	}

	if err := h.store.AddTasksToCycle(ctx, store.AddTasksToCycleParams{
		CycleID: cycleID,
		TaskIds: ids,
		AddedBy: userID,
	}); err != nil {
		return echo.NewHTTPError(http.StatusConflict, "one or more tasks are already in a cycle")
	}

	return c.JSON(http.StatusOK, map[string]any{"added": len(ids)})
}

// RemoveCycleTask detaches a single task from a cycle.
func (h *CycleHandler) RemoveCycleTask(c *echo.Context) error {
	cycleID, err := uuid.Parse(c.Param("cycleId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cycle ID")
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
	existing, err := h.store.GetCycleByID(ctx, cycleID)
	if err != nil || existing.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusNotFound, "cycle not found")
	}

	if err := h.store.RemoveTaskFromCycle(ctx, store.RemoveTaskFromCycleParams{
		CycleID: cycleID,
		TaskID:  taskID,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove task from cycle")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "task removed from cycle"})
}

// ListUnassignedTasks returns project tasks not currently in any cycle. For the
// "Existing" tab of the Add Task picker.
func (h *CycleHandler) ListUnassignedTasks(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
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
	rows, err := h.store.ListUnassignedProjectTasks(ctx, store.ListUnassignedProjectTasksParams{
		ProjectID: projectID,
		Limit:     int32(limit),
		Search:    searchParam,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list tasks")
	}

	out := make([]CycleTaskResponse, len(rows))
	for i, t := range rows {
		out[i] = CycleTaskResponse{
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

// ActiveCycleResponse is the shape for the workspace active cycles dashboard.
type ActiveCycleResponse struct {
	CycleResponse
}

// ListActiveCycles returns cycles currently active (today in [start, end])
// across all projects the user is a member of.
func (h *CycleHandler) ListActiveCycles(c *echo.Context) error {
	userIDStr := c.Request().Header.Get(auth.HeaderUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	ctx := c.Request().Context()
	rows, err := h.store.ListActiveCyclesForUser(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list active cycles")
	}

	out := make([]CycleResponse, len(rows))
	for i, r := range rows {
		out[i] = CycleResponse{
			ID:             r.ID,
			ProjectID:      r.ProjectID,
			Title:          r.Title,
			Description:    textToStringPtr(r.Description),
			StartDate:      dateToString(r.StartDate),
			EndDate:        dateToString(r.EndDate),
			Status:         "active",
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

// ListProjectCyclesAll returns every cycle on a project (for the switcher dropdown).
// No pagination. Filter is applied by project membership via ProjectMiddleware.
func (h *CycleHandler) ListAllProjectCycles(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}
	ctx := c.Request().Context()
	rows, err := h.store.ListProjectCyclesAll(ctx, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list cycles")
	}
	type sibling struct {
		ID        uuid.UUID `json:"id"`
		Title     string    `json:"title"`
		StartDate string    `json:"start_date"`
		EndDate   string    `json:"end_date"`
		Status    string    `json:"status"`
	}
	out := make([]sibling, len(rows))
	for i, r := range rows {
		out[i] = sibling{
			ID:        r.ID,
			Title:     r.Title,
			StartDate: dateToString(r.StartDate),
			EndDate:   dateToString(r.EndDate),
			Status:    derivedCycleStatus(r.StartDate, r.EndDate),
		}
	}
	return c.JSON(http.StatusOK, out)
}
