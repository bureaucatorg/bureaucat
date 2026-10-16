package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/auth"
	"bereaucat/internal/store"
	"bereaucat/internal/workflow"
)

// WorkflowStateConfig is a state's node in the workflow graph.
type WorkflowStateConfig struct {
	StateID         uuid.UUID   `json:"state_id"`
	IsEntry         bool        `json:"is_entry"`
	PosX            float64     `json:"pos_x"`
	PosY            float64     `json:"pos_y"`
	AddAssignees    []uuid.UUID `json:"add_assignees"`
	RemoveAssignees []uuid.UUID `json:"remove_assignees"`
}

// WorkflowTransitionConfig is an allowed transition and who may make it.
type WorkflowTransitionConfig struct {
	FromStateID    uuid.UUID   `json:"from_state_id"`
	ToStateID      uuid.UUID   `json:"to_state_id"`
	AllowAssignees bool        `json:"allow_assignees"`
	AllowCreator   bool        `json:"allow_creator"`
	MoverIDs       []uuid.UUID `json:"mover_ids"`
}

// WorkflowResponse is a project's workflow configuration.
type WorkflowResponse struct {
	Configured  bool                       `json:"configured"`
	Enabled     bool                       `json:"enabled"`
	Bypass      bool                       `json:"bypass"`
	States      []WorkflowStateConfig      `json:"states"`
	Transitions []WorkflowTransitionConfig `json:"transitions"`
}

// SaveWorkflowRequest replaces a project's workflow graph.
type SaveWorkflowRequest struct {
	States      []WorkflowStateConfig      `json:"states"`
	Transitions []WorkflowTransitionConfig `json:"transitions"`
}

// SetWorkflowEnabledRequest toggles workflow enforcement.
type SetWorkflowEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

