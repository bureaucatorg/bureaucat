package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/auth"
	"bereaucat/internal/store"
)

// ProjectStateStat is the top-level task count for one project state.
type ProjectStateStat struct {
	Name      string  `json:"name"`
	StateType string  `json:"state_type"`
	Color     *string `json:"color,omitempty"`
	Count     int     `json:"count"`
}

// WorkloadStat is the number of open tasks and subtasks assigned to one member.
type WorkloadStat struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	OpenTasks int       `json:"open_tasks"`
}

// ProjectStatsDay holds every per-day series for one day.
type ProjectStatsDay struct {
	Day         string `json:"day"`
	Tasks       int    `json:"tasks"`
	Subtasks    int    `json:"subtasks"`
	Completed   int    `json:"completed"`
	Activity    int    `json:"activity"`
	Comments    int    `json:"comments"`
	Attachments int    `json:"attachments"`
	Pages       int    `json:"pages"`
	Cycles      int    `json:"cycles"`
	Modules     int    `json:"modules"`
}

// ProjectStatsResponse is the payload for the project stats tab.
type ProjectStatsResponse struct {
	Totals struct {
		Tasks            int64 `json:"tasks"`
		Subtasks         int64 `json:"subtasks"`
		BlockedTasks     int64 `json:"blocked_tasks"`
		BlockingTasks    int64 `json:"blocking_tasks"`
		UnassignedOpen   int64 `json:"unassigned_open"`
		Pages            int64 `json:"pages"`
		Members          int64 `json:"members"`
		Cycles           int64 `json:"cycles"`
		Modules          int64 `json:"modules"`
		Attachments      int64 `json:"attachments"`
		AttachmentsBytes int64 `json:"attachments_bytes"`
	} `json:"totals"`
	TasksByState    []ProjectStateStat `json:"tasks_by_state"`
	TasksByPriority []StatCount        `json:"tasks_by_priority"`
	Workload        []WorkloadStat     `json:"workload"`
	Series          struct {
		From string            `json:"from"`
		To   string            `json:"to"`
		Days []ProjectStatsDay `json:"days"`
	} `json:"series"`
}

// GetProjectStats returns totals, breakdowns, workload and per-day trends for one project.
//
//	@Summary		Get project stats
//	@Description	Returns project totals, task breakdowns, open-task workload per member and per-day series for the date range.
//	@Tags			Projects
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			from		query		string	false	"Start date (YYYY-MM-DD), defaults to 29 days before to"
//	@Param			to			query		string	false	"End date (YYYY-MM-DD), defaults to today"
//	@Success		200			{object}	ProjectStatsResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/stats [get]
func (h *ProjectHandler) GetProjectStats(c *echo.Context) error {
	projectID, err := uuid.Parse(c.Request().Header.Get(auth.HeaderProjectID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}
	ctx := c.Request().Context()
	fromDate, toDate := statsRange(c)

	var resp ProjectStatsResponse

	totals, err := h.store.ProjectStatsTotals(ctx, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load project totals")
	}
	t := &resp.Totals
	t.Tasks, t.Subtasks = totals.Tasks, totals.Subtasks
	t.BlockedTasks, t.BlockingTasks, t.UnassignedOpen = totals.BlockedTasks, totals.BlockingTasks, totals.UnassignedOpen
	t.Pages, t.Members, t.Cycles, t.Modules = totals.Pages, totals.Members, totals.Cycles, totals.Modules
	t.Attachments, t.AttachmentsBytes = totals.Attachments, totals.AttachmentsBytes

	states, err := h.store.ProjectTasksByState(ctx, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load task states")
	}
	resp.TasksByState = make([]ProjectStateStat, len(states))
	for i, s := range states {
		resp.TasksByState[i] = ProjectStateStat{Name: s.Name, StateType: s.StateType, Color: textToStringPtr(s.Color), Count: int(s.Count)}
	}

	priorities, err := h.store.ProjectTasksByPriority(ctx, projectID)
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

	workload, err := h.store.ProjectWorkload(ctx, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load workload")
	}
	resp.Workload = make([]WorkloadStat, len(workload))
	for i, w := range workload {
		resp.Workload[i] = WorkloadStat{
			UserID:    w.UserID,
			Username:  w.Username,
			FirstName: w.FirstName,
			LastName:  w.LastName,
			AvatarURL: textToStringPtr(w.AvatarUrl),
			OpenTasks: int(w.OpenTasks),
		}
	}

	series, err := h.store.ProjectStatsSeries(ctx, store.ProjectStatsSeriesParams{
		ProjectID: projectID,
		FromDate:  pgtype.Date{Time: fromDate, Valid: true},
		ToDate:    pgtype.Date{Time: toDate, Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load project series")
	}
	resp.Series.From = fromDate.Format(statsDateLayout)
	resp.Series.To = toDate.Format(statsDateLayout)
	resp.Series.Days = make([]ProjectStatsDay, len(series))
	for i, d := range series {
		resp.Series.Days[i] = ProjectStatsDay{
			Day:         d.Day.Time.Format(statsDateLayout),
			Tasks:       int(d.Tasks),
			Subtasks:    int(d.Subtasks),
			Completed:   int(d.Completed),
			Activity:    int(d.Activity),
			Comments:    int(d.Comments),
			Attachments: int(d.Attachments),
			Pages:       int(d.Pages),
			Cycles:      int(d.Cycles),
			Modules:     int(d.Modules),
		}
	}

	return c.JSON(http.StatusOK, resp)
}
