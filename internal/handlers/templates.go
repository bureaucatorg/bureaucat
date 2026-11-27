package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"

	"bereaucat/internal/auth"
	"bereaucat/internal/store"
)

const (
	FieldText     = "text"
	FieldTextarea = "textarea"
	FieldNumber   = "number"
	FieldSelect   = "select"

	maxTemplateFields  = 50
	maxFieldLabelLen   = 200
	maxFieldOptions    = 100
	maxFieldValueLen   = 10000
	templateNoResponse = "No response"
)

var templateFieldKeyRe = regexp.MustCompile(`^[a-z0-9_]{1,50}$`)

// TemplateField is one input of a structured template.
type TemplateField struct {
	Key         string   `json:"key"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Required    bool     `json:"required"`
	Options     []string `json:"options,omitempty"`
	Multiple    bool     `json:"multiple,omitempty"`
	Min         *float64 `json:"min,omitempty"`
	Max         *float64 `json:"max,omitempty"`
}

// TemplateResponse represents a task template in API responses.
type TemplateResponse struct {
	ID               uuid.UUID       `json:"id"`
	Name             string          `json:"name"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	Fields           []TemplateField `json:"fields"`
	DefaultPriority  *int            `json:"default_priority"`
	DefaultLabels    []uuid.UUID     `json:"default_labels"`
	DefaultAssignees []uuid.UUID     `json:"default_assignees"`
	CreatedBy        uuid.UUID       `json:"created_by"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// CreateTemplateRequest represents the request to create a template.
type CreateTemplateRequest struct {
	Name             string          `json:"name"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	Fields           []TemplateField `json:"fields"`
	DefaultPriority  *int            `json:"default_priority"`
	DefaultLabels    []uuid.UUID     `json:"default_labels"`
	DefaultAssignees []uuid.UUID     `json:"default_assignees"`
}

// UpdateTemplateRequest represents the request to update a template.
// default_priority is only changed when the key is present; null clears it.
type UpdateTemplateRequest struct {
	Name             *string          `json:"name"`
	Title            *string          `json:"title"`
	Description      *string          `json:"description"`
	Fields           *[]TemplateField `json:"fields"`
	DefaultPriority  NullableInt      `json:"default_priority"`
	DefaultLabels    *[]uuid.UUID     `json:"default_labels"`
	DefaultAssignees *[]uuid.UUID     `json:"default_assignees"`
}

// NullableInt distinguishes an absent JSON key from an explicit null.
type NullableInt struct {
	Set   bool
	Value *int
}

func (n *NullableInt) UnmarshalJSON(b []byte) error {
	n.Set = true
	if string(b) == "null" {
		n.Value = nil
		return nil
	}
	var v int
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	n.Value = &v
	return nil
}

func toTemplateResponse(t store.TaskTemplate) TemplateResponse {
	fields := []TemplateField{}
	if len(t.Fields) > 0 {
		_ = json.Unmarshal(t.Fields, &fields)
	}
	var prio *int
	if t.DefaultPriority.Valid {
		p := int(t.DefaultPriority.Int32)
		prio = &p
	}
	labels := t.DefaultLabels
	if labels == nil {
		labels = []uuid.UUID{}
	}
	assignees := t.DefaultAssignees
	if assignees == nil {
		assignees = []uuid.UUID{}
	}
	return TemplateResponse{
		ID:               t.ID,
		Name:             t.Name,
		Title:            t.Title,
		Description:      t.Description,
		Fields:           fields,
		DefaultPriority:  prio,
		DefaultLabels:    labels,
		DefaultAssignees: assignees,
		CreatedBy:        t.CreatedBy,
		CreatedAt:        t.CreatedAt.Time,
		UpdatedAt:        t.UpdatedAt.Time,
	}
}

