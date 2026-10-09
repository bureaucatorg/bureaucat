package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/activity"
	"bereaucat/internal/auth"
	"bereaucat/internal/store"
)

const (
	blockerRelationBlockedBy = "blocked_by"
	blockerRelationBlocking  = "blocking"
)

// TaskBlockersResponse lists the tasks blocking this task and the tasks it
// blocks, shaped like subtasks so the UI can reuse the subtask list.
type TaskBlockersResponse struct {
	BlockedBy []SubtaskResponse `json:"blocked_by"`
	Blocking  []SubtaskResponse `json:"blocking"`
}

// AddTaskBlockersRequest links existing tasks to this task. Relation is
// "blocked_by" (they block this task) or "blocking" (this task blocks them).
type AddTaskBlockersRequest struct {
	TaskIDs  []string `json:"task_ids"`
	Relation string   `json:"relation"`
}

func (h *TaskHandler) taskFromPath(c *echo.Context) (store.GetTaskByProjectAndNumberRow, error) {
	projectID, err := uuid.Parse(c.Request().Header.Get(auth.HeaderProjectID))
	if err != nil {
		return store.GetTaskByProjectAndNumberRow{}, echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}
	taskNum, err := strconv.Atoi(c.Param("taskNum"))
	if err != nil {
		return store.GetTaskByProjectAndNumberRow{}, echo.NewHTTPError(http.StatusBadRequest, "invalid task number")
	}
	task, err := h.store.GetTaskByProjectAndNumber(c.Request().Context(), store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return store.GetTaskByProjectAndNumberRow{}, echo.NewHTTPError(http.StatusNotFound, "task not found")
	}
	return task, nil
}

// ListTaskBlockers returns both sides of a task's blocker links.
//
//	@Summary		List task blockers
//	@Tags			Task Blockers
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			taskNum		path		int		true	"Task number"
//	@Success		200			{object}	TaskBlockersResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum}/blockers [get]
func (h *TaskHandler) ListTaskBlockers(c *echo.Context) error {
	task, err := h.taskFromPath(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	rows, err := h.store.ListTaskBlockerLinks(ctx, task.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list blockers")
	}

	ids := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	assigneesByTask, err := h.assigneesForTasks(ctx, ids)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load blocker assignees")
	}

	out := TaskBlockersResponse{BlockedBy: []SubtaskResponse{}, Blocking: []SubtaskResponse{}}
	for _, r := range rows {
		assignees := assigneesByTask[r.ID]
		if assignees == nil {
			assignees = []AssigneeResponse{}
		}
		item := SubtaskResponse{
			ID:               r.ID,
			ProjectKey:       r.ProjectKey,
			TaskNumber:       int(r.TaskNumber),
			TaskID:           r.ProjectKey + "-" + strconv.Itoa(int(r.TaskNumber)),
			Title:            r.Title,
			StateID:          r.StateID,
			StateName:        r.StateName,
			StateType:        r.StateType,
			StateColor:       textToString(r.StateColor, "#6B7280"),
			Priority:         int(r.Priority),
			CreatedBy:        r.CreatedBy,
			CreatorFirstName: r.CreatorFirstName,
			CreatorLastName:  r.CreatorLastName,
			CreatorAvatarURL: textToStringPtr(r.CreatorAvatarUrl),
			Assignees:        assignees,
		}
		if r.IsBlockedBy {
			out.BlockedBy = append(out.BlockedBy, item)
		} else {
			out.Blocking = append(out.Blocking, item)
		}
	}
	return c.JSON(http.StatusOK, out)
}

