package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/store"
)

const newUserDefaultsKey = "new_user_defaults"

// NewUserDefaultProject is a project every newly created user joins.
type NewUserDefaultProject struct {
	ProjectID     uuid.UUID `json:"project_id"`
	Role          string    `json:"role"`
	ProjectKey    string    `json:"project_key,omitempty"`
	ProjectName   string    `json:"project_name,omitempty"`
	WorkspaceName string    `json:"workspace_name,omitempty"`
}

// NewUserDefaults lists the projects (and so workspaces) new users are added to.
type NewUserDefaults struct {
	Projects []NewUserDefaultProject `json:"projects"`
}

func loadNewUserDefaults(ctx context.Context, q store.Querier) NewUserDefaults {
	defaults := NewUserDefaults{Projects: []NewUserDefaultProject{}}
	setting, err := q.GetSetting(ctx, newUserDefaultsKey)
	if err != nil {
		return defaults
	}
	_ = json.Unmarshal(setting.Value, &defaults)
	return defaults
}

// applyNewUserDefaults adds a just-created user to the configured default projects.
func applyNewUserDefaults(ctx context.Context, q store.Querier, userID uuid.UUID) {
	defaults := loadNewUserDefaults(ctx, q)
	if len(defaults.Projects) == 0 {
		return
	}
	byRole := map[string][]uuid.UUID{}
	for _, p := range defaults.Projects {
		byRole[p.Role] = append(byRole[p.Role], p.ProjectID)
	}
	for role, ids := range byRole {
		_ = q.AddUserToDefaultProjects(ctx, store.AddUserToDefaultProjectsParams{
			ProjectIds: ids,
			UserID:     userID,
			Role:       role,
		})
	}
}

// withProjectDetails drops deleted projects and fills in key, name and workspace.
func withProjectDetails(ctx context.Context, q store.Querier, defaults NewUserDefaults) (NewUserDefaults, error) {
	ids := make([]uuid.UUID, len(defaults.Projects))
	for i, p := range defaults.Projects {
		ids[i] = p.ProjectID
	}
	rows, err := q.ListProjectsWithWorkspaceByIDs(ctx, ids)
	if err != nil {
		return defaults, err
	}
	byID := make(map[uuid.UUID]store.ListProjectsWithWorkspaceByIDsRow, len(rows))
	for _, r := range rows {
		byID[r.ID] = r
	}
	out := NewUserDefaults{Projects: []NewUserDefaultProject{}}
	for _, p := range defaults.Projects {
		r, ok := byID[p.ProjectID]
		if !ok {
			continue
		}
		p.ProjectKey, p.ProjectName, p.WorkspaceName = r.ProjectKey, r.Name, r.WorkspaceName
		out.Projects = append(out.Projects, p)
	}
	return out, nil
}

// GetNewUserDefaults returns the default projects for new users (admin only).
//
//	@Summary		Get new user defaults
//	@Tags			Admin - Settings
//	@Produce		json
//	@Success		200	{object}	NewUserDefaults
//	@Security		BearerAuth
//	@Router			/admin/settings/new_user_defaults [get]
func (h *SettingsHandler) GetNewUserDefaults(c *echo.Context) error {
	ctx := c.Request().Context()
	defaults, err := withProjectDetails(ctx, h.store, loadNewUserDefaults(ctx, h.store))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load new user defaults")
	}
	return c.JSON(http.StatusOK, defaults)
}

// UpdateNewUserDefaults replaces the default projects for new users (admin only).
//
//	@Summary		Update new user defaults
//	@Description	Projects (with role guest or member) that newly created users join, along with their workspaces. Existing users are not changed.
//	@Tags			Admin - Settings
//	@Accept			json
//	@Produce		json
//	@Param			body	body		NewUserDefaults	true	"Default projects"
//	@Success		200		{object}	NewUserDefaults
//	@Failure		400		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/settings/new_user_defaults [put]
func (h *SettingsHandler) UpdateNewUserDefaults(c *echo.Context) error {
	var req NewUserDefaults
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	clean := NewUserDefaults{Projects: []NewUserDefaultProject{}}
	seen := map[uuid.UUID]bool{}
	for _, p := range req.Projects {
		if p.Role != "guest" && p.Role != "member" {
			return echo.NewHTTPError(http.StatusBadRequest, "role must be guest or member")
		}
		if seen[p.ProjectID] {
			continue
		}
		seen[p.ProjectID] = true
		clean.Projects = append(clean.Projects, NewUserDefaultProject{ProjectID: p.ProjectID, Role: p.Role})
	}

	ctx := c.Request().Context()
	detailed, err := withProjectDetails(ctx, h.store, clean)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to look up projects")
	}
	if len(detailed.Projects) != len(clean.Projects) {
		return echo.NewHTTPError(http.StatusBadRequest, "one or more projects do not exist")
	}

	value, err := json.Marshal(clean)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to marshal settings")
	}
	if _, err := h.store.UpsertSetting(ctx, store.UpsertSettingParams{Key: newUserDefaultsKey, Value: value}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update new user defaults")
	}
	return c.JSON(http.StatusOK, detailed)
}