// normalizeTemplateFields trims and validates a field list, dropping
// attributes that don't apply to a field's type.
func normalizeTemplateFields(fields []TemplateField) ([]TemplateField, error) {
	if len(fields) > maxTemplateFields {
		return nil, fmt.Errorf("a template can have at most %d fields", maxTemplateFields)
	}
	out := make([]TemplateField, 0, len(fields))
	seen := map[string]bool{}
	for i, f := range fields {
		f.Key = strings.TrimSpace(f.Key)
		f.Label = strings.TrimSpace(f.Label)
		f.Description = strings.TrimSpace(f.Description)
		if f.Label == "" || len(f.Label) > maxFieldLabelLen {
			return nil, fmt.Errorf("field %d: label is required (max %d characters)", i+1, maxFieldLabelLen)
		}
		if !templateFieldKeyRe.MatchString(f.Key) {
			return nil, fmt.Errorf("field %q: key must be 1-50 lowercase letters, digits or underscores", f.Label)
		}
		if seen[f.Key] {
			return nil, fmt.Errorf("field %q: duplicate key %q", f.Label, f.Key)
		}
		seen[f.Key] = true

		switch f.Type {
		case FieldText, FieldTextarea:
			f.Options, f.Multiple, f.Min, f.Max = nil, false, nil, nil
		case FieldNumber:
			f.Options, f.Multiple = nil, false
			if f.Min != nil && f.Max != nil && *f.Min > *f.Max {
				return nil, fmt.Errorf("field %q: min cannot be greater than max", f.Label)
			}
		case FieldSelect:
			f.Min, f.Max = nil, nil
			opts := make([]string, 0, len(f.Options))
			optSeen := map[string]bool{}
			for _, o := range f.Options {
				o = strings.TrimSpace(o)
				if o == "" || optSeen[o] {
					continue
				}
				optSeen[o] = true
				opts = append(opts, o)
			}
			if len(opts) == 0 || len(opts) > maxFieldOptions {
				return nil, fmt.Errorf("field %q: a select needs 1-%d options", f.Label, maxFieldOptions)
			}
			f.Options = opts
		default:
			return nil, fmt.Errorf("field %q: type must be one of text, textarea, number, select", f.Label)
		}
		out = append(out, f)
	}
	return out, nil
}

// checkTemplateDefaults ensures default labels belong to the project and
// default assignees are project members.
func checkTemplateDefaults(ctx context.Context, q store.Querier, projectID uuid.UUID, prio *int, labels, assignees []uuid.UUID) error {
	if prio != nil && (*prio < 0 || *prio > 4) {
		return errors.New("default_priority must be between 0 and 4")
	}
	for _, id := range labels {
		l, err := q.GetProjectLabelByID(ctx, id)
		if err != nil || l.ProjectID != projectID {
			return fmt.Errorf("default label %s does not belong to this project", id)
		}
	}
	for _, id := range assignees {
		ok, err := q.IsProjectMember(ctx, store.IsProjectMemberParams{ProjectID: projectID, UserID: id})
		if err != nil || !ok {
			return fmt.Errorf("default assignee %s is not a project member", id)
		}
	}
	return nil
}

type renderedField struct {
	label  string
	values []string
	multi  bool
}

// checkTemplateValues validates submitted values against the template's
// fields, returning them in field order ready for rendering.
func checkTemplateValues(fields []TemplateField, values map[string]any) ([]renderedField, error) {
	known := map[string]bool{}
	for _, f := range fields {
		known[f.Key] = true
	}
	var problems []string
	for k := range values {
		if !known[k] {
			problems = append(problems, fmt.Sprintf("unknown field %q", k))
		}
	}

	out := make([]renderedField, 0, len(fields))
	for _, f := range fields {
		vals, err := fieldValue(f, values[f.Key])
		if err != "" {
			problems = append(problems, fmt.Sprintf("%q %s", f.Label, err))
			continue
		}
		if f.Required && len(vals) == 0 {
			problems = append(problems, fmt.Sprintf("%q is required", f.Label))
			continue
		}
		out = append(out, renderedField{label: f.Label, values: vals, multi: f.Multiple})
	}
	if len(problems) > 0 {
		return nil, errors.New("template fields invalid: " + strings.Join(problems, "; "))
	}
	return out, nil
}

