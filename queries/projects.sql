-- ==================== PROJECTS ====================

-- name: CreateProject :one
INSERT INTO projects (project_key, name, description, icon_id, cover_id, created_by, workspace_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, project_key, name, description, icon_id, cover_id, created_by, created_at, updated_at, deleted_at, disabled, workspace_id;

-- name: GetProjectByID :one
SELECT id, project_key, name, description, icon_id, cover_id, created_by, created_at, updated_at, deleted_at, disabled, workspace_id
FROM projects
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetProjectByKey :one
SELECT id, project_key, name, description, icon_id, cover_id, created_by, created_at, updated_at, deleted_at, disabled, workspace_id
FROM projects
WHERE project_key = $1 AND deleted_at IS NULL;

-- name: SetProjectDisabled :one
UPDATE projects
SET disabled = sqlc.arg('disabled'), updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING id, project_key, name, description, icon_id, cover_id, created_by, created_at, updated_at, deleted_at, disabled, workspace_id;

-- name: UpdateProject :one
UPDATE projects
SET name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    icon_id = COALESCE(sqlc.narg('icon_id'), icon_id),
    cover_id = COALESCE(sqlc.narg('cover_id'), cover_id),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, project_key, name, description, icon_id, cover_id, created_by, created_at, updated_at, deleted_at, disabled, workspace_id;

-- name: UpdateProjectWorkspace :one
UPDATE projects
SET workspace_id = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, project_key, name, description, icon_id, cover_id, created_by, created_at, updated_at, deleted_at, disabled, workspace_id;

-- name: ListProjectMembersMissingFromWorkspace :many
-- Project members who are NOT members of the given workspace. Used to preview
-- who would lose visibility of the project when it moves to that workspace.
SELECT u.id, u.username, u.email, u.first_name, u.last_name, u.avatar_url
FROM project_members pm
JOIN users u ON pm.user_id = u.id
WHERE pm.project_id = $1
  AND NOT EXISTS (
    SELECT 1 FROM workspace_members wm
    WHERE wm.workspace_id = $2 AND wm.user_id = pm.user_id
  )
ORDER BY u.first_name ASC, u.last_name ASC;

-- name: AddProjectMembersToWorkspace :exec
-- Adds every member of the project as a member of the workspace, skipping any
-- who are already members. Used when moving a project to keep members' access.
INSERT INTO workspace_members (workspace_id, user_id)
SELECT $1, pm.user_id
FROM project_members pm
WHERE pm.project_id = $2
ON CONFLICT (workspace_id, user_id) DO NOTHING;

-- name: SoftDeleteProject :exec
UPDATE projects
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: RestoreProject :exec
UPDATE projects
SET deleted_at = NULL, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NOT NULL;

-- name: ListDeletedProjects :many
SELECT p.id, p.project_key, p.name, p.description, p.created_by, p.created_at, p.updated_at, p.deleted_at, p.workspace_id,
       w.name AS workspace_name,
       u.first_name AS creator_first_name, u.last_name AS creator_last_name, u.username AS creator_username
FROM projects p
JOIN workspaces w ON w.id = p.workspace_id
JOIN users u ON u.id = p.created_by
WHERE p.deleted_at IS NOT NULL
ORDER BY p.deleted_at DESC
LIMIT $1 OFFSET $2;

-- name: CountDeletedProjects :one
SELECT COUNT(*) FROM projects WHERE deleted_at IS NOT NULL;

-- name: ListUserProjects :many
SELECT p.id, p.project_key, p.name, p.description, p.icon_id, p.cover_id, p.created_by, p.created_at, p.updated_at, p.deleted_at, p.workspace_id, pm.role
FROM projects p
JOIN project_members pm ON p.id = pm.project_id
WHERE pm.user_id = $1 AND p.deleted_at IS NULL
ORDER BY p.name ASC
LIMIT $2 OFFSET $3;

-- name: CountUserProjects :one
SELECT COUNT(*)
FROM projects p
JOIN project_members pm ON p.id = pm.project_id
WHERE pm.user_id = $1 AND p.deleted_at IS NULL;

-- name: ListAllProjects :many
SELECT p.id, p.project_key, p.name, p.description, p.icon_id, p.cover_id, p.created_by, p.created_at, p.updated_at, p.deleted_at, p.workspace_id, 'admin' AS role
FROM projects p
WHERE p.deleted_at IS NULL
ORDER BY p.name ASC
LIMIT $1 OFFSET $2;

-- name: CountAllProjects :one
SELECT COUNT(*)
FROM projects p
WHERE p.deleted_at IS NULL;

-- name: ListUserProjectsFiltered :many
SELECT p.id, p.project_key, p.name, p.description, p.icon_id, p.cover_id, p.created_by, p.created_at, p.updated_at, p.deleted_at, p.workspace_id, pm.role
FROM projects p
JOIN project_members pm ON p.id = pm.project_id
WHERE pm.user_id = $1 AND p.deleted_at IS NULL
  AND (sqlc.narg('workspace_id')::uuid IS NULL OR p.workspace_id = sqlc.narg('workspace_id'))
  AND (sqlc.narg('search')::text IS NULL OR p.name ILIKE '%' || sqlc.narg('search') || '%' OR p.project_key ILIKE '%' || sqlc.narg('search') || '%' OR p.description ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY p.name ASC
LIMIT $2 OFFSET $3;

-- name: CountUserProjectsFiltered :one
SELECT COUNT(*)
FROM projects p
JOIN project_members pm ON p.id = pm.project_id
WHERE pm.user_id = $1 AND p.deleted_at IS NULL
  AND (sqlc.narg('workspace_id')::uuid IS NULL OR p.workspace_id = sqlc.narg('workspace_id'))
  AND (sqlc.narg('search')::text IS NULL OR p.name ILIKE '%' || sqlc.narg('search') || '%' OR p.project_key ILIKE '%' || sqlc.narg('search') || '%' OR p.description ILIKE '%' || sqlc.narg('search') || '%');

-- name: ListAllProjectsFiltered :many
SELECT p.id, p.project_key, p.name, p.description, p.icon_id, p.cover_id, p.created_by, p.created_at, p.updated_at, p.deleted_at, p.workspace_id, 'admin' AS role
FROM projects p
WHERE p.deleted_at IS NULL
  AND (sqlc.narg('workspace_id')::uuid IS NULL OR p.workspace_id = sqlc.narg('workspace_id'))
  AND (sqlc.narg('search')::text IS NULL OR p.name ILIKE '%' || sqlc.narg('search') || '%' OR p.project_key ILIKE '%' || sqlc.narg('search') || '%' OR p.description ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY p.name ASC
LIMIT $1 OFFSET $2;

-- name: CountAllProjectsFiltered :one
SELECT COUNT(*)
FROM projects p
WHERE p.deleted_at IS NULL
  AND (sqlc.narg('workspace_id')::uuid IS NULL OR p.workspace_id = sqlc.narg('workspace_id'))
  AND (sqlc.narg('search')::text IS NULL OR p.name ILIKE '%' || sqlc.narg('search') || '%' OR p.project_key ILIKE '%' || sqlc.narg('search') || '%' OR p.description ILIKE '%' || sqlc.narg('search') || '%');

-- name: ProjectKeyExists :one
SELECT EXISTS (
    SELECT 1 FROM projects
    WHERE project_key = $1 AND deleted_at IS NULL
) AS exists;

-- ==================== PROJECT MEMBERS ====================

-- name: AddProjectMember :one
INSERT INTO project_members (project_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING id, project_id, user_id, role, joined_at;

-- name: GetProjectMember :one
SELECT pm.id, pm.project_id, pm.user_id, pm.role, pm.joined_at,
       u.username, u.email, u.first_name, u.last_name, u.avatar_url
FROM project_members pm
JOIN users u ON pm.user_id = u.id
WHERE pm.project_id = $1 AND pm.user_id = $2;

-- name: GetProjectMemberRole :one
SELECT role FROM project_members
WHERE project_id = $1 AND user_id = $2;

-- name: UpdateProjectMemberRole :exec
UPDATE project_members
SET role = $3
WHERE project_id = $1 AND user_id = $2;

-- name: RemoveProjectMember :exec
DELETE FROM project_members
WHERE project_id = $1 AND user_id = $2;

-- name: ListProjectMembers :many
SELECT pm.id, pm.project_id, pm.user_id, pm.role, pm.joined_at,
       u.username, u.email, u.first_name, u.last_name, u.avatar_url
FROM project_members pm
JOIN users u ON pm.user_id = u.id
WHERE pm.project_id = $1
ORDER BY pm.joined_at ASC;

-- name: CountProjectMembers :one
SELECT COUNT(*) FROM project_members WHERE project_id = $1;

-- name: IsProjectMember :one
SELECT EXISTS (
    SELECT 1 FROM project_members
    WHERE project_id = $1 AND user_id = $2
) AS is_member;

-- ==================== PROJECT STATES ====================

-- name: CreateProjectState :one
INSERT INTO project_states (project_id, state_type, name, color, position, is_default)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, project_id, state_type, name, color, position, is_default, created_at;

-- name: GetProjectStateByID :one
SELECT id, project_id, state_type, name, color, position, is_default, created_at
FROM project_states
WHERE id = $1;

-- name: UpdateProjectState :one
UPDATE project_states
SET name = COALESCE(sqlc.narg('name'), name),
    color = COALESCE(sqlc.narg('color'), color),
    position = COALESCE(sqlc.narg('position'), position)
WHERE id = $1
RETURNING id, project_id, state_type, name, color, position, is_default, created_at;

-- name: DeleteProjectState :exec
DELETE FROM project_states WHERE id = $1;

-- name: SetDefaultProjectState :exec
-- Marks exactly one state as default for the project, clearing any previous
-- default. Atomic in a single statement (no transaction needed).
UPDATE project_states
SET is_default = (id = sqlc.arg('state_id'))
WHERE project_id = sqlc.arg('project_id');

-- name: ListProjectStates :many
SELECT id, project_id, state_type, name, color, position, is_default, created_at
FROM project_states
WHERE project_id = $1
ORDER BY position ASC, created_at ASC;

-- name: GetDefaultProjectState :one
SELECT id, project_id, state_type, name, color, position, is_default, created_at
FROM project_states
WHERE project_id = $1 AND is_default = true
LIMIT 1;

-- name: CountTasksInState :one
SELECT COUNT(*) FROM tasks WHERE state_id = $1 AND deleted_at IS NULL;

-- ==================== PROJECT LABELS ====================

-- name: CreateProjectLabel :one
INSERT INTO project_labels (project_id, name, color)
VALUES ($1, $2, $3)
RETURNING id, project_id, name, color, created_at;

-- name: GetProjectLabelByID :one
SELECT id, project_id, name, color, created_at
FROM project_labels
WHERE id = $1;

-- name: UpdateProjectLabel :one
UPDATE project_labels
SET name = COALESCE(sqlc.narg('name'), name),
    color = COALESCE(sqlc.narg('color'), color)
WHERE id = $1
RETURNING id, project_id, name, color, created_at;

-- name: DeleteProjectLabel :exec
DELETE FROM project_labels WHERE id = $1;

-- name: ListProjectLabels :many
SELECT id, project_id, name, color, created_at
FROM project_labels
WHERE project_id = $1
ORDER BY name ASC;

-- ==================== TASKS ====================

-- name: CreateTask :one
INSERT INTO tasks (project_id, task_number, title, description, state_id, priority, created_by, start_date, due_date, parent_task_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, sqlc.narg('parent_task_id'))
RETURNING id, project_id, task_number, title, description, state_id, priority, created_by, start_date, due_date, parent_task_id, created_at, updated_at, deleted_at;

-- name: GetNextTaskNumber :one
SELECT COALESCE(MAX(task_number), 0) + 1 AS next_number
FROM tasks
WHERE project_id = $1;

-- name: GetTaskByID :one
SELECT t.id, t.project_id, t.task_number, t.title, t.description, t.state_id, t.priority, t.created_by, t.start_date, t.due_date, t.parent_task_id, t.created_at, t.updated_at, t.deleted_at,
       p.project_key,
       ps.name as state_name, ps.state_type, ps.color as state_color,
       u.username as creator_username, u.first_name as creator_first_name, u.last_name as creator_last_name, u.avatar_url as creator_avatar_url,
       pt.task_number as parent_task_number, pt.title as parent_task_title,
       (SELECT COUNT(*) FROM tasks st WHERE st.parent_task_id = t.id AND st.deleted_at IS NULL)::bigint as subtask_count
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
JOIN users u ON t.created_by = u.id
LEFT JOIN tasks pt ON t.parent_task_id = pt.id AND pt.deleted_at IS NULL
WHERE t.id = $1 AND t.deleted_at IS NULL;

-- name: GetTaskByProjectAndNumber :one
SELECT t.id, t.project_id, t.task_number, t.title, t.description, t.state_id, t.priority, t.created_by, t.start_date, t.due_date, t.parent_task_id, t.created_at, t.updated_at, t.deleted_at,
       p.project_key,
       ps.name as state_name, ps.state_type, ps.color as state_color,
       u.username as creator_username, u.first_name as creator_first_name, u.last_name as creator_last_name, u.avatar_url as creator_avatar_url,
       pt.task_number as parent_task_number, pt.title as parent_task_title,
       (SELECT COUNT(*) FROM tasks st WHERE st.parent_task_id = t.id AND st.deleted_at IS NULL)::bigint as subtask_count
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
JOIN users u ON t.created_by = u.id
LEFT JOIN tasks pt ON t.parent_task_id = pt.id AND pt.deleted_at IS NULL
WHERE t.project_id = $1 AND t.task_number = $2 AND t.deleted_at IS NULL;

-- name: UpdateTask :one
UPDATE tasks
SET title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    state_id = COALESCE(sqlc.narg('state_id'), state_id),
    priority = COALESCE(sqlc.narg('priority'), priority),
    start_date = CASE WHEN sqlc.arg('update_start_date')::bool THEN sqlc.narg('start_date') ELSE start_date END,
    due_date = CASE WHEN sqlc.arg('update_due_date')::bool THEN sqlc.narg('due_date') ELSE due_date END,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, project_id, task_number, title, description, state_id, priority, created_by, start_date, due_date, parent_task_id, created_at, updated_at, deleted_at;

-- name: MoveTask :one
-- Move a task to a different project, assigning a new project-local task number
-- and state. Cycle/module links and labels are handled separately by the caller.
UPDATE tasks
SET project_id = sqlc.arg('project_id'),
    task_number = sqlc.arg('task_number'),
    state_id = sqlc.arg('state_id'),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING id, project_id, task_number, title, description, state_id, priority, created_by, start_date, due_date, parent_task_id, created_at, updated_at, deleted_at;

-- name: DeleteTaskCycleLinks :exec
DELETE FROM cycle_tasks WHERE task_id = $1;

-- name: DeleteTaskModuleLinks :exec
DELETE FROM module_tasks WHERE task_id = $1;

-- name: SoftDeleteTask :exec
UPDATE tasks
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListSubtasks :many
-- Direct children of a task, in task-number order. Used on the parent's detail page.
SELECT t.id, t.project_id, t.task_number, t.title, t.state_id, t.priority,
       t.created_by, u.first_name as creator_first_name, u.last_name as creator_last_name, u.avatar_url as creator_avatar_url,
       p.project_key, ps.name as state_name, ps.state_type, ps.color as state_color
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
JOIN users u ON t.created_by = u.id
WHERE t.parent_task_id = sqlc.arg('parent_id')::uuid AND t.deleted_at IS NULL
ORDER BY t.task_number ASC;

-- name: ListSubtaskIDsForMove :many
-- Child ids + state name of a task, used to cascade a cross-project move.
SELECT t.id, ps.name as state_name
FROM tasks t
JOIN project_states ps ON t.state_id = ps.id
WHERE t.parent_task_id = sqlc.arg('parent_id')::uuid AND t.deleted_at IS NULL
ORDER BY t.task_number ASC;

-- name: CascadeSoftDeleteSubtasks :exec
-- Soft-delete all children of a task (cascade-together on parent delete).
UPDATE tasks
SET deleted_at = NOW(), updated_at = NOW()
WHERE parent_task_id = sqlc.arg('parent_id')::uuid AND deleted_at IS NULL;

-- name: ListSubtaskCandidates :many
-- Picker source for "attach an existing task as a subtask". Returns project
-- tasks eligible to become a child of the given parent. Excludes: the parent
-- itself, tasks already parented to this same parent, and tasks that already
-- have their own children (attaching one would break the one-level rule).
-- Tasks parented elsewhere ARE included (they can be re-parented) and flagged
-- via has_parent so the UI can warn.
SELECT t.id, t.project_id, t.task_number, t.title, t.state_id, t.priority,
       p.project_key, ps.name as state_name, ps.state_type, ps.color as state_color,
       pt.task_number AS parent_task_number, pt.title AS parent_title
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
LEFT JOIN tasks pt ON t.parent_task_id = pt.id AND pt.deleted_at IS NULL
WHERE t.project_id = $1 AND t.deleted_at IS NULL
  AND t.id <> sqlc.arg('parent_id')::uuid
  AND (t.parent_task_id IS DISTINCT FROM sqlc.arg('parent_id')::uuid)
  AND NOT EXISTS (
      SELECT 1 FROM tasks c
      WHERE c.parent_task_id = t.id AND c.deleted_at IS NULL
  )
  AND (sqlc.narg('search')::text IS NULL
       OR t.title ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY t.created_at DESC
LIMIT $2;

-- name: GetTaskAttachEligibility :one
-- Validates a candidate before attaching it as a subtask: its project and
-- whether it already has children (which would break the one-level rule).
SELECT t.id, t.project_id,
       (SELECT COUNT(*) FROM tasks c WHERE c.parent_task_id = t.id AND c.deleted_at IS NULL)::int AS subtask_count
FROM tasks t
WHERE t.id = $1 AND t.deleted_at IS NULL;

-- name: SetTaskParent :exec
-- Sets (or clears) a task's parent. Used to attach/re-parent an existing task
-- as a subtask.
UPDATE tasks
SET parent_task_id = sqlc.narg('parent_task_id'), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListProjectTasks :many
SELECT t.id, t.project_id, t.task_number, t.title, t.description, t.state_id, t.priority, t.created_by, t.created_at, t.updated_at, t.deleted_at,
       p.project_key,
       ps.name as state_name, ps.state_type, ps.color as state_color,
       u.username as creator_username
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
JOIN users u ON t.created_by = u.id
WHERE t.project_id = $1 AND t.deleted_at IS NULL
ORDER BY t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountProjectTasks :one
SELECT COUNT(*)
FROM tasks
WHERE project_id = $1 AND deleted_at IS NULL;

-- Filtered list and count are now built dynamically by internal/store/tasks_filter.go
-- from a FilterTree. The projection here is documented for reference by that runner.

-- name: ListAssigneesForTasks :many
SELECT ta.task_id, ta.id, ta.user_id, ta.assigned_at,
       u.username, u.email, u.first_name, u.last_name, u.avatar_url
FROM task_assignees ta
JOIN users u ON ta.user_id = u.id
WHERE ta.task_id = ANY(@task_ids::uuid[])
ORDER BY ta.assigned_at ASC;

-- name: ListLabelsForTasks :many
SELECT tl.task_id, tl.label_id, tl.added_at,
       pl.name, pl.color
FROM task_labels tl
JOIN project_labels pl ON tl.label_id = pl.id
WHERE tl.task_id = ANY(@task_ids::uuid[])
ORDER BY pl.name ASC;

-- name: ListTasksByAssignee :many
SELECT t.id, t.project_id, t.task_number, t.title, t.state_id, t.priority,
       p.project_key, ps.name as state_name, ps.state_type, ps.color as state_color
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
JOIN task_assignees ta ON t.id = ta.task_id
WHERE ta.user_id = $1 AND t.deleted_at IS NULL AND p.deleted_at IS NULL
  AND ps.state_type NOT IN ('completed', 'cancelled')
  AND (sqlc.narg('workspace_id')::uuid IS NULL OR p.workspace_id = sqlc.narg('workspace_id'))
ORDER BY t.updated_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTasksByAssignee :one
SELECT COUNT(*)
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
JOIN task_assignees ta ON t.id = ta.task_id
WHERE ta.user_id = $1 AND t.deleted_at IS NULL AND p.deleted_at IS NULL
  AND ps.state_type NOT IN ('completed', 'cancelled')
  AND (sqlc.narg('workspace_id')::uuid IS NULL OR p.workspace_id = sqlc.narg('workspace_id'));

-- ==================== GLOBAL SEARCH ====================

-- name: SearchUserTasks :many
-- Matches tasks by title, description, or composed task key ("KEY-123") across
-- projects the user is a member of.
SELECT t.id, t.task_number, t.title, t.description,
       p.id AS project_id, p.project_key, p.name AS project_name,
       ps.name AS state_name, ps.state_type, ps.color AS state_color,
       t.updated_at
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
WHERE EXISTS (SELECT 1 FROM project_members pm WHERE pm.project_id = p.id AND pm.user_id = @user_id)
  AND t.deleted_at IS NULL AND p.deleted_at IS NULL
  AND (
    t.title ILIKE '%' || @query::text || '%'
    OR t.description ILIKE '%' || @query::text || '%'
    OR (p.project_key || '-' || t.task_number::text) ILIKE '%' || @query::text || '%'
  )
ORDER BY
  CASE WHEN (p.project_key || '-' || t.task_number::text) ILIKE @query::text || '%' THEN 0
       WHEN t.title ILIKE @query::text || '%' THEN 1
       ELSE 2 END,
  t.updated_at DESC
LIMIT @limit_count;

-- name: SearchAllTasks :many
-- Admin variant: matches across all projects.
SELECT t.id, t.task_number, t.title, t.description,
       p.id AS project_id, p.project_key, p.name AS project_name,
       ps.name AS state_name, ps.state_type, ps.color AS state_color,
       t.updated_at
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
WHERE t.deleted_at IS NULL AND p.deleted_at IS NULL
  AND (
    t.title ILIKE '%' || @query::text || '%'
    OR t.description ILIKE '%' || @query::text || '%'
    OR (p.project_key || '-' || t.task_number::text) ILIKE '%' || @query::text || '%'
  )
ORDER BY
  CASE WHEN (p.project_key || '-' || t.task_number::text) ILIKE @query::text || '%' THEN 0
       WHEN t.title ILIKE @query::text || '%' THEN 1
       ELSE 2 END,
  t.updated_at DESC
LIMIT @limit_count;

-- name: SearchUserProjects :many
-- Matches projects by key, name, or description across projects the user is a member of.
SELECT p.id, p.project_key, p.name, p.description, p.icon_id
FROM projects p
JOIN project_members pm ON pm.project_id = p.id
WHERE pm.user_id = @user_id AND p.deleted_at IS NULL
  AND (
    p.project_key ILIKE '%' || @query::text || '%'
    OR p.name ILIKE '%' || @query::text || '%'
    OR p.description ILIKE '%' || @query::text || '%'
  )
ORDER BY
  CASE WHEN p.project_key ILIKE @query::text || '%' THEN 0
       WHEN p.name ILIKE @query::text || '%' THEN 1
       ELSE 2 END,
  p.name ASC
LIMIT @limit_count;

-- name: SearchAllProjects :many
-- Admin variant: matches across all projects.
SELECT p.id, p.project_key, p.name, p.description, p.icon_id
FROM projects p
WHERE p.deleted_at IS NULL
  AND (
    p.project_key ILIKE '%' || @query::text || '%'
    OR p.name ILIKE '%' || @query::text || '%'
    OR p.description ILIKE '%' || @query::text || '%'
  )
ORDER BY
  CASE WHEN p.project_key ILIKE @query::text || '%' THEN 0
       WHEN p.name ILIKE @query::text || '%' THEN 1
       ELSE 2 END,
  p.name ASC
LIMIT @limit_count;

-- ==================== TASK ASSIGNEES ====================

-- name: AddTaskAssignee :one
INSERT INTO task_assignees (task_id, user_id, assigned_by)
VALUES ($1, $2, $3)
RETURNING id, task_id, user_id, assigned_at, assigned_by;

-- name: RemoveTaskAssignee :exec
DELETE FROM task_assignees
WHERE task_id = $1 AND user_id = $2;

-- name: ListTaskAssignees :many
SELECT ta.id, ta.task_id, ta.user_id, ta.assigned_at, ta.assigned_by,
       u.username, u.email, u.first_name, u.last_name, u.avatar_url
FROM task_assignees ta
JOIN users u ON ta.user_id = u.id
WHERE ta.task_id = $1
ORDER BY ta.assigned_at ASC;

-- name: IsTaskAssignee :one
SELECT EXISTS (
    SELECT 1 FROM task_assignees
    WHERE task_id = $1 AND user_id = $2
) AS is_assignee;

-- ==================== TASK LABELS ====================

-- name: AddTaskLabel :exec
INSERT INTO task_labels (task_id, label_id, added_by)
VALUES ($1, $2, $3);

-- name: RemoveTaskLabel :exec
DELETE FROM task_labels
WHERE task_id = $1 AND label_id = $2;

-- name: ListTaskLabels :many
SELECT tl.task_id, tl.label_id, tl.added_at, tl.added_by,
       pl.name, pl.color
FROM task_labels tl
JOIN project_labels pl ON tl.label_id = pl.id
WHERE tl.task_id = $1
ORDER BY pl.name ASC;

-- name: HasTaskLabel :one
SELECT EXISTS (
    SELECT 1 FROM task_labels
    WHERE task_id = $1 AND label_id = $2
) AS has_label;

-- ==================== COMMENTS ====================

-- name: CreateComment :one
INSERT INTO comments (task_id, content, created_by)
VALUES ($1, $2, $3)
RETURNING id, task_id, content, version, created_by, created_at, updated_at, deleted_at;

-- name: GetCommentByID :one
SELECT c.id, c.task_id, c.content, c.version, c.created_by, c.created_at, c.updated_at, c.deleted_at,
       u.username, u.first_name, u.last_name, u.avatar_url
FROM comments c
JOIN users u ON c.created_by = u.id
WHERE c.id = $1 AND c.deleted_at IS NULL;

-- name: UpdateComment :one
UPDATE comments
SET content = $2,
    version = version + 1,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, task_id, content, version, created_by, created_at, updated_at, deleted_at;

-- name: SoftDeleteComment :exec
UPDATE comments
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListTaskComments :many
SELECT c.id, c.task_id, c.content, c.version, c.created_by, c.created_at, c.updated_at, c.deleted_at,
       u.username, u.first_name, u.last_name, u.avatar_url
FROM comments c
JOIN users u ON c.created_by = u.id
WHERE c.task_id = $1 AND c.deleted_at IS NULL
ORDER BY c.created_at ASC;

-- name: CountTaskComments :one
SELECT COUNT(*) FROM comments WHERE task_id = $1 AND deleted_at IS NULL;

-- ==================== ACTIVITY LOG ====================

-- name: CreateActivityLog :one
INSERT INTO activity_log (task_id, activity_type, actor_id, field_name, old_value, new_value, checksum, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, task_id, activity_type, actor_id, field_name, old_value, new_value, created_at, checksum;

-- name: GetLastActivityChecksum :one
SELECT checksum
FROM activity_log
WHERE task_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: ListTaskActivity :many
SELECT al.id, al.task_id, al.activity_type, al.actor_id, al.field_name, al.old_value, al.new_value, al.created_at, al.checksum,
       u.username, u.first_name, u.last_name, u.avatar_url
FROM activity_log al
JOIN users u ON al.actor_id = u.id
WHERE al.task_id = $1
ORDER BY al.created_at ASC;

-- name: VerifyActivityChain :many
SELECT id, task_id, activity_type, actor_id, field_name, old_value, new_value, created_at, checksum
FROM activity_log
WHERE task_id = $1
ORDER BY created_at ASC;

-- ==================== TASK TEMPLATES ====================

-- name: CreateTaskTemplate :one
INSERT INTO task_templates (project_id, name, title, description, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTaskTemplateByID :one
SELECT * FROM task_templates WHERE id = $1;

-- name: ListTaskTemplates :many
SELECT * FROM task_templates WHERE project_id = $1 ORDER BY name ASC;

-- name: UpdateTaskTemplate :one
UPDATE task_templates
SET name = COALESCE(sqlc.narg('name'), name),
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTaskTemplate :exec
DELETE FROM task_templates WHERE id = $1;

-- ==================== USER ACTIVITY ====================

-- name: ListUserActivity :many
SELECT id, task_id, activity_type, actor_id, field_name, old_value, new_value, created_at,
       username, first_name, last_name,
       task_number, project_key, task_title
FROM (
  -- Activities from activity_log
  SELECT al.id, al.task_id, al.activity_type, al.actor_id, al.field_name, al.old_value, al.new_value, al.created_at,
         u.username, u.first_name, u.last_name,
         t.task_number, p.project_key, t.title as task_title
  FROM activity_log al
  JOIN users u ON al.actor_id = u.id
  JOIN tasks t ON al.task_id = t.id
  JOIN projects p ON t.project_id = p.id
  WHERE al.actor_id = $1

  UNION ALL

  -- Synthetic "task_created" entries for imported tasks with no activity_log
  SELECT t.id as id, t.id as task_id, 'task_created'::activity_type as activity_type, t.created_by as actor_id,
         NULL::varchar(100) as field_name, NULL::jsonb as old_value, NULL::jsonb as new_value, t.created_at,
         u.username, u.first_name, u.last_name,
         t.task_number, p.project_key, t.title as task_title
  FROM tasks t
  JOIN users u ON t.created_by = u.id
  JOIN projects p ON t.project_id = p.id
  WHERE t.created_by = $1
    AND t.deleted_at IS NULL
    AND NOT EXISTS (
      SELECT 1 FROM activity_log al
      WHERE al.task_id = t.id AND al.actor_id = $1 AND al.activity_type = 'task_created'
    )
) combined
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUserActivity :one
SELECT (
  (SELECT COUNT(*) FROM activity_log al1 WHERE al1.actor_id = $1)
  +
  (SELECT COUNT(*) FROM tasks t
   WHERE t.created_by = $1 AND t.deleted_at IS NULL
   AND NOT EXISTS (
     SELECT 1 FROM activity_log al2
     WHERE al2.task_id = t.id AND al2.actor_id = $1 AND al2.activity_type = 'task_created'
   ))
)::bigint;

-- name: ListUserActivityDates :many
SELECT activity_date, SUM(cnt)::int as activity_count
FROM (
  SELECT DATE(al.created_at AT TIME ZONE 'UTC') as activity_date, COUNT(*) as cnt
  FROM activity_log al
  WHERE al.actor_id = $1 AND al.created_at >= $2
  GROUP BY DATE(al.created_at AT TIME ZONE 'UTC')

  UNION ALL

  SELECT DATE(t.created_at AT TIME ZONE 'UTC') as activity_date, COUNT(*) as cnt
  FROM tasks t
  WHERE t.created_by = $1 AND t.deleted_at IS NULL AND t.created_at >= $2
    AND NOT EXISTS (
      SELECT 1 FROM activity_log al
      WHERE al.task_id = t.id AND al.actor_id = $1 AND al.activity_type = 'task_created'
    )
  GROUP BY DATE(t.created_at AT TIME ZONE 'UTC')
) combined
GROUP BY activity_date
ORDER BY activity_date ASC;

-- ==================== TASK PARTICIPANTS ====================

-- name: ListTaskParticipants :many
-- Everyone involved with a task: its creator, current assignees, and anyone who
-- has commented (non-deleted comments). Used to fan out notifications.
SELECT DISTINCT user_id FROM (
  SELECT t.created_by AS user_id
  FROM tasks t
  WHERE t.id = sqlc.arg('task_id') AND t.deleted_at IS NULL

  UNION

  SELECT ta.user_id
  FROM task_assignees ta
  WHERE ta.task_id = sqlc.arg('task_id')

  UNION

  SELECT c.created_by AS user_id
  FROM comments c
  WHERE c.task_id = sqlc.arg('task_id') AND c.deleted_at IS NULL
) participants;

-- ==================== IMPORT HELPERS ====================

-- name: GetProjectStateByProjectAndName :one
SELECT id, project_id, state_type, name, color, position, is_default, created_at
FROM project_states
WHERE project_id = $1 AND name = $2
LIMIT 1;

-- name: GetProjectLabelByProjectAndName :one
SELECT id, project_id, name, color, created_at
FROM project_labels
WHERE project_id = $1 AND name = $2
LIMIT 1;
