package assistant

import (
	"context"
	"net/http"
	"strings"
)

const (
	InPath  = "path"
	InQuery = "query"
	InBody  = "body"
)

// Tool maps one model-callable tool to one Bureaucat API route. Confirm marks risky
// writes (deletes, moves, membership, project config) that wait for user approval.
type Tool struct {
	Name        string
	Title       string
	Description string
	Method      string
	Path        string
	Params      []Param
	Confirm     bool
}

// Param is a tool argument. Type is a JSON Schema type, "string[]"/"integer[]" for arrays,
// or "string|null" for a nullable string.
type Param struct {
	Name        string
	Type        string
	In          string
	Required    bool
	Description string
	Enum        []string
}

func (p Param) oneOf(values ...string) Param { p.Enum = values; return p }
func (p Param) req() Param                   { p.Required = true; return p }

func pathStr(name, desc string) Param { return Param{Name: name, Type: "string", In: InPath, Required: true, Description: desc} }
func pathInt(name, desc string) Param { return Param{Name: name, Type: "integer", In: InPath, Required: true, Description: desc} }
func query(name, typ, desc string) Param { return Param{Name: name, Type: typ, In: InQuery, Description: desc} }
func body(name, typ, desc string) Param  { return Param{Name: name, Type: typ, In: InBody, Description: desc} }

const filterHelp = `Optional FilterTree object; all children are ANDed: {"children":[{"predicate":{"field":F,"op":OP,"value":V}}]}.
Fields/ops: search contains(str, also matches KEY-12); title contains|not_contains|is|is_not(str), is_empty|is_set; description contains|not_contains, is_empty|is_set;
state in|not_in(state uuids); state_type in|not_in(backlog|unstarted|started|completed|cancelled); priority in|not_in(int[]), gte|lte(int);
created_by in|not_in(uuids or "@me"), is_me|is_not_me; assignees has_any|has_none|has_all(uuids or "@me"), is_empty|is_set; labels has_any|has_none|has_all(label uuids), is_empty|is_set;
start_date|due_date before|after(anchor), between({"from":a,"to":a}), is_empty|is_set; due_date overdue; created_at|updated_at before|after|between; comment_count eq|ne|gt|gte|lt|lte(int).
Anchors: YYYY-MM-DD or today, yesterday, tomorrow, this_week, last_week, next_week, this_month, last_month, next_month, last_7_days, last_30_days, last_90_days.`

var (
	pProject   = pathStr("project_key", "Project key, uppercase, e.g. ENG")
	pTask      = pathInt("task_number", "Task number within the project, e.g. 42 for ENG-42")
	pUser      = pathStr("user_id", "User UUID (the user_id field, not a membership id)")
	pCycle     = pathStr("cycle_id", "Cycle UUID")
	pModule    = pathStr("module_id", "Module UUID")
	pPage      = query("page", "integer", "Page number, default 1")
	pPerPage   = query("per_page", "integer", "Results per page, 1-100, default 20")
	pLimit     = query("limit", "integer", "Max results, 1-200, default 50")
	pSearch    = query("search", "string", "Case-insensitive title search")
	pTaskUUIDs = body("task_ids", "string[]", "Task UUIDs (the id field of a task, not task numbers)").req()
	pPriority  = body("priority", "integer", "0 none, 1 low, 2 medium, 3 high, 4 urgent")
	pColor     = body("color", "string", "Hex colour, e.g. #3B82F6")
	pUpload    = body("upload_id", "string", "upload_id of a file the user attached in this chat").req()
)

const (
	pp  = "/projects/{project_key}"
	ppt = pp + "/tasks/{task_number}"
)

