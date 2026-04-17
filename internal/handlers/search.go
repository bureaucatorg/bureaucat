package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/auth"
	"bereaucat/internal/store"
)

// SearchHandler handles global search endpoints.
type SearchHandler struct {
	store store.Querier
}

// NewSearchHandler creates a new search handler.
func NewSearchHandler(s store.Querier) *SearchHandler {
	return &SearchHandler{store: s}
}

// Per-category result cap. Kept small so the palette stays responsive and the
// total payload is bounded regardless of query breadth.
const searchResultLimit = 8

// SearchTaskResult represents a task in global search results.
type SearchTaskResult struct {
	ID          uuid.UUID `json:"id"`
	TaskNumber  int       `json:"task_number"`
	TaskKey     string    `json:"task_key"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	ProjectID   uuid.UUID `json:"project_id"`
	ProjectKey  string    `json:"project_key"`
	ProjectName string    `json:"project_name"`
	StateName   string    `json:"state_name"`
	StateType   string    `json:"state_type"`
	StateColor  string    `json:"state_color"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SearchCycleResult represents a cycle in global search results.
type SearchCycleResult struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
	ProjectID   uuid.UUID `json:"project_id"`
	ProjectKey  string    `json:"project_key"`
	ProjectName string    `json:"project_name"`
}

// SearchProjectResult represents a project in global search results.
type SearchProjectResult struct {
	ID          uuid.UUID `json:"id"`
	ProjectKey  string    `json:"project_key"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IconURL     *string   `json:"icon_url,omitempty"`
}

// SearchResponse bundles results across entity types.
type SearchResponse struct {
	Tasks    []SearchTaskResult    `json:"tasks"`
	Cycles   []SearchCycleResult   `json:"cycles"`
	Projects []SearchProjectResult `json:"projects"`
}

// Search performs a global search across tasks, cycles, and projects the
// authenticated user has access to (or all records, if the user is admin).
//
//	@Summary		Global search
//	@Description	Search across tasks, cycles, and projects the user has access to.
//	@Tags			Search
//	@Produce		json
//	@Param			q	query		string	true	"Search query"
//	@Success		200	{object}	SearchResponse
//	@Failure		401	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/search [get]
func (h *SearchHandler) Search(c *echo.Context) error {
	userIDStr := c.Request().Header.Get(auth.HeaderUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	q := strings.TrimSpace(c.QueryParam("q"))
	empty := SearchResponse{Tasks: []SearchTaskResult{}, Cycles: []SearchCycleResult{}, Projects: []SearchProjectResult{}}
	if q == "" {
		return c.JSON(http.StatusOK, empty)
	}

	ctx := c.Request().Context()
	isAdmin := c.Request().Header.Get(auth.HeaderUserType) == "admin"

	resp := empty

	// Tasks
	if isAdmin {
		rows, err := h.store.SearchAllTasks(ctx, store.SearchAllTasksParams{
			Query:      q,
			LimitCount: searchResultLimit,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to search tasks")
		}
		resp.Tasks = make([]SearchTaskResult, len(rows))
		for i, r := range rows {
			resp.Tasks[i] = SearchTaskResult{
				ID:          r.ID,
				TaskNumber:  int(r.TaskNumber),
				TaskKey:     r.ProjectKey + "-" + strconv.Itoa(int(r.TaskNumber)),
				Title:       r.Title,
				Description: textToStringPtr(r.Description),
				ProjectID:   r.ProjectID,
				ProjectKey:  r.ProjectKey,
				ProjectName: r.ProjectName,
				StateName:   r.StateName,
				StateType:   r.StateType,
				StateColor:  textToString(r.StateColor, "#6B7280"),
				UpdatedAt:   r.UpdatedAt.Time,
			}
		}
	} else {
		rows, err := h.store.SearchUserTasks(ctx, store.SearchUserTasksParams{
			UserID:     userID,
			Query:      q,
			LimitCount: searchResultLimit,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to search tasks")
		}
		resp.Tasks = make([]SearchTaskResult, len(rows))
		for i, r := range rows {
			resp.Tasks[i] = SearchTaskResult{
				ID:          r.ID,
				TaskNumber:  int(r.TaskNumber),
				TaskKey:     r.ProjectKey + "-" + strconv.Itoa(int(r.TaskNumber)),
				Title:       r.Title,
				Description: textToStringPtr(r.Description),
				ProjectID:   r.ProjectID,
				ProjectKey:  r.ProjectKey,
				ProjectName: r.ProjectName,
				StateName:   r.StateName,
				StateType:   r.StateType,
				StateColor:  textToString(r.StateColor, "#6B7280"),
				UpdatedAt:   r.UpdatedAt.Time,
			}
		}
	}

	// Cycles
	if isAdmin {
		rows, err := h.store.SearchAllCycles(ctx, store.SearchAllCyclesParams{
			Query:      q,
			LimitCount: searchResultLimit,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to search cycles")
		}
		resp.Cycles = make([]SearchCycleResult, len(rows))
		for i, r := range rows {
			resp.Cycles[i] = SearchCycleResult{
				ID:          r.ID,
				Title:       r.Title,
				StartDate:   r.StartDate.Time.Format(cycleDateLayout),
				EndDate:     r.EndDate.Time.Format(cycleDateLayout),
				ProjectID:   r.ProjectID,
				ProjectKey:  r.ProjectKey,
				ProjectName: r.ProjectName,
			}
		}
	} else {
		rows, err := h.store.SearchUserCycles(ctx, store.SearchUserCyclesParams{
			UserID:     userID,
			Query:      q,
			LimitCount: searchResultLimit,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to search cycles")
		}
		resp.Cycles = make([]SearchCycleResult, len(rows))
		for i, r := range rows {
			resp.Cycles[i] = SearchCycleResult{
				ID:          r.ID,
				Title:       r.Title,
				StartDate:   r.StartDate.Time.Format(cycleDateLayout),
				EndDate:     r.EndDate.Time.Format(cycleDateLayout),
				ProjectID:   r.ProjectID,
				ProjectKey:  r.ProjectKey,
				ProjectName: r.ProjectName,
			}
		}
	}

	// Projects
	if isAdmin {
		rows, err := h.store.SearchAllProjects(ctx, store.SearchAllProjectsParams{
			Query:      q,
			LimitCount: searchResultLimit,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to search projects")
		}
		resp.Projects = make([]SearchProjectResult, len(rows))
		for i, r := range rows {
			resp.Projects[i] = SearchProjectResult{
				ID:          r.ID,
				ProjectKey:  r.ProjectKey,
				Name:        r.Name,
				Description: textToStringPtr(r.Description),
				IconURL:     pgtypeUUIDToURL(r.IconID),
			}
		}
	} else {
		rows, err := h.store.SearchUserProjects(ctx, store.SearchUserProjectsParams{
			UserID:     userID,
			Query:      q,
			LimitCount: searchResultLimit,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to search projects")
		}
		resp.Projects = make([]SearchProjectResult, len(rows))
		for i, r := range rows {
			resp.Projects[i] = SearchProjectResult{
				ID:          r.ID,
				ProjectKey:  r.ProjectKey,
				Name:        r.Name,
				Description: textToStringPtr(r.Description),
				IconURL:     pgtypeUUIDToURL(r.IconID),
			}
		}
	}

	return c.JSON(http.StatusOK, resp)
}