func fieldValue(f TemplateField, raw any) ([]string, string) {
	if raw == nil {
		return nil, ""
	}
	switch f.Type {
	case FieldText, FieldTextarea:
		s, ok := raw.(string)
		if !ok {
			return nil, "must be a string"
		}
		s = strings.TrimSpace(s)
		if len(s) > maxFieldValueLen {
			return nil, fmt.Sprintf("must be at most %d characters", maxFieldValueLen)
		}
		if f.Type == FieldText {
			s = strings.Join(strings.Fields(s), " ")
		}
		if s == "" {
			return nil, ""
		}
		return []string{s}, ""
	case FieldNumber:
		var n float64
		switch v := raw.(type) {
		case float64:
			n = v
		case string:
			v = strings.TrimSpace(v)
			if v == "" {
				return nil, ""
			}
			p, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, "must be a number"
			}
			n = p
		default:
			return nil, "must be a number"
		}
		if f.Min != nil && n < *f.Min {
			return nil, fmt.Sprintf("must be at least %s", strconv.FormatFloat(*f.Min, 'f', -1, 64))
		}
		if f.Max != nil && n > *f.Max {
			return nil, fmt.Sprintf("must be at most %s", strconv.FormatFloat(*f.Max, 'f', -1, 64))
		}
		return []string{strconv.FormatFloat(n, 'f', -1, 64)}, ""
	case FieldSelect:
		var picked []string
		switch v := raw.(type) {
		case string:
			if v != "" {
				picked = []string{v}
			}
		case []any:
			for _, item := range v {
				s, ok := item.(string)
				if !ok {
					return nil, "must be a list of options"
				}
				picked = append(picked, s)
			}
		default:
			return nil, "must be one of the listed options"
		}
		if !f.Multiple && len(picked) > 1 {
			return nil, "accepts a single option"
		}
		allowed := map[string]bool{}
		for _, o := range f.Options {
			allowed[o] = true
		}
		seen := map[string]bool{}
		out := []string{}
		for _, p := range picked {
			if !allowed[p] {
				return nil, fmt.Sprintf("has invalid option %q (allowed: %s)", p, strings.Join(f.Options, ", "))
			}
			if !seen[p] {
				seen[p] = true
				out = append(out, p)
			}
		}
		if len(out) == 0 {
			return nil, ""
		}
		return out, ""
	}
	return nil, "has an unsupported type"
}

// renderTemplateDescription matches the format of the free-form extra text:
// the UI treats descriptions starting with "<" as HTML and the rest as markdown.
func renderTemplateDescription(fields []renderedField, extra string) string {
	extraEmpty := strings.TrimSpace(extra) == "" || strings.TrimSpace(extra) == "<p></p>"
	if !extraEmpty && !strings.HasPrefix(extra, "<") {
		var b strings.Builder
		for _, f := range fields {
			b.WriteString("### " + f.label + "\n\n")
			switch {
			case len(f.values) == 0:
				b.WriteString("_" + templateNoResponse + "_\n\n")
			case f.multi:
				for _, v := range f.values {
					b.WriteString("- " + v + "\n")
				}
				b.WriteString("\n")
			default:
				b.WriteString(f.values[0] + "\n\n")
			}
		}
		b.WriteString("### Additional details\n\n" + extra)
		return b.String()
	}

	var b strings.Builder
	for _, f := range fields {
		b.WriteString("<h3>" + html.EscapeString(f.label) + "</h3>")
		switch {
		case len(f.values) == 0:
			b.WriteString("<p><em>" + templateNoResponse + "</em></p>")
		case f.multi:
			b.WriteString("<ul>")
			for _, v := range f.values {
				b.WriteString("<li><p>" + html.EscapeString(v) + "</p></li>")
			}
			b.WriteString("</ul>")
		default:
			lines := strings.Split(f.values[0], "\n")
			for i := range lines {
				lines[i] = html.EscapeString(strings.TrimRight(lines[i], "\r"))
			}
			b.WriteString("<p>" + strings.Join(lines, "<br>") + "</p>")
		}
	}
	if !extraEmpty {
		b.WriteString("<h3>Additional details</h3>" + extra)
	}
	return b.String()
}

// errTemplateRequired is returned when a project requires a template and none was given.
var errTemplateRequired = errors.New("this project requires a task template; create the task from a template")