var catalog = []Tool{
	// Me, users, search
	{Name: "get_me", Title: "Get my profile", Description: "Get the current user's profile.", Method: http.MethodGet, Path: "/me"},
	{Name: "list_my_tasks", Title: "List my tasks", Description: "Tasks assigned to the current user across all projects.", Method: http.MethodGet, Path: "/me/tasks",
		Params: []Param{pPage, pPerPage, query("workspace_id", "string", "Optional workspace UUID to scope results")}},
	{Name: "list_notifications", Title: "List notifications", Description: "The current user's notifications.", Method: http.MethodGet, Path: "/me/notifications",
		Params: []Param{pPage, pPerPage}},
	{Name: "mark_notification_read", Title: "Mark notification read", Description: "Mark one notification as read.", Method: http.MethodPost, Path: "/me/notifications/{notification_id}/read",
		Params: []Param{pathStr("notification_id", "Notification UUID")}},
	{Name: "mark_all_notifications_read", Title: "Mark all notifications read", Description: "Mark all notifications as read.", Method: http.MethodPost, Path: "/me/notifications/read_all"},
	{Name: "search", Title: "Search", Description: "Search tasks (title, description, KEY-N), cycles and projects the user can see. Max 8 results per category.", Method: http.MethodGet, Path: "/search",
		Params: []Param{query("q", "string", "Search text").req()}},
	{Name: "get_user", Title: "Get user", Description: "Get a user's profile by UUID.", Method: http.MethodGet, Path: "/users/{user_id}", Params: []Param{pUser}},
	{Name: "get_user_activity", Title: "Get user activity", Description: "Recent activity performed by a user.", Method: http.MethodGet, Path: "/users/{user_id}/activity",
		Params: []Param{pUser, pPage, pPerPage}},
	{Name: "list_active_cycles", Title: "List active cycles", Description: "Cycles running today across the user's projects.", Method: http.MethodGet, Path: "/cycles/active"},
	{Name: "list_active_modules", Title: "List active modules", Description: "In-progress modules across the user's projects.", Method: http.MethodGet, Path: "/modules/active"},

	// Workspaces
	{Name: "list_workspaces", Title: "List workspaces", Description: "Workspaces the user belongs to.", Method: http.MethodGet, Path: "/workspaces",
		Params: []Param{pPage, pPerPage, pSearch}},
	{Name: "get_workspace", Title: "Get workspace", Description: "Get a workspace by key.", Method: http.MethodGet, Path: "/workspaces/{workspace_key}",
		Params: []Param{pathStr("workspace_key", "Workspace key")}},
	{Name: "list_workspace_members", Title: "List workspace members", Description: "Members of a workspace (use user_id for user UUIDs).", Method: http.MethodGet, Path: "/workspaces/{workspace_key}/members",
		Params: []Param{pathStr("workspace_key", "Workspace key")}},

	// Projects
	{Name: "list_projects", Title: "List projects", Description: "Projects the user can access.", Method: http.MethodGet, Path: "/projects",
		Params: []Param{pPage, pPerPage, pSearch, query("workspace_id", "string", "Optional workspace UUID")}},
	{Name: "get_project", Title: "Get project", Description: "Get a project, including the user's role in it.", Method: http.MethodGet, Path: pp, Params: []Param{pProject}},
	{Name: "get_project_stats", Title: "Get project stats", Description: "Task counts, state buckets and per-member load for a project.", Method: http.MethodGet, Path: pp + "/stats", Params: []Param{pProject}},
	{Name: "create_project", Title: "Create project", Description: "Create a project with default states.", Method: http.MethodPost, Path: "/projects", Confirm: true,
		Params: []Param{
			body("project_key", "string", "2-10 chars, A-Z/0-9").req(),
			body("name", "string", "Project name").req(),
			body("description", "string", "Description"),
			body("workspace_id", "string", "Workspace UUID").req(),
		}},
	{Name: "update_project", Title: "Update project", Description: "Update project name/description (project admin).", Method: http.MethodPatch, Path: pp, Confirm: true,
		Params: []Param{pProject, body("name", "string", "New name"), body("description", "string", "New description")}},
	{Name: "delete_project", Title: "Delete project", Description: "Delete a project (project admin).", Method: http.MethodDelete, Path: pp, Confirm: true, Params: []Param{pProject}},

	// Project members
	{Name: "list_project_members", Title: "List project members", Description: "Members and roles of a project. Use user_id for user UUIDs.", Method: http.MethodGet, Path: pp + "/members", Params: []Param{pProject}},
	{Name: "search_users_for_project", Title: "Search users", Description: "Search the user directory to add to a project (project admin).", Method: http.MethodGet, Path: pp + "/members/search",
		Params: []Param{pProject, query("q", "string", "Name, username or email").req()}},
	{Name: "add_project_member", Title: "Add project member", Description: "Add a user to a project (project admin).", Method: http.MethodPost, Path: pp + "/members", Confirm: true,
		Params: []Param{pProject, body("user_id", "string", "User UUID").req(), body("role", "string", "Role").req().oneOf("guest", "member", "admin")}},
	{Name: "update_project_member_role", Title: "Change member role", Description: "Change a member's project role (project admin).", Method: http.MethodPatch, Path: pp + "/members/{user_id}", Confirm: true,
		Params: []Param{pProject, pUser, body("role", "string", "Role").req().oneOf("guest", "member", "admin")}},
	{Name: "remove_project_member", Title: "Remove project member", Description: "Remove a user from a project (project admin).", Method: http.MethodDelete, Path: pp + "/members/{user_id}", Confirm: true,
		Params: []Param{pProject, pUser}},

	// States
	{Name: "list_states", Title: "List states", Description: "Workflow states of a project.", Method: http.MethodGet, Path: pp + "/states", Params: []Param{pProject}},
	{Name: "create_state", Title: "Create state", Description: "Create a workflow state (project admin).", Method: http.MethodPost, Path: pp + "/states", Confirm: true,
		Params: []Param{pProject, body("name", "string", "State name").req(),
			body("state_type", "string", "State type").req().oneOf("backlog", "unstarted", "started", "completed", "cancelled"),
			pColor, body("position", "integer", "Sort position")}},
	{Name: "update_state", Title: "Update state", Description: "Rename/recolour/reorder a state (project admin).", Method: http.MethodPatch, Path: pp + "/states/{state_id}", Confirm: true,
		Params: []Param{pProject, pathStr("state_id", "State UUID"), body("name", "string", "Name"), pColor, body("position", "integer", "Sort position")}},
	{Name: "set_default_state", Title: "Set default state", Description: "Make a state the default for new tasks (project admin).", Method: http.MethodPost, Path: pp + "/states/{state_id}/default", Confirm: true,
		Params: []Param{pProject, pathStr("state_id", "State UUID")}},
	{Name: "delete_state", Title: "Delete state", Description: "Delete an unused, non-default state (project admin).", Method: http.MethodDelete, Path: pp + "/states/{state_id}", Confirm: true,
		Params: []Param{pProject, pathStr("state_id", "State UUID")}},

	// Labels
	{Name: "list_labels", Title: "List labels", Description: "Labels of a project.", Method: http.MethodGet, Path: pp + "/labels", Params: []Param{pProject}},
	{Name: "create_label", Title: "Create label", Description: "Create a project label.", Method: http.MethodPost, Path: pp + "/labels",
		Params: []Param{pProject, body("name", "string", "Label name").req(), pColor}},
	{Name: "update_label", Title: "Update label", Description: "Rename/recolour a label (project admin).", Method: http.MethodPatch, Path: pp + "/labels/{label_id}", Confirm: true,
		Params: []Param{pProject, pathStr("label_id", "Label UUID"), body("name", "string", "Name"), pColor}},
	{Name: "delete_label", Title: "Delete label", Description: "Delete a label (project admin).", Method: http.MethodDelete, Path: pp + "/labels/{label_id}", Confirm: true,
		Params: []Param{pProject, pathStr("label_id", "Label UUID")}},

	// Templates
	{Name: "list_templates", Title: "List templates", Description: "Task templates of a project.", Method: http.MethodGet, Path: pp + "/templates", Params: []Param{pProject}},
	{Name: "create_template", Title: "Create template", Description: "Create a task template (project admin).", Method: http.MethodPost, Path: pp + "/templates",
		Params: []Param{pProject, body("name", "string", "Template name").req(), body("title", "string", "Default task title"), body("description", "string", "Default task description (markdown)")}},
	{Name: "update_template", Title: "Update template", Description: "Update a task template (project admin).", Method: http.MethodPatch, Path: pp + "/templates/{template_id}", Confirm: true,
		Params: []Param{pProject, pathStr("template_id", "Template UUID"), body("name", "string", "Name"), body("title", "string", "Title"), body("description", "string", "Description")}},
	{Name: "delete_template", Title: "Delete template", Description: "Delete a task template (project admin).", Method: http.MethodDelete, Path: pp + "/templates/{template_id}", Confirm: true,
		Params: []Param{pProject, pathStr("template_id", "Template UUID")}},

	// Pages
	{Name: "list_pages", Title: "List pages", Description: "Wiki pages of a project (without content).", Method: http.MethodGet, Path: pp + "/pages",
		Params: []Param{pProject, query("q", "string", "Search title and content")}},
	{Name: "get_page", Title: "Get page", Description: "Get a page with its markdown content.", Method: http.MethodGet, Path: pp + "/pages/{page_number}",
		Params: []Param{pProject, pathInt("page_number", "Page number")}},
	{Name: "create_page", Title: "Create page", Description: "Create a wiki page.", Method: http.MethodPost, Path: pp + "/pages",
		Params: []Param{pProject, body("title", "string", "Title").req(), body("content", "string", "Markdown content")}},
	{Name: "update_page", Title: "Update page", Description: "Update a wiki page.", Method: http.MethodPatch, Path: pp + "/pages/{page_number}", Confirm: true,
		Params: []Param{pProject, pathInt("page_number", "Page number"), body("title", "string", "Title"), body("content", "string", "Markdown content (replaces existing)")}},
	{Name: "delete_page", Title: "Delete page", Description: "Delete a wiki page.", Method: http.MethodDelete, Path: pp + "/pages/{page_number}", Confirm: true,
		Params: []Param{pProject, pathInt("page_number", "Page number")}},

	// Views
	{Name: "list_views", Title: "List views", Description: "Saved task views (user's private + shared).", Method: http.MethodGet, Path: pp + "/views", Params: []Param{pProject}},
	{Name: "get_view", Title: "Get view", Description: "Get a saved view by slug.", Method: http.MethodGet, Path: pp + "/views/{slug}", Params: []Param{pProject, pathStr("slug", "View slug")}},
	{Name: "create_view", Title: "Create view", Description: "Create a saved task view.", Method: http.MethodPost, Path: pp + "/views",
		Params: []Param{pProject, body("name", "string", "View name").req(), body("description", "string", "Description"),
			body("visibility", "string", "Default private").oneOf("private", "shared"),
			body("filter_tree", "object", filterHelp),
			body("group_by", "string", "Grouping").oneOf("state", "state_type", "priority", "assignee", "label", "due_bucket", "start_bucket", "created_bucket", "updated_bucket"),
			body("sort_by", "string", "Sort field").oneOf("created_at", "updated_at", "priority", "due_date", "start_date", "title"),
			body("sort_dir", "string", "Sort direction").oneOf("asc", "desc"),
			body("default_tab", "string", "Default tab").oneOf("tasks", "board")}},
	{Name: "update_view", Title: "Update view", Description: "Update a saved view (owner, or project admin for shared views).", Method: http.MethodPatch, Path: pp + "/views/{slug}",
		Params: []Param{pProject, pathStr("slug", "View slug"), body("name", "string", "Name"), body("description", "string", "Description"),
			body("visibility", "string", "Owner only").oneOf("private", "shared"), body("filter_tree", "object", filterHelp),
			body("group_by", "string", "Grouping").oneOf("state", "state_type", "priority", "assignee", "label", "due_bucket", "start_bucket", "created_bucket", "updated_bucket"),
			body("sort_by", "string", "Sort field").oneOf("created_at", "updated_at", "priority", "due_date", "start_date", "title"),
			body("sort_dir", "string", "Sort direction").oneOf("asc", "desc"), body("default_tab", "string", "Default tab").oneOf("tasks", "board"),
			body("position", "integer", "Sort position")}},
	{Name: "delete_view", Title: "Delete view", Description: "Delete a saved view.", Method: http.MethodDelete, Path: pp + "/views/{slug}", Confirm: true,
		Params: []Param{pProject, pathStr("slug", "View slug")}},

	// Cycles
	{Name: "list_cycles", Title: "List cycles", Description: "Cycles (sprints) of a project with progress.", Method: http.MethodGet, Path: pp + "/cycles", Params: []Param{pProject, pPage, pPerPage}},
	{Name: "list_all_cycles", Title: "List all cycles", Description: "All cycles of a project, unpaginated (id, title, dates, status).", Method: http.MethodGet, Path: pp + "/cycles/all", Params: []Param{pProject}},
	{Name: "get_cycle", Title: "Get cycle", Description: "Get a cycle.", Method: http.MethodGet, Path: pp + "/cycles/{cycle_id}", Params: []Param{pProject, pCycle}},
	{Name: "list_cycle_tasks", Title: "List cycle tasks", Description: "Tasks in a cycle.", Method: http.MethodGet, Path: pp + "/cycles/{cycle_id}/tasks",
		Params: []Param{pProject, pCycle, query("assignee", "string", "Filter by assignee user UUID")}},
	{Name: "get_cycle_metrics", Title: "Get cycle metrics", Description: "Progress breakdown of a cycle.", Method: http.MethodGet, Path: pp + "/cycles/{cycle_id}/metrics", Params: []Param{pProject, pCycle}},
	{Name: "list_tasks_without_cycle", Title: "List tasks without cycle", Description: "Project tasks not in any cycle.", Method: http.MethodGet, Path: pp + "/cycles/unassigned-tasks",
		Params: []Param{pProject, pLimit, pSearch}},
	{Name: "create_cycle", Title: "Create cycle", Description: "Create a cycle; dates must not overlap other cycles (project admin).", Method: http.MethodPost, Path: pp + "/cycles",
		Params: []Param{pProject, body("title", "string", "Title").req(), body("description", "string", "Description"),
			body("start_date", "string", "YYYY-MM-DD").req(), body("end_date", "string", "YYYY-MM-DD").req()}},
	{Name: "update_cycle", Title: "Update cycle", Description: "Update a cycle (project admin).", Method: http.MethodPatch, Path: pp + "/cycles/{cycle_id}",
		Params: []Param{pProject, pCycle, body("title", "string", "Title"), body("description", "string", "Description"),
			body("start_date", "string", "YYYY-MM-DD"), body("end_date", "string", "YYYY-MM-DD")}},
	{Name: "delete_cycle", Title: "Delete cycle", Description: "Delete a cycle (project admin).", Method: http.MethodDelete, Path: pp + "/cycles/{cycle_id}", Confirm: true, Params: []Param{pProject, pCycle}},
	{Name: "add_cycle_tasks", Title: "Add tasks to cycle", Description: "Add tasks to a cycle; a task can be in only one cycle (project admin).", Method: http.MethodPost, Path: pp + "/cycles/{cycle_id}/tasks",
		Params: []Param{pProject, pCycle, pTaskUUIDs}},
	{Name: "remove_cycle_task", Title: "Remove task from cycle", Description: "Remove a task from a cycle (project admin).", Method: http.MethodDelete, Path: pp + "/cycles/{cycle_id}/tasks/{task_id}",
		Params: []Param{pProject, pCycle, pathStr("task_id", "Task UUID")}},

	// Modules
	{Name: "list_modules", Title: "List modules", Description: "Modules of a project.", Method: http.MethodGet, Path: pp + "/modules",
		Params: []Param{pProject, pPage, pPerPage,
			query("status", "string", "Status filter").oneOf("backlog", "planned", "in_progress", "paused", "completed", "cancelled"),
			query("lead_id", "string", "Lead user UUID"),
			query("sort_by", "string", "Sort field").oneOf("created_at", "end_date", "progress"),
			query("sort_dir", "string", "Sort direction").oneOf("asc", "desc")}},
	{Name: "get_module", Title: "Get module", Description: "Get a module.", Method: http.MethodGet, Path: pp + "/modules/{module_id}", Params: []Param{pProject, pModule}},
	{Name: "list_module_tasks", Title: "List module tasks", Description: "Tasks in a module.", Method: http.MethodGet, Path: pp + "/modules/{module_id}/tasks",
		Params: []Param{pProject, pModule, query("assignee", "string", "Filter by assignee user UUID")}},
	{Name: "list_module_members", Title: "List module members", Description: "Members of a module.", Method: http.MethodGet, Path: pp + "/modules/{module_id}/members", Params: []Param{pProject, pModule}},
	{Name: "get_module_metrics", Title: "Get module metrics", Description: "Progress breakdown of a module.", Method: http.MethodGet, Path: pp + "/modules/{module_id}/metrics", Params: []Param{pProject, pModule}},
	{Name: "list_tasks_not_in_module", Title: "List tasks not in module", Description: "Project tasks not in the given module.", Method: http.MethodGet, Path: pp + "/modules/tasks-picker",
		Params: []Param{pProject, query("module_id", "string", "Module UUID").req(), pLimit, pSearch}},
	{Name: "create_module", Title: "Create module", Description: "Create a module (project admin).", Method: http.MethodPost, Path: pp + "/modules",
		Params: []Param{pProject, body("title", "string", "Title").req(), body("description", "string", "Description"),
			body("status", "string", "Default backlog").oneOf("backlog", "planned", "in_progress", "paused", "completed", "cancelled"),
			body("start_date", "string", "YYYY-MM-DD"), body("end_date", "string", "YYYY-MM-DD"),
			body("lead_id", "string", "Lead user UUID (project member)"), body("member_ids", "string[]", "Member user UUIDs (project members)")}},
	{Name: "update_module", Title: "Update module", Description: "Update a module (project admin). Use clear_* flags to unset dates or lead.", Method: http.MethodPatch, Path: pp + "/modules/{module_id}",
		Params: []Param{pProject, pModule, body("title", "string", "Title"), body("description", "string", "Description"),
			body("status", "string", "Status").oneOf("backlog", "planned", "in_progress", "paused", "completed", "cancelled"),
			body("start_date", "string", "YYYY-MM-DD"), body("end_date", "string", "YYYY-MM-DD"), body("lead_id", "string", "Lead user UUID"),
			body("clear_start_date", "boolean", "Unset start date"), body("clear_end_date", "boolean", "Unset end date"), body("clear_lead", "boolean", "Unset lead")}},
	{Name: "delete_module", Title: "Delete module", Description: "Delete a module (project admin).", Method: http.MethodDelete, Path: pp + "/modules/{module_id}", Confirm: true, Params: []Param{pProject, pModule}},
	{Name: "duplicate_module", Title: "Duplicate module", Description: "Copy a module (description, lead, members) with optional tasks (project admin).", Method: http.MethodPost, Path: pp + "/modules/{module_id}/duplicate",
		Params: []Param{pProject, pModule, body("title", "string", "New title").req(), body("start_date", "string", "YYYY-MM-DD"), body("end_date", "string", "YYYY-MM-DD"),
			body("task_ids", "string[]", "Task UUIDs to link to the copy")}},
	{Name: "add_module_tasks", Title: "Add tasks to module", Description: "Add tasks to a module (project admin).", Method: http.MethodPost, Path: pp + "/modules/{module_id}/tasks",
		Params: []Param{pProject, pModule, pTaskUUIDs}},
	{Name: "remove_module_task", Title: "Remove task from module", Description: "Remove a task from a module (project admin).", Method: http.MethodDelete, Path: pp + "/modules/{module_id}/tasks/{task_id}",
		Params: []Param{pProject, pModule, pathStr("task_id", "Task UUID")}},
	{Name: "add_module_member", Title: "Add module member", Description: "Add a project member to a module (project admin).", Method: http.MethodPost, Path: pp + "/modules/{module_id}/members",
		Params: []Param{pProject, pModule, body("user_id", "string", "User UUID").req()}},
	{Name: "remove_module_member", Title: "Remove module member", Description: "Remove a member from a module (project admin).", Method: http.MethodDelete, Path: pp + "/modules/{module_id}/members/{user_id}",
		Params: []Param{pProject, pModule, pUser}},

	// Tasks
	{Name: "list_tasks", Title: "List tasks", Description: "List top-level tasks of a project (subtasks excluded) with optional filters.", Method: http.MethodGet, Path: pp + "/tasks",
		Params: []Param{pProject, pPage, pPerPage,
			query("sort_by", "string", "Sort field").oneOf("created_at", "updated_at", "priority", "due_date", "start_date", "title"),
			query("sort_dir", "string", "Sort direction").oneOf("asc", "desc"),
			{Name: "filter", Type: "object", In: InQuery, Description: filterHelp},
			query("view", "string", "Saved view slug to apply when no filter is given")}},
	{Name: "get_task", Title: "Get task", Description: "Get full task details: state, assignees, labels, parent, cycle, modules, followers.", Method: http.MethodGet, Path: ppt, Params: []Param{pProject, pTask}},
	{Name: "create_task", Title: "Create task", Description: "Create a task, or a subtask when parent_task_number is set.", Method: http.MethodPost, Path: pp + "/tasks",
		Params: []Param{pProject, body("title", "string", "Title").req(),
			body("description", "string", "Markdown. Mention users as [@Display Name](/profile/<user-uuid>)"),
			body("state_id", "string", "State UUID; default state if omitted"), pPriority,
			body("start_date", "string", "RFC3339, e.g. 2026-10-16T00:00:00Z"), body("due_date", "string", "RFC3339, not before start_date"),
			body("assignees", "string[]", "Assignee user UUIDs"), body("labels", "string[]", "Label UUIDs"),
			body("parent_task_number", "integer", "Parent task number (must be top-level) to create a subtask")}},
	{Name: "update_task", Title: "Update task", Description: "Update task fields. Assignees, labels, parent, cycle and modules have their own tools.", Method: http.MethodPatch, Path: ppt,
		Params: []Param{pProject, pTask, body("title", "string", "Title"), body("description", "string", "Markdown (replaces existing)"),
			body("state_id", "string", "State UUID; the project workflow may reject the transition"), pPriority,
			body("start_date", "string|null", "RFC3339 to set, null to clear"), body("due_date", "string|null", "RFC3339 to set, null to clear")}},
	{Name: "delete_task", Title: "Delete task", Description: "Delete a task and its subtasks (creator or project admin).", Method: http.MethodDelete, Path: ppt, Confirm: true, Params: []Param{pProject, pTask}},
	{Name: "delete_tasks", Title: "Delete tasks", Description: "Bulk delete tasks by number (project admin).", Method: http.MethodPost, Path: pp + "/tasks/delete", Confirm: true,
		Params: []Param{pProject, body("task_numbers", "integer[]", "Task numbers").req()}},
	{Name: "move_task", Title: "Move task", Description: "Move a top-level task (with subtasks) to another project; states/labels remapped by name, cycle/module links dropped.", Method: http.MethodPost, Path: ppt + "/move", Confirm: true,
		Params: []Param{pProject, pTask, body("target_project_key", "string", "Destination project key").req()}},
	{Name: "move_tasks", Title: "Move tasks", Description: "Bulk move top-level tasks to another project.", Method: http.MethodPost, Path: pp + "/tasks/move", Confirm: true,
		Params: []Param{pProject, body("target_project_key", "string", "Destination project key").req(), body("task_numbers", "integer[]", "Task numbers").req()}},
	{Name: "add_assignee", Title: "Assign user", Description: "Assign a user to a task.", Method: http.MethodPost, Path: ppt + "/assignees",
		Params: []Param{pProject, pTask, body("user_id", "string", "User UUID").req()}},
	{Name: "remove_assignee", Title: "Unassign user", Description: "Remove an assignee from a task.", Method: http.MethodDelete, Path: ppt + "/assignees/{user_id}",
		Params: []Param{pProject, pTask, pUser}},
	{Name: "add_task_label", Title: "Add label to task", Description: "Add a label to a task.", Method: http.MethodPost, Path: ppt + "/labels",
		Params: []Param{pProject, pTask, body("label_id", "string", "Label UUID").req()}},
	{Name: "remove_task_label", Title: "Remove label from task", Description: "Remove a label from a task.", Method: http.MethodDelete, Path: ppt + "/labels/{label_id}",
		Params: []Param{pProject, pTask, pathStr("label_id", "Label UUID")}},
	{Name: "follow_task", Title: "Follow task", Description: "Follow a task to get its notifications.", Method: http.MethodPost, Path: ppt + "/follow", Params: []Param{pProject, pTask}},
	{Name: "unfollow_task", Title: "Unfollow task", Description: "Stop following a task.", Method: http.MethodDelete, Path: ppt + "/follow", Params: []Param{pProject, pTask}},

	// Subtasks and blockers
	{Name: "list_subtasks", Title: "List subtasks", Description: "Subtasks of a task.", Method: http.MethodGet, Path: ppt + "/subtasks", Params: []Param{pProject, pTask}},
	{Name: "list_subtask_candidates", Title: "List subtask candidates", Description: "Tasks that can be attached as subtasks.", Method: http.MethodGet, Path: ppt + "/subtasks/candidates",
		Params: []Param{pProject, pTask, pLimit, pSearch}},
	{Name: "attach_subtasks", Title: "Attach subtasks", Description: "Make tasks subtasks of this top-level task (also re-parents).", Method: http.MethodPost, Path: ppt + "/subtasks",
		Params: []Param{pProject, pTask, pTaskUUIDs}},
	{Name: "detach_subtask", Title: "Detach subtask", Description: "Make a subtask top-level again.", Method: http.MethodDelete, Path: ppt + "/subtasks/{task_id}",
		Params: []Param{pProject, pTask, pathStr("task_id", "Subtask UUID")}},
	{Name: "list_blockers", Title: "List blockers", Description: "Tasks blocking this task and tasks it blocks.", Method: http.MethodGet, Path: ppt + "/blockers", Params: []Param{pProject, pTask}},
	{Name: "list_blocker_candidates", Title: "List blocker candidates", Description: "Tasks that can be linked as blockers.", Method: http.MethodGet, Path: ppt + "/blockers/candidates",
		Params: []Param{pProject, pTask, pLimit, pSearch}},
	{Name: "add_blockers", Title: "Add blockers", Description: "Link blocker relations within the same project; no cycles allowed.", Method: http.MethodPost, Path: ppt + "/blockers",
		Params: []Param{pProject, pTask, pTaskUUIDs,
			body("relation", "string", "blocked_by: listed tasks block this task; blocking: this task blocks listed tasks").req().oneOf("blocked_by", "blocking")}},
	{Name: "remove_blocker", Title: "Remove blocker", Description: "Remove a blocker link in either direction.", Method: http.MethodDelete, Path: ppt + "/blockers/{task_id}",
		Params: []Param{pProject, pTask, pathStr("task_id", "Other task UUID")}},

	// Attachments. upload_id must be a file the user attached in this chat (enforced in agent.go).
	{Name: "list_task_attachments", Title: "List task attachments", Description: "Files attached to a task.", Method: http.MethodGet, Path: ppt + "/attachments", Params: []Param{pProject, pTask}},
	{Name: "attach_file_to_task", Title: "Attach file to task", Description: "Attach a file the user uploaded in this chat to a task.", Method: http.MethodPost, Path: ppt + "/attachments",
		Params: []Param{pProject, pTask, pUpload}},
	{Name: "attach_file_to_comment", Title: "Attach file to comment", Description: "Attach a file the user uploaded in this chat to a comment.", Method: http.MethodPost, Path: ppt + "/comments/{comment_id}/attachments",
		Params: []Param{pProject, pTask, pathStr("comment_id", "Comment UUID"), pUpload}},

	// Comments and activity
	{Name: "list_comments", Title: "List comments", Description: "Comments on a task.", Method: http.MethodGet, Path: ppt + "/comments", Params: []Param{pProject, pTask}},
	{Name: "create_comment", Title: "Add comment", Description: "Comment on a task.", Method: http.MethodPost, Path: ppt + "/comments",
		Params: []Param{pProject, pTask, body("content", "string", "Markdown. Mention users as [@Display Name](/profile/<user-uuid>)").req()}},
	{Name: "update_comment", Title: "Edit comment", Description: "Edit the user's own comment.", Method: http.MethodPatch, Path: ppt + "/comments/{comment_id}",
		Params: []Param{pProject, pTask, pathStr("comment_id", "Comment UUID"), body("content", "string", "New markdown content").req()}},
	{Name: "delete_comment", Title: "Delete comment", Description: "Delete a comment (author or project admin).", Method: http.MethodDelete, Path: ppt + "/comments/{comment_id}", Confirm: true,
		Params: []Param{pProject, pTask, pathStr("comment_id", "Comment UUID")}},
	{Name: "get_task_activity", Title: "Get task activity", Description: "Change history of a task.", Method: http.MethodGet, Path: ppt + "/activity", Params: []Param{pProject, pTask}},
}