// GetWorkflow returns the project's workflow configuration.
//
//	@Summary		Get workflow
//	@Description	Get the project's workflow states, transitions and rules.
//	@Tags			Workflow
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Success		200			{object}	WorkflowResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/workflow [get]
func (h *TaskHandler) GetWorkflow(c *echo.Context) error {
	projectID, err := uuid.Parse(c.Request().Header.Get(auth.HeaderProjectID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}
	resp, err := h.loadWorkflow(c, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load workflow")
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *TaskHandler) loadWorkflow(c *echo.Context, projectID uuid.UUID) (WorkflowResponse, error) {
	ctx := c.Request().Context()
	resp := WorkflowResponse{
		Bypass:      c.Request().Header.Get(auth.HeaderProjectRole) == "admin",
		States:      []WorkflowStateConfig{},
		Transitions: []WorkflowTransitionConfig{},
	}

	wf, err := h.store.GetProjectWorkflow(ctx, projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return resp, nil
	}
	if err != nil {
		return resp, err
	}
	resp.Configured = true
	resp.Enabled = wf.Enabled

	states, err := h.store.ListWorkflowStates(ctx, projectID)
	if err != nil {
		return resp, err
	}
	actions, err := h.store.ListWorkflowStateActions(ctx, projectID)
	if err != nil {
		return resp, err
	}
	stateIdx := map[uuid.UUID]int{}
	for _, s := range states {
		stateIdx[s.StateID] = len(resp.States)
		resp.States = append(resp.States, WorkflowStateConfig{
			StateID: s.StateID, IsEntry: s.IsEntry, PosX: s.PosX, PosY: s.PosY,
			AddAssignees: []uuid.UUID{}, RemoveAssignees: []uuid.UUID{},
		})
	}
	for _, a := range actions {
		i, ok := stateIdx[a.StateID]
		if !ok {
			continue
		}
		if a.Action == workflow.ActionAddAssignee {
			resp.States[i].AddAssignees = append(resp.States[i].AddAssignees, a.UserID)
		} else {
			resp.States[i].RemoveAssignees = append(resp.States[i].RemoveAssignees, a.UserID)
		}
	}

	transitions, err := h.store.ListWorkflowTransitions(ctx, projectID)
	if err != nil {
		return resp, err
	}
	movers, err := h.store.ListWorkflowTransitionMovers(ctx, projectID)
	if err != nil {
		return resp, err
	}
	trIdx := map[uuid.UUID]int{}
	for _, t := range transitions {
		trIdx[t.ID] = len(resp.Transitions)
		resp.Transitions = append(resp.Transitions, WorkflowTransitionConfig{
			FromStateID: t.FromStateID, ToStateID: t.ToStateID,
			AllowAssignees: t.AllowAssignees, AllowCreator: t.AllowCreator,
			MoverIDs: []uuid.UUID{},
		})
	}
	for _, m := range movers {
		if i, ok := trIdx[m.TransitionID]; ok {
			resp.Transitions[i].MoverIDs = append(resp.Transitions[i].MoverIDs, m.UserID)
		}
	}
	return resp, nil
}

// SaveWorkflow replaces the project's workflow graph atomically.
//
//	@Summary		Save workflow
//	@Description	Replace the project's workflow states, transitions and rules. Enforcement is unchanged.
//	@Tags			Workflow
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string				true	"Project key"
//	@Param			body		body		SaveWorkflowRequest	true	"Workflow graph"
//	@Success		200			{object}	WorkflowResponse
//	@Failure		400			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/workflow [put]
func (h *TaskHandler) SaveWorkflow(c *echo.Context) error {
	projectID, err := uuid.Parse(c.Request().Header.Get(auth.HeaderProjectID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}
	var req SaveWorkflowRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	ctx := c.Request().Context()

	projectStates, err := h.store.ListProjectStates(ctx, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load states")
	}
	validStates := map[uuid.UUID]bool{}
	for _, s := range projectStates {
		validStates[s.ID] = true
	}
	members, err := h.store.ListProjectMembers(ctx, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load members")
	}
	validUsers := map[uuid.UUID]bool{}
	for _, m := range members {
		validUsers[m.UserID] = true
	}
	usersValid := func(ids []uuid.UUID) bool {
		for _, id := range ids {
			if !validUsers[id] {
				return false
			}
		}
		return true
	}

	hasEntry := false
	seenStates := map[uuid.UUID]bool{}
	for _, s := range req.States {
		if !validStates[s.StateID] || seenStates[s.StateID] {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid or duplicate state in workflow")
		}
		seenStates[s.StateID] = true
		if !usersValid(s.AddAssignees) || !usersValid(s.RemoveAssignees) {
			return echo.NewHTTPError(http.StatusBadRequest, "on-enter actions may only reference project members")
		}
		hasEntry = hasEntry || s.IsEntry
	}
	if !hasEntry {
		return echo.NewHTTPError(http.StatusBadRequest, "mark at least one state as an entry state")
	}
	seenEdges := map[[2]uuid.UUID]bool{}
	for _, t := range req.Transitions {
		key := [2]uuid.UUID{t.FromStateID, t.ToStateID}
		if t.FromStateID == t.ToStateID || !seenStates[t.FromStateID] || !seenStates[t.ToStateID] || seenEdges[key] {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid or duplicate transition in workflow")
		}
		seenEdges[key] = true
		if !usersValid(t.MoverIDs) {
			return echo.NewHTTPError(http.StatusBadRequest, "transition movers must be project members")
		}
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save workflow")
	}
	defer tx.Rollback(ctx)
	if err := saveWorkflowTx(c, store.New(tx), projectID, req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save workflow")
	}
	if err := tx.Commit(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save workflow")
	}

	resp, err := h.loadWorkflow(c, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load workflow")
	}
	return c.JSON(http.StatusOK, resp)
}

func saveWorkflowTx(c *echo.Context, q *store.Queries, projectID uuid.UUID, req SaveWorkflowRequest) error {
	ctx := c.Request().Context()
	if err := q.UpsertProjectWorkflow(ctx, projectID); err != nil {
		return err
	}
	if err := q.DeleteWorkflowTransitions(ctx, projectID); err != nil {
		return err
	}
	if err := q.DeleteWorkflowStates(ctx, projectID); err != nil {
		return err
	}
	for _, s := range req.States {
		if err := q.CreateWorkflowState(ctx, store.CreateWorkflowStateParams{
			StateID: s.StateID, ProjectID: projectID, IsEntry: s.IsEntry, PosX: s.PosX, PosY: s.PosY,
		}); err != nil {
			return err
		}
		for action, ids := range map[string][]uuid.UUID{
			workflow.ActionAddAssignee:    s.AddAssignees,
			workflow.ActionRemoveAssignee: s.RemoveAssignees,
		} {
			for _, id := range ids {
				if err := q.AddWorkflowStateAction(ctx, store.AddWorkflowStateActionParams{StateID: s.StateID, Action: action, UserID: id}); err != nil {
					return err
				}
			}
		}
	}
	for _, t := range req.Transitions {
		id, err := q.CreateWorkflowTransition(ctx, store.CreateWorkflowTransitionParams{
			ProjectID: projectID, FromStateID: t.FromStateID, ToStateID: t.ToStateID,
			AllowAssignees: t.AllowAssignees, AllowCreator: t.AllowCreator,
		})
		if err != nil {
			return err
		}
		for _, uid := range t.MoverIDs {
			if err := q.AddWorkflowTransitionMover(ctx, store.AddWorkflowTransitionMoverParams{TransitionID: id, UserID: uid}); err != nil {
				return err
			}
		}
	}
	return nil
}

// SetWorkflowEnabled switches workflow enforcement on or off, keeping the configuration.
//
//	@Summary		Toggle workflow
//	@Description	Enable or disable workflow enforcement. A workflow must be saved before it can be enabled.
//	@Tags			Workflow
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string						true	"Project key"
//	@Param			body		body		SetWorkflowEnabledRequest	true	"Enabled flag"
//	@Success		200			{object}	WorkflowResponse
//	@Failure		400			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/workflow/enabled [put]
func (h *TaskHandler) SetWorkflowEnabled(c *echo.Context) error {
	projectID, err := uuid.Parse(c.Request().Header.Get(auth.HeaderProjectID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}
	var req SetWorkflowEnabledRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	rows, err := h.store.SetProjectWorkflowEnabled(c.Request().Context(), store.SetProjectWorkflowEnabledParams{
		ProjectID: projectID, Enabled: req.Enabled,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update workflow")
	}
	if rows == 0 && req.Enabled {
		return echo.NewHTTPError(http.StatusBadRequest, "save a workflow before enabling it")
	}
	resp, err := h.loadWorkflow(c, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load workflow")
	}
	return c.JSON(http.StatusOK, resp)
}