// applyTemplate validates a create request against its template (or the
// project's require-template setting) and fills in the description and defaults.
func applyTemplate(ctx context.Context, q store.Querier, projectID uuid.UUID, req *CreateTaskRequest) error {
	if req.TemplateID == nil || *req.TemplateID == "" {
		if req.ParentTaskNumber != nil {
			return nil
		}
		project, err := q.GetProjectByID(ctx, projectID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to load project")
		}
		if project.RequireTemplate {
			return echo.NewHTTPError(http.StatusBadRequest, errTemplateRequired.Error())
		}
		if len(req.Fields) > 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "fields require template_id")
		}
		return nil
	}

	templateID, err := uuid.Parse(*req.TemplateID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid template_id")
	}
	tmpl, err := q.GetTaskTemplateByID(ctx, templateID)
	if err != nil || tmpl.ProjectID != projectID {
		return echo.NewHTTPError(http.StatusBadRequest, "template not found in this project")
	}
	t := toTemplateResponse(tmpl)

	if len(t.Fields) > 0 {
		rendered, err := checkTemplateValues(t.Fields, req.Fields)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		extra := ""
		if req.Description != nil {
			extra = *req.Description
		}
		desc := renderTemplateDescription(rendered, extra)
		req.Description = &desc
	} else if len(req.Fields) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "this template has no fields")
	}

	if strings.TrimSpace(req.Title) == "" {
		req.Title = t.Title
	}
	if req.Priority == nil && t.DefaultPriority != nil {
		req.Priority = t.DefaultPriority
	}
	if req.Labels == nil {
		for _, id := range t.DefaultLabels {
			if l, err := q.GetProjectLabelByID(ctx, id); err == nil && l.ProjectID == projectID {
				req.Labels = append(req.Labels, id.String())
			}
		}
	}
	if req.Assignees == nil {
		for _, id := range t.DefaultAssignees {
			if ok, err := q.IsProjectMember(ctx, store.IsProjectMemberParams{ProjectID: projectID, UserID: id}); err == nil && ok {
				req.Assignees = append(req.Assignees, id.String())
			}
		}
	}
	return nil
}

func templateProjectID(c *echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Request().Header.Get(auth.HeaderProjectID))
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusInternalServerError, "invalid project ID in context")
	}
	return id, nil
}

// loadProjectTemplate fetches a template and ensures it belongs to the project in the URL.
func (h *ProjectHandler) loadProjectTemplate(c *echo.Context, projectID uuid.UUID) (store.TaskTemplate, error) {
	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		return store.TaskTemplate{}, echo.NewHTTPError(http.StatusBadRequest, "invalid template ID")
	}
	t, err := h.store.GetTaskTemplateByID(c.Request().Context(), templateID)
	if err != nil || t.ProjectID != projectID {
		return store.TaskTemplate{}, echo.NewHTTPError(http.StatusNotFound, "template not found")
	}
	return t, nil
}

func priorityToInt4(p *int) pgtype.Int4 {
	if p == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*p), Valid: true}
}

// ListTemplates returns project task templates.
//
//	@Summary		List templates
//	@Description	Returns all task templates for a project.
//	@Tags			Project Templates
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Success		200			{array}		TemplateResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/templates [get]
func (h *ProjectHandler) ListTemplates(c *echo.Context) error {
	projectID, err := templateProjectID(c)
	if err != nil {
		return err
	}
	templates, err := h.store.ListTaskTemplates(c.Request().Context(), projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list templates")
	}
	resp := make([]TemplateResponse, len(templates))
	for i, t := range templates {
		resp[i] = toTemplateResponse(t)
	}
	return c.JSON(http.StatusOK, resp)
}