var toolIndex = func() map[string]*Tool {
	m := make(map[string]*Tool, len(catalog))
	for i := range catalog {
		m[catalog[i].Name] = &catalog[i]
	}
	return m
}()

func toolByName(name string) *Tool { return toolIndex[name] }

// LocalTool is a read-only tool a channel adapter adds to its runs, e.g. reading more chat
// history. The Meta it returns is stored with the result, like a user message's.
type LocalTool struct {
	Name        string
	Description string
	Params      []Param
	Run         func(ctx context.Context, args map[string]any) (string, Meta)
}

func findLocal(tools []LocalTool, name string) *LocalTool {
	for i := range tools {
		if tools[i].Name == name {
			return &tools[i]
		}
	}
	return nil
}

func toolSpecs(local []LocalTool) []ToolSpec {
	specs := make([]ToolSpec, 0, len(catalog)+len(local))
	for _, t := range catalog {
		desc := t.Description
		if t.Confirm {
			desc += " Requires user approval."
		}
		specs = append(specs, ToolSpec{Name: t.Name, Description: desc, Schema: schemaFor(t.Params)})
	}
	for _, t := range local {
		specs = append(specs, ToolSpec{Name: t.Name, Description: t.Description, Schema: schemaFor(t.Params)})
	}
	return specs
}

func schemaFor(params []Param) map[string]any {
	props := map[string]any{}
	required := []string{}
	for _, p := range params {
		prop := map[string]any{"description": p.Description}
		switch {
		case strings.HasSuffix(p.Type, "[]"):
			prop["type"] = "array"
			prop["items"] = map[string]any{"type": strings.TrimSuffix(p.Type, "[]")}
		case strings.Contains(p.Type, "|"):
			prop["type"] = strings.Split(p.Type, "|")
		default:
			prop["type"] = p.Type
		}
		if len(p.Enum) > 0 {
			prop["enum"] = p.Enum
		}
		props[p.Name] = prop
		if p.Required {
			required = append(required, p.Name)
		}
	}
	return map[string]any{"type": "object", "properties": props, "required": required}
}
