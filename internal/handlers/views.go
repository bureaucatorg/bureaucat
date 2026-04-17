package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/auth"
	"bereaucat/internal/store"
)

// ViewHandler exposes CRUD for saved project views.
type ViewHandler struct {
	store store.Querier
}

func NewViewHandler(s store.Querier) *ViewHandler {
	return &ViewHandler{store: s}
}

// ViewResponse is the API shape for a saved view.
type ViewResponse struct {
	ID          uuid.UUID       `json:"id"`
	ProjectID   uuid.UUID       `json:"project_id"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Visibility  string          `json:"visibility"`
	OwnerID     uuid.UUID       `json:"owner_id"`
	FilterTree  json.RawMessage `json:"filter_tree"`
	GroupBy     string          `json:"group_by"`
	SortBy      string          `json:"sort_by"`
	SortDir     string          `json:"sort_dir"`
	DefaultTab  string          `json:"default_tab"`
	Position    int             `json:"position"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// CreateViewRequest is the body for POST /views.
type CreateViewRequest struct {
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Visibility  string          `json:"visibility"` // "private" or "shared"
	FilterTree  json.RawMessage `json:"filter_tree"`
	GroupBy     string          `json:"group_by"`
	SortBy      string          `json:"sort_by"`
	SortDir     string          `json:"sort_dir"`
	DefaultTab  string          `json:"default_tab"` // "tasks" or "board"
}

// UpdateViewRequest is the body for PATCH /views/:slug. All fields optional.
type UpdateViewRequest struct {
	Name        *string          `json:"name"`
	Description *string          `json:"description"`
	Visibility  *string          `json:"visibility"`
	FilterTree  *json.RawMessage `json:"filter_tree"`
	GroupBy     *string          `json:"group_by"`
	SortBy      *string          `json:"sort_by"`
	SortDir     *string          `json:"sort_dir"`
	DefaultTab  *string          `json:"default_tab"`
	Position    *int             `json:"position"`
}

// ReorderViewsRequest is the body for PATCH /views/reorder.
type ReorderViewsRequest struct {
	Items []ReorderViewItem `json:"items"`
}

type ReorderViewItem struct {
	ID       uuid.UUID `json:"id"`
	Position int       `json:"position"`
}

// Valid group-by and sort values — closed list, mirrored on the frontend.
var (
	validGroupBy = map[string]struct{}{
		"state": {}, "state_type": {}, "priority": {}, "assignee": {},
		"label": {}, "due_bucket": {},
		"start_bucket": {}, "created_bucket": {}, "updated_bucket": {},
	}
	validSortBy = map[string]struct{}{
		"created_at": {}, "updated_at": {}, "priority": {},
		"due_date": {}, "start_date": {}, "title": {},
	}
	validSortDir = map[string]struct{}{
		"asc": {}, "desc": {},
	}
	validVisibility = map[string]struct{}{
		"private": {}, "shared": {},
	}
	validDefaultTab = map[string]struct{}{
		"tasks": {}, "board": {},
	}
)

// viewData holds the common fields needed by viewToResponse and access checks.
// Each sqlc-generated Row type is converted to this before use.
type viewData struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	Slug        string
	Name        string
	Description pgtype.Text
	Visibility  store.ViewVisibility
	OwnerID     uuid.UUID
	FilterTree  []byte
	GroupBy     string
	SortBy      string
	SortDir     string
	DefaultTab  string
	Position    int32
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
}

