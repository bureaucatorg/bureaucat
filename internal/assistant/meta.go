package assistant

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

// Meta is per-message context supplied by a channel adapter. It is stored as JSON so new
// kinds of context can be added without migrations.
type Meta struct {
	Page *PageContext `json:"page,omitempty"`
}

// PageContext is the web page the user was on when sending a message. It comes from the
// browser, so it is validated and only ever used as a hint; tools still enforce access.
type PageContext struct {
	Kind       string `json:"kind"`
	ProjectKey string `json:"project_key,omitempty"`
	TaskNumber int    `json:"task_number,omitempty"`
	PageNumber int    `json:"page_number,omitempty"`
	CycleID    string `json:"cycle_id,omitempty"`
	ModuleID   string `json:"module_id,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	Tab        string `json:"tab,omitempty"`
}

var (
	projectKeyRe = regexp.MustCompile(`^[A-Z0-9]{2,10}$`)
	tabRe        = regexp.MustCompile(`^[a-z_-]{1,20}$`)
)

var staticPages = map[string]string{
	"dashboard":      "the dashboard",
	"projects":       "the projects list",
	"workspaces":     "the workspaces list",
	"views":          "the saved views list",
	"graph":          "the task graph",
	"active_cycles":  "the active cycles overview",
	"active_modules": "the active modules overview",
	"settings":       "their settings",
}

// sanitize drops anything that does not match a known page shape.
func (m Meta) sanitize() Meta {
	if m.Page == nil {
		return Meta{}
	}
	p := *m.Page
	isUUID := func(s string) bool { _, err := uuid.Parse(s); return err == nil }
	needsProject := p.Kind == "project" || p.Kind == "task" || p.Kind == "page" || p.Kind == "cycle" || p.Kind == "module"
	ok := true
	switch {
	case staticPages[p.Kind] != "":
		p = PageContext{Kind: p.Kind}
	case needsProject && !projectKeyRe.MatchString(p.ProjectKey):
		ok = false
	case p.Kind == "project":
		if p.Tab != "" && !tabRe.MatchString(p.Tab) {
			p.Tab = ""
		}
		p = PageContext{Kind: p.Kind, ProjectKey: p.ProjectKey, Tab: p.Tab}
	case p.Kind == "task":
		ok = p.TaskNumber > 0
		p = PageContext{Kind: p.Kind, ProjectKey: p.ProjectKey, TaskNumber: p.TaskNumber}
	case p.Kind == "page":
		ok = p.PageNumber > 0
		p = PageContext{Kind: p.Kind, ProjectKey: p.ProjectKey, PageNumber: p.PageNumber}
	case p.Kind == "cycle":
		ok = isUUID(p.CycleID)
		p = PageContext{Kind: p.Kind, ProjectKey: p.ProjectKey, CycleID: p.CycleID}
	case p.Kind == "module":
		ok = isUUID(p.ModuleID)
		p = PageContext{Kind: p.Kind, ProjectKey: p.ProjectKey, ModuleID: p.ModuleID}
	case p.Kind == "profile":
		ok = isUUID(p.UserID)
		p = PageContext{Kind: p.Kind, UserID: p.UserID}
	default:
		ok = false
	}
	if !ok {
		return Meta{}
	}
	return Meta{Page: &p}
}

func (p PageContext) describe() string {
	switch p.Kind {
	case "project":
		if p.Tab != "" {
			return fmt.Sprintf("project %s (%s tab)", p.ProjectKey, p.Tab)
		}
		return "project " + p.ProjectKey
	case "task":
		return fmt.Sprintf("task %s-%d", p.ProjectKey, p.TaskNumber)
	case "page":
		return fmt.Sprintf("wiki page %d in project %s", p.PageNumber, p.ProjectKey)
	case "cycle":
		return fmt.Sprintf("cycle %s in project %s", p.CycleID, p.ProjectKey)
	case "module":
		return fmt.Sprintf("module %s in project %s", p.ModuleID, p.ProjectKey)
	case "profile":
		return "the profile of user " + p.UserID
	default:
		return staticPages[p.Kind]
	}
}