// ListBlockerCandidates serves the blocker picker: same-project tasks not yet
// linked to this task.
//
//	@Summary		List blocker candidates
//	@Tags			Task Blockers
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			taskNum		path		int		true	"Task number"
//	@Param			search		query		string	false	"Title search"
//	@Param			limit		query		int		false	"Max results (default 50, max 200)"
//	@Success		200			{array}		SubtaskCandidateResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum}/blockers/candidates [get]
func (h *TaskHandler) ListBlockerCandidates(c *echo.Context) error {
	task, err := h.taskFromPath(c)
	if err != nil {
		return err
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 200 {
		limit = 50
	}
	searchParam := pgtype.Text{}
	if s := strings.TrimSpace(c.QueryParam("search")); s != "" {
		searchParam = pgtype.Text{String: s, Valid: true}
	}

	rows, err := h.store.ListBlockerCandidates(c.Request().Context(), store.ListBlockerCandidatesParams{
		ProjectID: task.ProjectID,
		TaskID:    task.ID,
		Limit:     int32(limit),
		Search:    searchParam,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list candidate tasks")
	}

	out := make([]SubtaskCandidateResponse, len(rows))
	for i, t := range rows {
		out[i] = SubtaskCandidateResponse{
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
		}
	}
	return c.JSON(http.StatusOK, out)
}

// AddTaskBlockers links existing same-project tasks to this task in the given
// direction. All links are validated (including cycle checks) before any is
// written, and the whole batch is applied in one transaction.
//
//	@Summary		Add task blockers
//	@Tags			Task Blockers
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string					true	"Project key"
//	@Param			taskNum		path		int						true	"Task number"
//	@Param			body		body		AddTaskBlockersRequest	true	"Tasks and relation"
//	@Success		200			{object}	map[string]int
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum}/blockers [post]
func (h *TaskHandler) AddTaskBlockers(c *echo.Context) error {
	userID, err := uuid.Parse(c.Request().Header.Get(auth.HeaderUserID))
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user")
	}

	var req AddTaskBlockersRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if len(req.TaskIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "task_ids is required")
	}
	if req.Relation != blockerRelationBlockedBy && req.Relation != blockerRelationBlocking {
		return echo.NewHTTPError(http.StatusBadRequest, "relation must be 'blocked_by' or 'blocking'")
	}

	task, err := h.taskFromPath(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add blockers")
	}
	defer tx.Rollback(ctx)
	q := store.New(tx)

	if err := q.LockProjectBlockers(ctx, task.ProjectID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add blockers")
	}

	others := make([]store.GetBlockerTaskRefRow, 0, len(req.TaskIDs))
	for _, idStr := range req.TaskIDs {
		otherID, err := uuid.Parse(idStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid task id: "+idStr)
		}
		if otherID == task.ID {
			return echo.NewHTTPError(http.StatusBadRequest, "a task cannot block itself")
		}
		other, err := q.GetBlockerTaskRef(ctx, otherID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "task not found: "+idStr)
		}
		if other.ProjectID != task.ProjectID {
			return echo.NewHTTPError(http.StatusBadRequest, "task does not belong to this project: "+idStr)
		}
		blockedID, blockerID := blockerPair(task.ID, other.ID, req.Relation)
		cycle, err := q.BlockerLinkWouldCycle(ctx, store.BlockerLinkWouldCycleParams{
			BlockedID: blockedID,
			BlockerID: blockerID,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to add blockers")
		}
		if cycle {
			return echo.NewHTTPError(http.StatusBadRequest, "linking "+task.ProjectKey+"-"+strconv.Itoa(int(other.TaskNumber))+" would create a circular dependency")
		}
		others = append(others, other)
	}

	added := make([]store.GetBlockerTaskRefRow, 0, len(others))
	for _, other := range others {
		blockedID, blockerID := blockerPair(task.ID, other.ID, req.Relation)
		n, err := q.AddTaskBlocker(ctx, store.AddTaskBlockerParams{
			TaskID:        blockedID,
			BlockerTaskID: blockerID,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to add blockers")
		}
		if n > 0 {
			added = append(added, other)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add blockers")
	}

	for _, other := range added {
		h.logBlockerActivity(ctx, activity.BlockerAdded, task, other, req.Relation, userID)
	}

	return c.JSON(http.StatusOK, map[string]int{"added": len(added)})
}

// RemoveTaskBlocker removes the link between this task and another, in
// whichever direction it exists.
//
//	@Summary		Remove task blocker
//	@Tags			Task Blockers
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			taskNum		path		int		true	"Task number"
//	@Param			taskId		path		string	true	"Linked task UUID"
//	@Success		200			{object}	MessageResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum}/blockers/{taskId} [delete]
func (h *TaskHandler) RemoveTaskBlocker(c *echo.Context) error {
	userID, err := uuid.Parse(c.Request().Header.Get(auth.HeaderUserID))
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user")
	}
	otherID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
	}

	task, err := h.taskFromPath(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	other, err := h.store.GetBlockerTaskRef(ctx, otherID)
	if err != nil || other.ProjectID != task.ProjectID {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	removed, err := h.store.RemoveTaskBlockerLink(ctx, store.RemoveTaskBlockerLinkParams{
		AID: task.ID,
		BID: other.ID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove blocker")
	}
	if len(removed) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "link not found")
	}

	for _, r := range removed {
		relation := blockerRelationBlocking
		if r.TaskID == task.ID {
			relation = blockerRelationBlockedBy
		}
		h.logBlockerActivity(ctx, activity.BlockerRemoved, task, other, relation, userID)
	}

	return c.JSON(http.StatusOK, MessageResponse{Message: "blocker removed"})
}

// blockerPair returns (blocked, blocker) for a link between task and other.
func blockerPair(taskID, otherID uuid.UUID, relation string) (uuid.UUID, uuid.UUID) {
	if relation == blockerRelationBlockedBy {
		return taskID, otherID
	}
	return otherID, taskID
}

// logBlockerActivity records the link change on both tasks, each from its own
// point of view (field_name is the relation as seen from that task).
func (h *TaskHandler) logBlockerActivity(ctx context.Context, t activity.ActivityType, task store.GetTaskByProjectAndNumberRow, other store.GetBlockerTaskRefRow, relation string, actorID uuid.UUID) {
	inverse := blockerRelationBlockedBy
	if relation == blockerRelationBlockedBy {
		inverse = blockerRelationBlocking
	}
	sides := []struct {
		taskID   uuid.UUID
		relation string
		ref      map[string]interface{}
	}{
		{task.ID, relation, map[string]interface{}{
			"task_id": task.ProjectKey + "-" + strconv.Itoa(int(other.TaskNumber)),
			"title":   other.Title,
		}},
		{other.ID, inverse, map[string]interface{}{
			"task_id": task.ProjectKey + "-" + strconv.Itoa(int(task.TaskNumber)),
			"title":   task.Title,
		}},
	}
	for _, s := range sides {
		params := activity.LogActivityParams{
			TaskID:       s.taskID,
			ActivityType: t,
			ActorID:      actorID,
			FieldName:    activity.StringPtr(s.relation),
		}
		if t == activity.BlockerAdded {
			params.NewValue = s.ref
		} else {
			params.OldValue = s.ref
		}
		h.activityService.LogActivity(ctx, params)
	}
}
