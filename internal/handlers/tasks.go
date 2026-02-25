package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/activity"
	"bereaucat/internal/auth"
	"bereaucat/internal/store"
)

// TaskHandler handles task-related endpoints.
type TaskHandler struct {
	store           store.Querier
	activityService *activity.Service
}

// NewTaskHandler creates a new task handler.
func NewTaskHandler(store store.Querier, activityService *activity.Service) *TaskHandler {
	return &TaskHandler{
		store:           store,
		activityService: activityService,
	}
}

// TaskResponse represents a task in API responses.
type TaskResponse struct {
	ID              uuid.UUID          `json:"id"`
	ProjectKey      string             `json:"project_key"`
	TaskNumber      int                `json:"task_number"`
	TaskID          string             `json:"task_id"` // e.g., "DEVOP-123"
	Title           string             `json:"title"`
	Description     *string            `json:"description,omitempty"`
	StateID         uuid.UUID          `json:"state_id"`
	StateName       string             `json:"state_name"`
	StateType       string             `json:"state_type"`
	StateColor      string             `json:"state_color"`
	Priority        int                `json:"priority"`
	CreatedBy        uuid.UUID          `json:"created_by"`
	CreatorUsername  string             `json:"creator_username"`
	CreatorFirstName string            `json:"creator_first_name"`
	CreatorLastName  string            `json:"creator_last_name"`
	Assignees       []AssigneeResponse `json:"assignees,omitempty"`
	Labels          []TaskLabelInfo    `json:"labels,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// AssigneeResponse represents a task assignee.
type AssigneeResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}

// TaskLabelInfo represents a label on a task.
type TaskLabelInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}

// CreateTaskRequest represents the request to create a task.
type CreateTaskRequest struct {
	Title       string   `json:"title"`
	Description *string  `json:"description"`
	StateID     *string  `json:"state_id"`
	Priority    *int     `json:"priority"`
	Assignees   []string `json:"assignees"`
	Labels      []string `json:"labels"`
}

// UpdateTaskRequest represents the request to update a task.
type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	StateID     *string `json:"state_id"`
	Priority    *int    `json:"priority"`
}

// PaginatedTasksResponse represents a paginated list of tasks.
type PaginatedTasksResponse struct {
	Tasks      []TaskResponse `json:"tasks"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

// ListTasks returns paginated list of tasks.
func (h *TaskHandler) ListTasks(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
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

	ctx := c.Request().Context()

	// Parse optional filter parameters
	var stateID, createdByID pgtype.UUID
	var stateType store.NullStateType
	var priority pgtype.Int4
	var search pgtype.Text

	if s := c.QueryParam("state_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			stateID = pgtype.UUID{Bytes: id, Valid: true}
		}
	}
	if s := c.QueryParam("state_type"); s != "" {
		stateType = store.NullStateType{StateType: store.StateType(s), Valid: true}
	}
	if s := c.QueryParam("created_by"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			createdByID = pgtype.UUID{Bytes: id, Valid: true}
		}
	}
	if s := c.QueryParam("priority"); s != "" {
		if p, err := strconv.Atoi(s); err == nil {
			priority = pgtype.Int4{Int32: int32(p), Valid: true}
		}
	}
	if s := c.QueryParam("q"); s != "" {
		search = pgtype.Text{String: s, Valid: true}
	}

	// Get filtered tasks
	tasks, err := h.store.ListProjectTasksFiltered(ctx, store.ListProjectTasksFilteredParams{
		ProjectID: projectID,
		Limit:     int32(perPage),
		Offset:    int32(offset),
		StateID:   stateID,
		StateType: stateType,
		CreatedBy: createdByID,
		Priority:  priority,
		Search:    search,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list tasks")
	}

	// Get total count with same filters
	total, err := h.store.CountProjectTasksFiltered(ctx, store.CountProjectTasksFilteredParams{
		ProjectID: projectID,
		StateID:   stateID,
		StateType: stateType,
		CreatedBy: createdByID,
		Priority:  priority,
		Search:    search,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count tasks")
	}

	// Convert to response format
	taskResponses := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		taskResponses[i] = TaskResponse{
			ID:              t.ID,
			ProjectKey:      t.ProjectKey,
			TaskNumber:      int(t.TaskNumber),
			TaskID:          t.ProjectKey + "-" + strconv.Itoa(int(t.TaskNumber)),
			Title:           t.Title,
			Description:     textToStringPtr(t.Description),
			StateID:         t.StateID,
			StateName:       t.StateName,
			StateType:       t.StateType,
			StateColor:      textToString(t.StateColor, "#6B7280"),
			Priority:        int(t.Priority),
			CreatedBy:       t.CreatedBy,
			CreatorUsername: t.CreatorUsername,
			CreatedAt:       t.CreatedAt.Time,
			UpdatedAt:       t.UpdatedAt.Time,
		}
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, PaginatedTasksResponse{
		Tasks:      taskResponses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// CreateTask creates a new task.
func (h *TaskHandler) CreateTask(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	projectKey := c.Request().Header.Get(auth.HeaderProjectKey)

	userIDStr := c.Request().Header.Get(auth.HeaderUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	var req CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	ctx := c.Request().Context()

	// Get or use default state
	var stateID uuid.UUID
	if req.StateID != nil {
		stateID, err = uuid.Parse(*req.StateID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid state_id")
		}
	} else {
		// Use default state
		defaultState, err := h.store.GetDefaultProjectState(ctx, projectID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get default state")
		}
		stateID = defaultState.ID
	}

	priority := int32(0)
	if req.Priority != nil {
		priority = int32(*req.Priority)
	}

	// Get next task number
	nextNumber, err := h.store.GetNextTaskNumber(ctx, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get next task number")
	}

	// Create task
	task, err := h.store.CreateTask(ctx, store.CreateTaskParams{
		ProjectID:   projectID,
		TaskNumber:  int32(nextNumber),
		Title:       req.Title,
		Description: stringToPgtypeText(req.Description),
		StateID:     stateID,
		Priority:    priority,
		CreatedBy:   userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create task")
	}

	// Log activity
	h.activityService.LogActivity(ctx, activity.LogActivityParams{
		TaskID:       task.ID,
		ActivityType: activity.TaskCreated,
		ActorID:      userID,
		NewValue: map[string]interface{}{
			"title":       task.Title,
			"description": textToStringPtr(task.Description),
			"state_id":    task.StateID.String(),
			"priority":    task.Priority,
		},
	})

	// Add assignees
	for _, assigneeIDStr := range req.Assignees {
		assigneeID, err := uuid.Parse(assigneeIDStr)
		if err != nil {
			continue
		}
		assigneeUser, err := h.store.GetUserByID(ctx, assigneeID)
		if err != nil {
			continue
		}
		_, err = h.store.AddTaskAssignee(ctx, store.AddTaskAssigneeParams{
			TaskID:     task.ID,
			UserID:     assigneeID,
			AssignedBy: userID,
		})
		if err != nil {
			continue
		}
		h.activityService.LogActivity(ctx, activity.LogActivityParams{
			TaskID:       task.ID,
			ActivityType: activity.AssigneeAdded,
			ActorID:      userID,
			NewValue: map[string]interface{}{
				"user_id":    assigneeID.String(),
				"username":   assigneeUser.Username,
				"first_name": assigneeUser.FirstName,
				"last_name":  assigneeUser.LastName,
			},
		})
	}

	// Add labels
	for _, labelIDStr := range req.Labels {
		labelID, err := uuid.Parse(labelIDStr)
		if err != nil {
			continue
		}
		label, err := h.store.GetProjectLabelByID(ctx, labelID)
		if err != nil {
			continue
		}
		err = h.store.AddTaskLabel(ctx, store.AddTaskLabelParams{
			TaskID:  task.ID,
			LabelID: labelID,
			AddedBy: userID,
		})
		if err != nil {
			continue
		}
		h.activityService.LogActivity(ctx, activity.LogActivityParams{
			TaskID:       task.ID,
			ActivityType: activity.LabelAdded,
			ActorID:      userID,
			NewValue: map[string]interface{}{
				"label_id": labelID.String(),
				"name":     label.Name,
				"color":    label.Color,
			},
		})
	}

	// Get full task with state info
	fullTask, err := h.store.GetTaskByID(ctx, task.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get task")
	}

	// Get assignees and labels
	assignees := h.getTaskAssignees(ctx, task.ID)
	labels := h.getTaskLabels(ctx, task.ID)

	return c.JSON(http.StatusCreated, TaskResponse{
		ID:              fullTask.ID,
		ProjectKey:      projectKey,
		TaskNumber:      int(fullTask.TaskNumber),
		TaskID:          projectKey + "-" + strconv.Itoa(int(fullTask.TaskNumber)),
		Title:           fullTask.Title,
		Description:     textToStringPtr(fullTask.Description),
		StateID:         fullTask.StateID,
		StateName:       fullTask.StateName,
		StateType:       fullTask.StateType,
		StateColor:      textToString(fullTask.StateColor, "#6B7280"),
		Priority:        int(fullTask.Priority),
		CreatedBy:        fullTask.CreatedBy,
		CreatorUsername:  fullTask.CreatorUsername,
		CreatorFirstName: fullTask.CreatorFirstName,
		CreatorLastName:  fullTask.CreatorLastName,
		Assignees:       assignees,
		Labels:          labels,
		CreatedAt:       fullTask.CreatedAt.Time,
		UpdatedAt:       fullTask.UpdatedAt.Time,
	})
}

// GetTask returns task details.
func (h *TaskHandler) GetTask(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	projectKey := c.Request().Header.Get(auth.HeaderProjectKey)

	taskNumStr := c.Param("taskNum")
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task number")
	}

	ctx := c.Request().Context()

	task, err := h.store.GetTaskByProjectAndNumber(ctx, store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	// Get assignees and labels
	assignees := h.getTaskAssignees(ctx, task.ID)
	labels := h.getTaskLabels(ctx, task.ID)

	return c.JSON(http.StatusOK, TaskResponse{
		ID:              task.ID,
		ProjectKey:      projectKey,
		TaskNumber:      int(task.TaskNumber),
		TaskID:          projectKey + "-" + strconv.Itoa(int(task.TaskNumber)),
		Title:           task.Title,
		Description:     textToStringPtr(task.Description),
		StateID:         task.StateID,
		StateName:       task.StateName,
		StateType:       task.StateType,
		StateColor:      textToString(task.StateColor, "#6B7280"),
		Priority:        int(task.Priority),
		CreatedBy:        task.CreatedBy,
		CreatorUsername:  task.CreatorUsername,
		CreatorFirstName: task.CreatorFirstName,
		CreatorLastName:  task.CreatorLastName,
		Assignees:       assignees,
		Labels:          labels,
		CreatedAt:       task.CreatedAt.Time,
		UpdatedAt:       task.UpdatedAt.Time,
	})
}

// UpdateTask updates a task.
func (h *TaskHandler) UpdateTask(c *echo.Context) error {
	projectIDStr := c.Request().Header.Get(auth.HeaderProjectID)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}

	projectKey := c.Request().Header.Get(auth.HeaderProjectKey)

	userIDStr := c.Request().Header.Get(auth.HeaderUserID)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	taskNumStr := c.Param("taskNum")
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task number")
	}

	var req UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ctx := c.Request().Context()

	// Get current task
	oldTask, err := h.store.GetTaskByProjectAndNumber(ctx, store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	// Parse state ID if provided
	var stateID pgtype.UUID
	if req.StateID != nil {
		id, err := uuid.Parse(*req.StateID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid state_id")
		}
		stateID = pgtype.UUID{Bytes: id, Valid: true}
	}

	// Update task
	task, err := h.store.UpdateTask(ctx, store.UpdateTaskParams{
		ID:          oldTask.ID,
		Title:       stringToPgtypeText(req.Title),
		Description: stringToPgtypeText(req.Description),
		StateID:     stateID,
		Priority:    intToPgtypeInt4(req.Priority),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update task")
	}

	// Log changes
	if req.Title != nil && *req.Title != oldTask.Title {
		h.activityService.LogActivity(ctx, activity.LogActivityParams{
			TaskID:       task.ID,
			ActivityType: activity.TaskUpdated,
			ActorID:      userID,
			FieldName:    activity.StringPtr("title"),
			OldValue:     oldTask.Title,
			NewValue:     *req.Title,
		})
	}
	oldDesc := textToStringPtr(oldTask.Description)
	if req.Description != nil && (oldDesc == nil || *req.Description != *oldDesc) {
		h.activityService.LogActivity(ctx, activity.LogActivityParams{
			TaskID:       task.ID,
			ActivityType: activity.TaskUpdated,
			ActorID:      userID,
			FieldName:    activity.StringPtr("description"),
			OldValue:     oldDesc,
			NewValue:     *req.Description,
		})
	}
	if stateID.Valid && stateID.Bytes != oldTask.StateID {
		// Get old and new state names for activity log
		oldState, _ := h.store.GetProjectStateByID(ctx, oldTask.StateID)
		newState, _ := h.store.GetProjectStateByID(ctx, uuid.UUID(stateID.Bytes))
		h.activityService.LogActivity(ctx, activity.LogActivityParams{
			TaskID:       task.ID,
			ActivityType: activity.StateChanged,
			ActorID:      userID,
			OldValue: map[string]interface{}{
				"state_id": oldTask.StateID.String(),
				"name":     oldState.Name,
			},
			NewValue: map[string]interface{}{
				"state_id": uuid.UUID(stateID.Bytes).String(),
				"name":     newState.Name,
			},
		})
	}
	if req.Priority != nil && int32(*req.Priority) != oldTask.Priority {
		h.activityService.LogActivity(ctx, activity.LogActivityParams{
			TaskID:       task.ID,
			ActivityType: activity.TaskUpdated,
			ActorID:      userID,
			FieldName:    activity.StringPtr("priority"),
			OldValue:     oldTask.Priority,
			NewValue:     *req.Priority,
		})
	}

	// Get updated task with state info
	fullTask, err := h.store.GetTaskByID(ctx, task.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get task")
	}

	// Get assignees and labels
	assignees := h.getTaskAssignees(ctx, task.ID)
	labels := h.getTaskLabels(ctx, task.ID)

	return c.JSON(http.StatusOK, TaskResponse{
		ID:              fullTask.ID,
		ProjectKey:      projectKey,
		TaskNumber:      int(fullTask.TaskNumber),
		TaskID:          projectKey + "-" + strconv.Itoa(int(fullTask.TaskNumber)),
		Title:           fullTask.Title,
		Description:     textToStringPtr(fullTask.Description),
		StateID:         fullTask.StateID,
		StateName:       fullTask.StateName,
		StateType:       fullTask.StateType,
		StateColor:      textToString(fullTask.StateColor, "#6B7280"),
		Priority:        int(fullTask.Priority),
		CreatedBy:        fullTask.CreatedBy,
		CreatorUsername:  fullTask.CreatorUsername,
		CreatorFirstName: fullTask.CreatorFirstName,
		CreatorLastName:  fullTask.CreatorLastName,
		Assignees:       assignees,
		Labels:          labels,
		CreatedAt:       fullTask.CreatedAt.Time,
		UpdatedAt:       fullTask.UpdatedAt.Time,
	})
}

// DeleteTask soft deletes a task.
func (h *TaskHandler) DeleteTask(c *echo.Context) error {
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

	taskNumStr := c.Param("taskNum")
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task number")
	}

	ctx := c.Request().Context()

	// Get task
	task, err := h.store.GetTaskByProjectAndNumber(ctx, store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	// Log deletion
	h.activityService.LogActivity(ctx, activity.LogActivityParams{
		TaskID:       task.ID,
		ActivityType: activity.TaskDeleted,
		ActorID:      userID,
		OldValue: map[string]interface{}{
			"title":       task.Title,
			"description": textToStringPtr(task.Description),
			"state_id":    task.StateID.String(),
			"priority":    task.Priority,
		},
	})

	// Soft delete
	err = h.store.SoftDeleteTask(ctx, task.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete task")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "task deleted"})
}

// AddAssigneeRequest represents the request to add an assignee.
type AddAssigneeRequest struct {
	UserID string `json:"user_id"`
}

// AddAssignee adds an assignee to a task.
func (h *TaskHandler) AddAssignee(c *echo.Context) error {
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

	taskNumStr := c.Param("taskNum")
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task number")
	}

	var req AddAssigneeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	assigneeID, err := uuid.Parse(req.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}

	ctx := c.Request().Context()

	// Get task
	task, err := h.store.GetTaskByProjectAndNumber(ctx, store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	// Check if already assigned
	isAssigned, err := h.store.IsTaskAssignee(ctx, store.IsTaskAssigneeParams{
		TaskID: task.ID,
		UserID: assigneeID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check assignee")
	}
	if isAssigned {
		return echo.NewHTTPError(http.StatusConflict, "user is already assigned")
	}

	// Get assignee user info for activity log
	assigneeUser, err := h.store.GetUserByID(ctx, assigneeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "assignee user not found")
	}

	// Add assignee
	_, err = h.store.AddTaskAssignee(ctx, store.AddTaskAssigneeParams{
		TaskID:     task.ID,
		UserID:     assigneeID,
		AssignedBy: userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add assignee")
	}

	// Log activity
	h.activityService.LogActivity(ctx, activity.LogActivityParams{
		TaskID:       task.ID,
		ActivityType: activity.AssigneeAdded,
		ActorID:      userID,
		NewValue: map[string]interface{}{
			"user_id":    assigneeID.String(),
			"username":   assigneeUser.Username,
			"first_name": assigneeUser.FirstName,
			"last_name":  assigneeUser.LastName,
		},
	})

	return c.JSON(http.StatusCreated, map[string]string{"message": "assignee added"})
}

// RemoveAssignee removes an assignee from a task.
func (h *TaskHandler) RemoveAssignee(c *echo.Context) error {
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

	taskNumStr := c.Param("taskNum")
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task number")
	}

	assigneeIDStr := c.Param("userId")
	assigneeID, err := uuid.Parse(assigneeIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	ctx := c.Request().Context()

	// Get task
	task, err := h.store.GetTaskByProjectAndNumber(ctx, store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	// Get assignee user info for activity log
	assigneeUser, err := h.store.GetUserByID(ctx, assigneeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "assignee user not found")
	}

	// Remove assignee
	err = h.store.RemoveTaskAssignee(ctx, store.RemoveTaskAssigneeParams{
		TaskID: task.ID,
		UserID: assigneeID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove assignee")
	}

	// Log activity
	h.activityService.LogActivity(ctx, activity.LogActivityParams{
		TaskID:       task.ID,
		ActivityType: activity.AssigneeRemoved,
		ActorID:      userID,
		OldValue: map[string]interface{}{
			"user_id":    assigneeID.String(),
			"username":   assigneeUser.Username,
			"first_name": assigneeUser.FirstName,
			"last_name":  assigneeUser.LastName,
		},
	})

	return c.JSON(http.StatusOK, map[string]string{"message": "assignee removed"})
}

// AddLabelRequest represents the request to add a label.
type AddLabelRequest struct {
	LabelID string `json:"label_id"`
}

// AddLabel adds a label to a task.
func (h *TaskHandler) AddLabel(c *echo.Context) error {
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

	taskNumStr := c.Param("taskNum")
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task number")
	}

	var req AddLabelRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	labelID, err := uuid.Parse(req.LabelID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid label_id")
	}

	ctx := c.Request().Context()

	// Get task
	task, err := h.store.GetTaskByProjectAndNumber(ctx, store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	// Check if label already on task
	hasLabel, err := h.store.HasTaskLabel(ctx, store.HasTaskLabelParams{
		TaskID:  task.ID,
		LabelID: labelID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check label")
	}
	if hasLabel {
		return echo.NewHTTPError(http.StatusConflict, "label already on task")
	}

	// Get label info for activity log
	label, err := h.store.GetProjectLabelByID(ctx, labelID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "label not found")
	}

	// Add label
	err = h.store.AddTaskLabel(ctx, store.AddTaskLabelParams{
		TaskID:  task.ID,
		LabelID: labelID,
		AddedBy: userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add label")
	}

	// Log activity
	h.activityService.LogActivity(ctx, activity.LogActivityParams{
		TaskID:       task.ID,
		ActivityType: activity.LabelAdded,
		ActorID:      userID,
		NewValue: map[string]interface{}{
			"label_id": labelID.String(),
			"name":     label.Name,
			"color":    label.Color,
		},
	})

	return c.JSON(http.StatusCreated, map[string]string{"message": "label added"})
}

// RemoveLabel removes a label from a task.
func (h *TaskHandler) RemoveLabel(c *echo.Context) error {
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

	taskNumStr := c.Param("taskNum")
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task number")
	}

	labelIDStr := c.Param("labelId")
	labelID, err := uuid.Parse(labelIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid label ID")
	}

	ctx := c.Request().Context()

	// Get task
	task, err := h.store.GetTaskByProjectAndNumber(ctx, store.GetTaskByProjectAndNumberParams{
		ProjectID:  projectID,
		TaskNumber: int32(taskNum),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	// Get label info for activity log
	label, err := h.store.GetProjectLabelByID(ctx, labelID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "label not found")
	}

	// Remove label
	err = h.store.RemoveTaskLabel(ctx, store.RemoveTaskLabelParams{
		TaskID:  task.ID,
		LabelID: labelID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove label")
	}

	// Log activity
	h.activityService.LogActivity(ctx, activity.LogActivityParams{
		TaskID:       task.ID,
		ActivityType: activity.LabelRemoved,
		ActorID:      userID,
		OldValue: map[string]interface{}{
			"label_id": labelID.String(),
			"name":     label.Name,
			"color":    label.Color,
		},
	})

	return c.JSON(http.StatusOK, map[string]string{"message": "label removed"})
}

// Helper functions

func (h *TaskHandler) getTaskAssignees(ctx context.Context, taskID uuid.UUID) []AssigneeResponse {
	assignees, err := h.store.ListTaskAssignees(ctx, taskID)
	if err != nil {
		return []AssigneeResponse{}
	}

	result := make([]AssigneeResponse, len(assignees))
	for i, a := range assignees {
		result[i] = AssigneeResponse{
			ID:        a.ID,
			UserID:    a.UserID,
			Username:  a.Username,
			Email:     a.Email,
			FirstName: a.FirstName,
			LastName:  a.LastName,
		}
	}
	return result
}

func (h *TaskHandler) getTaskLabels(ctx context.Context, taskID uuid.UUID) []TaskLabelInfo {
	labels, err := h.store.ListTaskLabels(ctx, taskID)
	if err != nil {
		return []TaskLabelInfo{}
	}

	result := make([]TaskLabelInfo, len(labels))
	for i, l := range labels {
		result[i] = TaskLabelInfo{
			ID:    l.LabelID,
			Name:  l.Name,
			Color: textToString(l.Color, "#3B82F6"),
		}
	}
	return result
}
