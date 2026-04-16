package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/activity"
	"bereaucat/internal/auth"
	"bereaucat/internal/notifier"
	"bereaucat/internal/store"
)

// NullableTime distinguishes between an absent JSON field, an explicit null
// (clear the value), and a provided time value.
type NullableTime struct {
	Set   bool
	Value *time.Time
}

func (n *NullableTime) UnmarshalJSON(data []byte) error {
	n.Set = true
	if string(data) == "null" {
		return nil
	}
	var t time.Time
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	n.Value = &t
	return nil
}

func timestamptzToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	tt := t.Time
	return &tt
}

func timePtrToTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func timePtrEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}

// TaskHandler handles task-related endpoints.
type TaskHandler struct {
	store               store.Querier
	activityService     *activity.Service
	notificationService *notifier.Service
}

// NewTaskHandler creates a new task handler.
func NewTaskHandler(store store.Querier, activityService *activity.Service, notificationService *notifier.Service) *TaskHandler {
	return &TaskHandler{
		store:               store,
		activityService:     activityService,
		notificationService: notificationService,
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
	StartDate        *time.Time         `json:"start_date,omitempty"`
	DueDate          *time.Time         `json:"due_date,omitempty"`
	CreatedBy        uuid.UUID          `json:"created_by"`
	CreatorUsername  string             `json:"creator_username"`
	CreatorFirstName string            `json:"creator_first_name"`
	CreatorLastName  string            `json:"creator_last_name"`
	CreatorAvatarURL *string           `json:"creator_avatar_url,omitempty"`
	Assignees       []AssigneeResponse `json:"assignees,omitempty"`
	Labels          []TaskLabelInfo    `json:"labels,omitempty"`
	CommentCount    int                `json:"comment_count"`
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
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

// TaskLabelInfo represents a label on a task.
type TaskLabelInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}

// CreateTaskRequest represents the request to create a task.
type CreateTaskRequest struct {
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	StateID     *string    `json:"state_id"`
	Priority    *int       `json:"priority"`
	StartDate   *time.Time `json:"start_date"`
	DueDate     *time.Time `json:"due_date"`
	Assignees   []string   `json:"assignees"`
	Labels      []string   `json:"labels"`
}

// UpdateTaskRequest represents the request to update a task.
type UpdateTaskRequest struct {
	Title       *string      `json:"title"`
	Description *string      `json:"description"`
	StateID     *string      `json:"state_id"`
	Priority    *int         `json:"priority"`
	StartDate   NullableTime `json:"start_date"`
	DueDate     NullableTime `json:"due_date"`
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
//
//	@Summary		List tasks
//	@Description	Returns a paginated list of tasks with optional filters.
//	@Tags			Tasks
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(20)
//	@Param			state_id	query		string	false	"Filter by state ID"
//	@Param			state_type	query		string	false	"Filter by state type"
//	@Param			created_by	query		string	false	"Filter by creator ID"
//	@Param			priority	query		int		false	"Filter by priority"
//	@Param			q			query		string	false	"Search by title"
//	@Success		200			{object}	PaginatedTasksResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks [get]
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
	var stateID, createdByID, assignedToID pgtype.UUID
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
	if s := c.QueryParam("assigned_to"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			assignedToID = pgtype.UUID{Bytes: id, Valid: true}
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

	var fromDate, toDate pgtype.Timestamptz
	if s := c.QueryParam("from_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			fromDate = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}
	if s := c.QueryParam("to_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			// Set to end of day
			toDate = pgtype.Timestamptz{Time: t.Add(24*time.Hour - time.Nanosecond), Valid: true}
		}
	}

	// Get filtered tasks
	tasks, err := h.store.ListProjectTasksFiltered(ctx, store.ListProjectTasksFilteredParams{
		ProjectID:  projectID,
		Limit:      int32(perPage),
		Offset:     int32(offset),
		StateID:    stateID,
		StateType:  stateType,
		CreatedBy:  createdByID,
		AssignedTo: assignedToID,
		Priority:   priority,
		Search:     search,
		FromDate:   fromDate,
		ToDate:     toDate,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list tasks")
	}

	// Get total count with same filters
	total, err := h.store.CountProjectTasksFiltered(ctx, store.CountProjectTasksFilteredParams{
		ProjectID:  projectID,
		StateID:    stateID,
		StateType:  stateType,
		CreatedBy:  createdByID,
		AssignedTo: assignedToID,
		Priority:   priority,
		Search:     search,
		FromDate:   fromDate,
		ToDate:     toDate,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count tasks")
	}

	// Convert to response format with enrichment
	taskResponses := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		assignees := h.getTaskAssignees(ctx, t.ID)
		labels := h.getTaskLabels(ctx, t.ID)

		taskResponses[i] = TaskResponse{
			ID:               t.ID,
			ProjectKey:       t.ProjectKey,
			TaskNumber:       int(t.TaskNumber),
			TaskID:           t.ProjectKey + "-" + strconv.Itoa(int(t.TaskNumber)),
			Title:            t.Title,
			Description:      textToStringPtr(t.Description),
			StateID:          t.StateID,
			StateName:        t.StateName,
			StateType:        t.StateType,
			StateColor:       textToString(t.StateColor, "#6B7280"),
			Priority:         int(t.Priority),
			StartDate:        timestamptzToTimePtr(t.StartDate),
			DueDate:          timestamptzToTimePtr(t.DueDate),
			CreatedBy:        t.CreatedBy,
			CreatorUsername:   t.CreatorUsername,
			CreatorFirstName: t.CreatorFirstName,
			CreatorLastName:  t.CreatorLastName,
			CreatorAvatarURL: textToStringPtr(t.CreatorAvatarUrl),
			Assignees:        assignees,
			Labels:           labels,
			CommentCount:     int(t.CommentCount),
			CreatedAt:        t.CreatedAt.Time,
			UpdatedAt:        t.UpdatedAt.Time,
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
//
//	@Summary		Create task
//	@Description	Create a new task in the project.
//	@Tags			Tasks
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string				true	"Project key"
//	@Param			body		body		CreateTaskRequest	true	"Task details"
//	@Success		201			{object}	TaskResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks [post]
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

	if req.StartDate != nil && req.DueDate != nil && req.DueDate.Before(*req.StartDate) {
		return echo.NewHTTPError(http.StatusBadRequest, "due date cannot be before start date")
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
		StartDate:   timePtrToTimestamptz(req.StartDate),
		DueDate:     timePtrToTimestamptz(req.DueDate),
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

	// Send notifications for assignees and mentions
	if h.notificationService != nil {
		actorUser, _ := h.store.GetUserByID(ctx, userID)
		actorName := actorUser.FirstName + " " + actorUser.LastName
		if actorName == " " {
			actorName = actorUser.Username
		}
		taskNum := int(task.TaskNumber)
		baseURL := requestBaseURL(c)

		// Notify assignees
		for _, assigneeIDStr := range req.Assignees {
			assigneeID, err := uuid.Parse(assigneeIDStr)
			if err != nil || assigneeID == userID {
				continue
			}
			h.notificationService.Notify(ctx, notifier.Notification{
				Event:       notifier.EventTaskAssigned,
				RecipientID: assigneeID,
				ActorName:   actorName,
				ProjectKey:  projectKey,
				TaskNumber:  taskNum,
				TaskTitle:   req.Title,
				BaseURL:     baseURL,
			})
		}

		// Notify mentions in description
		if req.Description != nil {
			mentionedIDs := notifier.ParseMentions(*req.Description)
			for _, mentionedID := range mentionedIDs {
				if mentionedID == userID {
					continue
				}
				h.notificationService.Notify(ctx, notifier.Notification{
					Event:       notifier.EventMentioned,
					RecipientID: mentionedID,
					ActorName:   actorName,
					ProjectKey:  projectKey,
					TaskNumber:  taskNum,
					TaskTitle:   req.Title,
					BaseURL:     baseURL,
				})
			}
		}
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
		StartDate:       timestamptzToTimePtr(fullTask.StartDate),
		DueDate:         timestamptzToTimePtr(fullTask.DueDate),
		CreatedBy:        fullTask.CreatedBy,
		CreatorUsername:  fullTask.CreatorUsername,
		CreatorFirstName: fullTask.CreatorFirstName,
		CreatorLastName:  fullTask.CreatorLastName,
		CreatorAvatarURL: textToStringPtr(fullTask.CreatorAvatarUrl),
		Assignees:       assignees,
		Labels:          labels,
		CreatedAt:       fullTask.CreatedAt.Time,
		UpdatedAt:       fullTask.UpdatedAt.Time,
	})
}

// GetTask returns task details.
//
//	@Summary		Get task
//	@Description	Returns task details by task number.
//	@Tags			Tasks
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			taskNum		path		int		true	"Task number"
//	@Success		200			{object}	TaskResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum} [get]
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
		StartDate:       timestamptzToTimePtr(task.StartDate),
		DueDate:         timestamptzToTimePtr(task.DueDate),
		CreatedBy:        task.CreatedBy,
		CreatorUsername:  task.CreatorUsername,
		CreatorFirstName: task.CreatorFirstName,
		CreatorLastName:  task.CreatorLastName,
		CreatorAvatarURL: textToStringPtr(task.CreatorAvatarUrl),
		Assignees:       assignees,
		Labels:          labels,
		CreatedAt:       task.CreatedAt.Time,
		UpdatedAt:       task.UpdatedAt.Time,
	})
}

// UpdateTask updates a task.
//
//	@Summary		Update task
//	@Description	Update task fields. Changes are logged in the activity log.
//	@Tags			Tasks
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string				true	"Project key"
//	@Param			taskNum		path		int					true	"Task number"
//	@Param			body		body		UpdateTaskRequest	true	"Fields to update"
//	@Success		200			{object}	TaskResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum} [patch]
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

	// Build nullable date args: only applied when the field was present in the request.
	var startDateArg, dueDateArg pgtype.Timestamptz
	if req.StartDate.Set && req.StartDate.Value != nil {
		startDateArg = pgtype.Timestamptz{Time: *req.StartDate.Value, Valid: true}
	}
	if req.DueDate.Set && req.DueDate.Value != nil {
		dueDateArg = pgtype.Timestamptz{Time: *req.DueDate.Value, Valid: true}
	}

	// Validate against the task's post-update state: start must not be after due.
	effectiveStart := timestamptzToTimePtr(oldTask.StartDate)
	if req.StartDate.Set {
		effectiveStart = req.StartDate.Value
	}
	effectiveDue := timestamptzToTimePtr(oldTask.DueDate)
	if req.DueDate.Set {
		effectiveDue = req.DueDate.Value
	}
	if effectiveStart != nil && effectiveDue != nil && effectiveDue.Before(*effectiveStart) {
		return echo.NewHTTPError(http.StatusBadRequest, "due date cannot be before start date")
	}

	// Update task
	task, err := h.store.UpdateTask(ctx, store.UpdateTaskParams{
		ID:              oldTask.ID,
		Title:           stringToPgtypeText(req.Title),
		Description:     stringToPgtypeText(req.Description),
		StateID:         stateID,
		Priority:        intToPgtypeInt4(req.Priority),
		UpdateStartDate: req.StartDate.Set,
		StartDate:       startDateArg,
		UpdateDueDate:   req.DueDate.Set,
		DueDate:         dueDateArg,
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
	if req.StartDate.Set {
		oldStart := timestamptzToTimePtr(oldTask.StartDate)
		if !timePtrEqual(oldStart, req.StartDate.Value) {
			h.activityService.LogActivity(ctx, activity.LogActivityParams{
				TaskID:       task.ID,
				ActivityType: activity.TaskUpdated,
				ActorID:      userID,
				FieldName:    activity.StringPtr("start_date"),
				OldValue:     oldStart,
				NewValue:     req.StartDate.Value,
			})
		}
	}
	if req.DueDate.Set {
		oldDue := timestamptzToTimePtr(oldTask.DueDate)
		if !timePtrEqual(oldDue, req.DueDate.Value) {
			h.activityService.LogActivity(ctx, activity.LogActivityParams{
				TaskID:       task.ID,
				ActivityType: activity.TaskUpdated,
				ActorID:      userID,
				FieldName:    activity.StringPtr("due_date"),
				OldValue:     oldDue,
				NewValue:     req.DueDate.Value,
			})
		}
	}

	// Send mention notifications for newly added mentions in description
	if h.notificationService != nil && req.Description != nil {
		oldDescStr := ""
		if oldDesc != nil {
			oldDescStr = *oldDesc
		}
		newMentions := notifier.DiffMentions(oldDescStr, *req.Description)
		if len(newMentions) > 0 {
			actorUser, _ := h.store.GetUserByID(ctx, userID)
			actorName := actorUser.FirstName + " " + actorUser.LastName
			if actorName == " " {
				actorName = actorUser.Username
			}
			baseURL := requestBaseURL(c)
			for _, mentionedID := range newMentions {
				if mentionedID == userID {
					continue
				}
				h.notificationService.Notify(ctx, notifier.Notification{
					Event:       notifier.EventMentioned,
					RecipientID: mentionedID,
					ActorName:   actorName,
					ProjectKey:  projectKey,
					TaskNumber:  taskNum,
					TaskTitle:   oldTask.Title,
					BaseURL:     baseURL,
				})
			}
		}
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
		StartDate:       timestamptzToTimePtr(fullTask.StartDate),
		DueDate:         timestamptzToTimePtr(fullTask.DueDate),
		CreatedBy:        fullTask.CreatedBy,
		CreatorUsername:  fullTask.CreatorUsername,
		CreatorFirstName: fullTask.CreatorFirstName,
		CreatorLastName:  fullTask.CreatorLastName,
		CreatorAvatarURL: textToStringPtr(fullTask.CreatorAvatarUrl),
		Assignees:       assignees,
		Labels:          labels,
		CreatedAt:       fullTask.CreatedAt.Time,
		UpdatedAt:       fullTask.UpdatedAt.Time,
	})
}

// DeleteTask soft deletes a task.
//
//	@Summary		Delete task
//	@Description	Soft-delete a task. Requires project admin role or task creator.
//	@Tags			Tasks
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			taskNum		path		int		true	"Task number"
//	@Success		200			{object}	MessageResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum} [delete]
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

	// Only admins or the task creator can delete
	role := c.Request().Header.Get(auth.HeaderProjectRole)
	if role != "admin" && task.CreatedBy != userID {
		return echo.NewHTTPError(http.StatusForbidden, "only admins or the task creator can delete this task")
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
//
//	@Summary		Add assignee
//	@Description	Add a user as an assignee to a task.
//	@Tags			Task Assignees
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string				true	"Project key"
//	@Param			taskNum		path		int					true	"Task number"
//	@Param			body		body		AddAssigneeRequest	true	"Assignee details"
//	@Success		201			{object}	MessageResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		409			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum}/assignees [post]
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

	// Send notification to assignee (skip if self-assigning)
	if h.notificationService != nil && assigneeID != userID {
		username := c.Request().Header.Get(auth.HeaderUsername)
		actorUser, _ := h.store.GetUserByID(ctx, userID)
		actorName := actorUser.FirstName + " " + actorUser.LastName
		if actorName == " " {
			actorName = username
		}
		projectKey := c.Request().Header.Get(auth.HeaderProjectKey)
		h.notificationService.Notify(ctx, notifier.Notification{
			Event:       notifier.EventTaskAssigned,
			RecipientID: assigneeID,
			ActorName:   actorName,
			ProjectKey:  projectKey,
			TaskNumber:  taskNum,
			TaskTitle:   task.Title,
			BaseURL:     requestBaseURL(c),
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "assignee added"})
}

// RemoveAssignee removes an assignee from a task.
//
//	@Summary		Remove assignee
//	@Description	Remove a user from a task's assignees.
//	@Tags			Task Assignees
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			taskNum		path		int		true	"Task number"
//	@Param			userId		path		string	true	"User ID"
//	@Success		200			{object}	MessageResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum}/assignees/{userId} [delete]
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
//
//	@Summary		Add label to task
//	@Description	Add a project label to a task.
//	@Tags			Task Labels
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string			true	"Project key"
//	@Param			taskNum		path		int				true	"Task number"
//	@Param			body		body		AddLabelRequest	true	"Label details"
//	@Success		201			{object}	MessageResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		409			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum}/labels [post]
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
//
//	@Summary		Remove label from task
//	@Description	Remove a label from a task.
//	@Tags			Task Labels
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			taskNum		path		int		true	"Task number"
//	@Param			labelId		path		string	true	"Label ID"
//	@Success		200			{object}	MessageResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/tasks/{taskNum}/labels/{labelId} [delete]
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
			AvatarURL: textToStringPtr(a.AvatarUrl),
		}
	}
	return result
}

// requestBaseURL extracts the base URL from the request (e.g. "https://bureaucat.example.com").
func requestBaseURL(c *echo.Context) string {
	scheme := "https"
	if c.Request().TLS == nil {
		if proto := c.Request().Header.Get("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		} else {
			scheme = "http"
		}
	}
	return scheme + "://" + c.Request().Host
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