func fromListRow(v store.ListProjectViewsRow) viewData {
	return viewData{ID: v.ID, ProjectID: v.ProjectID, Slug: v.Slug, Name: v.Name, Description: v.Description, Visibility: v.Visibility, OwnerID: v.OwnerID, FilterTree: v.FilterTree, GroupBy: v.GroupBy, SortBy: v.SortBy, SortDir: v.SortDir, DefaultTab: v.DefaultTab, Position: v.Position, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

func fromCreateRow(v store.CreateProjectViewRow) viewData {
	return viewData{ID: v.ID, ProjectID: v.ProjectID, Slug: v.Slug, Name: v.Name, Description: v.Description, Visibility: v.Visibility, OwnerID: v.OwnerID, FilterTree: v.FilterTree, GroupBy: v.GroupBy, SortBy: v.SortBy, SortDir: v.SortDir, DefaultTab: v.DefaultTab, Position: v.Position, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

func fromSlugRow(v store.GetProjectViewBySlugRow) viewData {
	return viewData{ID: v.ID, ProjectID: v.ProjectID, Slug: v.Slug, Name: v.Name, Description: v.Description, Visibility: v.Visibility, OwnerID: v.OwnerID, FilterTree: v.FilterTree, GroupBy: v.GroupBy, SortBy: v.SortBy, SortDir: v.SortDir, DefaultTab: v.DefaultTab, Position: v.Position, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

func fromIDRow(v store.GetProjectViewByIDRow) viewData {
	return viewData{ID: v.ID, ProjectID: v.ProjectID, Slug: v.Slug, Name: v.Name, Description: v.Description, Visibility: v.Visibility, OwnerID: v.OwnerID, FilterTree: v.FilterTree, GroupBy: v.GroupBy, SortBy: v.SortBy, SortDir: v.SortDir, DefaultTab: v.DefaultTab, Position: v.Position, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

func fromUpdateRow(v store.UpdateProjectViewRow) viewData {
	return viewData{ID: v.ID, ProjectID: v.ProjectID, Slug: v.Slug, Name: v.Name, Description: v.Description, Visibility: v.Visibility, OwnerID: v.OwnerID, FilterTree: v.FilterTree, GroupBy: v.GroupBy, SortBy: v.SortBy, SortDir: v.SortDir, DefaultTab: v.DefaultTab, Position: v.Position, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

func viewToResponse(v viewData) ViewResponse {
	resp := ViewResponse{
		ID:         v.ID,
		ProjectID:  v.ProjectID,
		Slug:       v.Slug,
		Name:       v.Name,
		Visibility: string(v.Visibility),
		OwnerID:    v.OwnerID,
		FilterTree: v.FilterTree,
		GroupBy:    v.GroupBy,
		SortBy:     v.SortBy,
		SortDir:    v.SortDir,
		DefaultTab: v.DefaultTab,
		Position:   int(v.Position),
		CreatedAt:  v.CreatedAt.Time,
		UpdatedAt:  v.UpdatedAt.Time,
	}
	if v.Description.Valid {
		s := v.Description.String
		resp.Description = &s
	}
	return resp
}

func assertViewWriteAccess(v viewData, callerID uuid.UUID, role string) error {
	if v.OwnerID == callerID {
		return nil
	}
	if v.Visibility == "shared" && role == "admin" {
		return nil
	}
	return echo.NewHTTPError(http.StatusForbidden, "not allowed to edit this view")
}

// ListViews returns all views visible to the caller in the project (owner's
// private views plus any shared view).
//
//	@Summary		List project views
//	@Tags			Views
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Success		200			{array}		ViewResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/views [get]
func (h *ViewHandler) ListViews(c *echo.Context) error {
	projectID, callerID, err := projectAndCaller(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	rows, err := h.store.ListProjectViews(ctx, store.ListProjectViewsParams{
		ProjectID: projectID,
		OwnerID:   callerID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list views")
	}
	out := make([]ViewResponse, len(rows))
	for i, v := range rows {
		out[i] = viewToResponse(fromListRow(v))
	}
	return c.JSON(http.StatusOK, out)
}

// CreateView inserts a new saved view.
//
//	@Summary		Create a saved view
//	@Tags			Views
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string			true	"Project key"
//	@Param			body		body		CreateViewRequest	true	"View"
//	@Success		201			{object}	ViewResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/views [post]
func (h *ViewHandler) CreateView(c *echo.Context) error {
	projectID, callerID, err := projectAndCaller(c)
	if err != nil {
		return err
	}
	var req CreateViewRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if strings.TrimSpace(req.Name) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	if req.Visibility == "" {
		req.Visibility = "private"
	}
	if _, ok := validVisibility[req.Visibility]; !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid visibility")
	}
	if req.GroupBy == "" {
		req.GroupBy = "state_type"
	}
	if _, ok := validGroupBy[req.GroupBy]; !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid group_by")
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if _, ok := validSortBy[req.SortBy]; !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid sort_by")
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}
	if _, ok := validSortDir[req.SortDir]; !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid sort_dir")
	}
	if req.DefaultTab == "" {
		req.DefaultTab = "tasks"
	}
	if _, ok := validDefaultTab[req.DefaultTab]; !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid default_tab")
	}
	if len(req.FilterTree) == 0 {
		req.FilterTree = []byte(`{"children":[]}`)
	}
	if _, err := store.ParseFilterTree(req.FilterTree); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid filter_tree: "+err.Error())
	}

	ctx := c.Request().Context()
	slug, err := h.uniqueSlug(ctx, projectID, req.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate slug")
	}

	descParam := pgtype.Text{}
	if req.Description != nil {
		descParam = pgtype.Text{String: *req.Description, Valid: true}
	}

	view, err := h.store.CreateProjectView(ctx, store.CreateProjectViewParams{
		ProjectID:   projectID,
		Slug:        slug,
		Name:        req.Name,
		Description: descParam,
		Visibility:  store.ViewVisibility(req.Visibility),
		OwnerID:     callerID,
		FilterTree:  req.FilterTree,
		GroupBy:     req.GroupBy,
		SortBy:      req.SortBy,
		SortDir:     req.SortDir,
		DefaultTab:  req.DefaultTab,
		Position:    0,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create view")
	}
	return c.JSON(http.StatusCreated, viewToResponse(fromCreateRow(view)))
}

// GetView returns a single view.
//
//	@Summary		Get a saved view
//	@Tags			Views
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			slug		path		string	true	"View slug"
//	@Success		200			{object}	ViewResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/views/{slug} [get]
func (h *ViewHandler) GetView(c *echo.Context) error {
	projectID, callerID, err := projectAndCaller(c)
	if err != nil {
		return err
	}
	view, err := h.loadViewForRead(c.Request().Context(), projectID, callerID, c.Param("slug"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, viewToResponse(view))
}

// UpdateView patches a view. Owner always allowed; project admins may edit
// shared views but not another user's private view.
//
//	@Summary		Update a saved view
//	@Tags			Views
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string			true	"Project key"
//	@Param			slug		path		string			true	"View slug"
//	@Param			body		body		UpdateViewRequest	true	"Fields to update"
//	@Success		200			{object}	ViewResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/views/{slug} [patch]
func (h *ViewHandler) UpdateView(c *echo.Context) error {
	projectID, callerID, err := projectAndCaller(c)
	if err != nil {
		return err
	}
	role := c.Request().Header.Get(auth.HeaderProjectRole)
	ctx := c.Request().Context()

	slugRow, err := h.store.GetProjectViewBySlug(ctx, store.GetProjectViewBySlugParams{
		ProjectID: projectID,
		Slug:      c.Param("slug"),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "view not found")
	}
	view := fromSlugRow(slugRow)
	if err := assertViewWriteAccess(view, callerID, role); err != nil {
		return err
	}

	var req UpdateViewRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	params := store.UpdateProjectViewParams{ID: view.ID}
	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	}
	if req.Description != nil {
		params.Description = pgtype.Text{String: *req.Description, Valid: true}
	}
	if req.Visibility != nil {
		// Only owner can flip visibility.
		if view.OwnerID != callerID {
			return echo.NewHTTPError(http.StatusForbidden, "only the owner can change visibility")
		}
		if _, ok := validVisibility[*req.Visibility]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid visibility")
		}
		params.Visibility = store.NullViewVisibility{ViewVisibility: store.ViewVisibility(*req.Visibility), Valid: true}
	}
	if req.FilterTree != nil {
		if _, err := store.ParseFilterTree(*req.FilterTree); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid filter_tree: "+err.Error())
		}
		params.FilterTree = *req.FilterTree
	}
	if req.GroupBy != nil {
		if _, ok := validGroupBy[*req.GroupBy]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid group_by")
		}
		params.GroupBy = pgtype.Text{String: *req.GroupBy, Valid: true}
	}
	if req.SortBy != nil {
		if _, ok := validSortBy[*req.SortBy]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid sort_by")
		}
		params.SortBy = pgtype.Text{String: *req.SortBy, Valid: true}
	}
	if req.SortDir != nil {
		if _, ok := validSortDir[*req.SortDir]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid sort_dir")
		}
		params.SortDir = pgtype.Text{String: *req.SortDir, Valid: true}
	}
	if req.DefaultTab != nil {
		if _, ok := validDefaultTab[*req.DefaultTab]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid default_tab")
		}
		params.DefaultTab = pgtype.Text{String: *req.DefaultTab, Valid: true}
	}
	if req.Position != nil {
		params.Position = pgtype.Int4{Int32: int32(*req.Position), Valid: true}
	}

	updated, err := h.store.UpdateProjectView(ctx, params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update view")
	}
	return c.JSON(http.StatusOK, viewToResponse(fromUpdateRow(updated)))
}

// DeleteView soft-deletes a view.
//
//	@Summary		Delete a saved view
//	@Tags			Views
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			slug		path		string	true	"View slug"
//	@Success		200			{object}	MessageResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/views/{slug} [delete]
func (h *ViewHandler) DeleteView(c *echo.Context) error {
	projectID, callerID, err := projectAndCaller(c)
	if err != nil {
		return err
	}
	role := c.Request().Header.Get(auth.HeaderProjectRole)
	ctx := c.Request().Context()
	slugRow, err := h.store.GetProjectViewBySlug(ctx, store.GetProjectViewBySlugParams{
		ProjectID: projectID,
		Slug:      c.Param("slug"),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "view not found")
	}
	vd := fromSlugRow(slugRow)
	if err := assertViewWriteAccess(vd, callerID, role); err != nil {
		return err
	}
	if err := h.store.SoftDeleteProjectView(ctx, vd.ID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete view")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "view deleted"})
}

// ReorderViews updates positions for multiple views in a single request.
// Only writes positions for views the caller can edit (owner's own + admin-editable shared).
//
//	@Summary		Reorder saved views
//	@Tags			Views
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string					true	"Project key"
//	@Param			body		body		ReorderViewsRequest	true	"Ordered list"
//	@Success		200			{object}	MessageResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/views/reorder [patch]
func (h *ViewHandler) ReorderViews(c *echo.Context) error {
	projectID, callerID, err := projectAndCaller(c)
	if err != nil {
		return err
	}
	role := c.Request().Header.Get(auth.HeaderProjectRole)
	var req ReorderViewsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if len(req.Items) == 0 {
		return c.JSON(http.StatusOK, map[string]string{"message": "nothing to reorder"})
	}
	ctx := c.Request().Context()

	// Load each view to enforce access before writing.
	type reorderRow struct {
		ID          uuid.UUID `json:"id"`
		NewPosition int       `json:"new_position"`
	}
	allowed := make([]reorderRow, 0, len(req.Items))
	for _, item := range req.Items {
		idRow, err := h.store.GetProjectViewByID(ctx, item.ID)
		if err != nil {
			continue
		}
		v := fromIDRow(idRow)
		if v.ProjectID != projectID {
			continue
		}
		if err := assertViewWriteAccess(v, callerID, role); err != nil {
			continue // skip silently; don't leak which IDs are protected
		}
		allowed = append(allowed, reorderRow{ID: item.ID, NewPosition: item.Position})
	}
	if len(allowed) == 0 {
		return c.JSON(http.StatusOK, map[string]string{"message": "nothing to reorder"})
	}
	payload, err := json.Marshal(allowed)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to encode reorder payload")
	}
	if err := h.store.ReorderProjectViews(ctx, store.ReorderProjectViewsParams{
		ProjectID: projectID,
		Items:     payload,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to reorder views")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "reordered"})
}

// ----- helpers -----

// projectAndCaller pulls project ID and user ID from the middleware headers.
func projectAndCaller(c *echo.Context) (uuid.UUID, uuid.UUID, error) {
	projectID, err := uuid.Parse(c.Request().Header.Get(auth.HeaderProjectID))
	if err != nil {
		return uuid.Nil, uuid.Nil, echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID")
	}
	callerID, err := uuid.Parse(c.Request().Header.Get(auth.HeaderUserID))
	if err != nil {
		return uuid.Nil, uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}
	return projectID, callerID, nil
}

// loadViewForRead fetches a view and rejects private views the caller doesn't own.
func (h *ViewHandler) loadViewForRead(ctx context.Context, projectID, callerID uuid.UUID, slug string) (viewData, error) {
	v, err := h.store.GetProjectViewBySlug(ctx, store.GetProjectViewBySlugParams{
		ProjectID: projectID,
		Slug:      slug,
	})
	if err != nil {
		return viewData{}, echo.NewHTTPError(http.StatusNotFound, "view not found")
	}
	vd := fromSlugRow(v)
	if v.Visibility == "private" && v.OwnerID != callerID {
		return viewData{}, echo.NewHTTPError(http.StatusForbidden, "view not accessible")
	}
	return vd, nil
}

// slug generation — lowercase-alphanumeric-dashes, 64 chars max, with a random
// suffix added when there's a collision within the same project.
var slugCleanRE = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugCleanRE.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "view"
	}
	if len(s) > 48 {
		s = s[:48]
	}
	return s
}

func randomSuffix() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (h *ViewHandler) uniqueSlug(ctx context.Context, projectID uuid.UUID, name string) (string, error) {
	base := slugify(name)
	for attempt := 0; attempt < 5; attempt++ {
		candidate := base
		if attempt > 0 {
			candidate = base + "-" + randomSuffix()
		}
		exists, err := h.store.ProjectViewSlugExists(ctx, store.ProjectViewSlugExistsParams{
			ProjectID: projectID,
			Slug:      candidate,
		})
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", errors.New("could not generate unique slug")
}