// CreateTemplate creates a new task template.
//
//	@Summary		Create template
//	@Description	Create a new task template. Requires project admin role.
//	@Tags			Project Templates
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string					true	"Project key"
//	@Param			body		body		CreateTemplateRequest	true	"Template details"
//	@Success		201			{object}	TemplateResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/templates [post]
func (h *ProjectHandler) CreateTemplate(c *echo.Context) error {
	projectID, err := templateProjectID(c)
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(c.Request().Header.Get(auth.HeaderUserID))
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID")
	}

	var req CreateTemplateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if strings.TrimSpace(req.Name) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	ctx := c.Request().Context()
	fields, err := normalizeTemplateFields(req.Fields)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.DefaultLabels == nil {
		req.DefaultLabels = []uuid.UUID{}
	}
	if req.DefaultAssignees == nil {
		req.DefaultAssignees = []uuid.UUID{}
	}
	if err := checkTemplateDefaults(ctx, h.store, projectID, req.DefaultPriority, req.DefaultLabels, req.DefaultAssignees); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	fieldsJSON, _ := json.Marshal(fields)

	t, err := h.store.CreateTaskTemplate(ctx, store.CreateTaskTemplateParams{
		ProjectID:        projectID,
		Name:             req.Name,
		Title:            req.Title,
		Description:      req.Description,
		CreatedBy:        userID,
		Fields:           fieldsJSON,
		DefaultPriority:  priorityToInt4(req.DefaultPriority),
		DefaultLabels:    req.DefaultLabels,
		DefaultAssignees: req.DefaultAssignees,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create template")
	}
	return c.JSON(http.StatusCreated, toTemplateResponse(t))
}

// UpdateTemplate updates a task template.
//
//	@Summary		Update template
//	@Description	Update a task template. Requires project admin role.
//	@Tags			Project Templates
//	@Accept			json
//	@Produce		json
//	@Param			projectKey	path		string					true	"Project key"
//	@Param			templateId	path		string					true	"Template ID"
//	@Param			body		body		UpdateTemplateRequest	true	"Fields to update"
//	@Success		200			{object}	TemplateResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/templates/{templateId} [patch]
func (h *ProjectHandler) UpdateTemplate(c *echo.Context) error {
	projectID, err := templateProjectID(c)
	if err != nil {
		return err
	}
	existing, err := h.loadProjectTemplate(c, projectID)
	if err != nil {
		return err
	}

	var req UpdateTemplateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name cannot be empty")
	}

	ctx := c.Request().Context()
	var fieldsJSON []byte
	if req.Fields != nil {
		fields, err := normalizeTemplateFields(*req.Fields)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		fieldsJSON, _ = json.Marshal(fields)
	}
	var labels, assignees []uuid.UUID
	if req.DefaultLabels != nil {
		labels = append([]uuid.UUID{}, *req.DefaultLabels...)
	}
	if req.DefaultAssignees != nil {
		assignees = append([]uuid.UUID{}, *req.DefaultAssignees...)
	}
	if err := checkTemplateDefaults(ctx, h.store, projectID, req.DefaultPriority.Value, labels, assignees); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	t, err := h.store.UpdateTaskTemplate(ctx, store.UpdateTaskTemplateParams{
		ID:                 existing.ID,
		Name:               stringToPgtypeText(req.Name),
		Title:              stringToPgtypeText(req.Title),
		Description:        stringToPgtypeText(req.Description),
		Fields:             fieldsJSON,
		SetDefaultPriority: req.DefaultPriority.Set,
		DefaultPriority:    priorityToInt4(req.DefaultPriority.Value),
		DefaultLabels:      labels,
		DefaultAssignees:   assignees,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update template")
	}
	return c.JSON(http.StatusOK, toTemplateResponse(t))
}

// DeleteTemplate deletes a task template.
//
//	@Summary		Delete template
//	@Description	Delete a task template. Requires project admin role.
//	@Tags			Project Templates
//	@Produce		json
//	@Param			projectKey	path		string	true	"Project key"
//	@Param			templateId	path		string	true	"Template ID"
//	@Success		200			{object}	MessageResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/projects/{projectKey}/templates/{templateId} [delete]
func (h *ProjectHandler) DeleteTemplate(c *echo.Context) error {
	projectID, err := templateProjectID(c)
	if err != nil {
		return err
	}
	existing, err := h.loadProjectTemplate(c, projectID)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	project, err := h.store.GetProjectByID(ctx, projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load project")
	}
	if project.RequireTemplate {
		templates, err := h.store.ListTaskTemplates(ctx, projectID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list templates")
		}
		if len(templates) <= 1 {
			return echo.NewHTTPError(http.StatusBadRequest, "turn off 'require template' before deleting the last template")
		}
	}
	if err := h.store.DeleteTaskTemplate(ctx, existing.ID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete template")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "template deleted"})
}
