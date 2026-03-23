-- ==================== PROJECTS ====================

-- name: CreateProject :one
INSERT INTO projects (project_key, name, description, icon_id, cover_id, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, project_key, name, description, icon_id, cover_id, created_by, created_at, updated_at, deleted_at;

-- name: GetProjectByID :one
SELECT id, project_key, name, description, icon_id, cover_id, created_by, created_at, updated_at, deleted_at
FROM projects
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetProjectByKey :one
SELECT id, project_key, name, description, icon_id, cover_id, created_by, created_at, updated_at, deleted_at
FROM projects
WHERE project_key = $1 AND deleted_at IS NULL;

-- name: UpdateProject :one
UPDATE projects
SET name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    icon_id = COALESCE(sqlc.narg('icon_id'), icon_id),
    cover_id = COALESCE(sqlc.narg('cover_id'), cover_id),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, project_key, name, description, icon_id, cover_id, created_by, created_at, updated_at, deleted_at;

-- name: SoftDeleteProject :exec
UPDATE projects
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListUserProjects :many
SELECT p.id, p.project_key, p.name, p.description, p.icon_id, p.cover_id, p.created_by, p.created_at, p.updated_at, p.deleted_at, pm.role
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
SELECT p.id, p.project_key, p.name, p.description, p.icon_id, p.cover_id, p.created_by, p.created_at, p.updated_at, p.deleted_at, 'admin' AS role
FROM projects p
WHERE p.deleted_at IS NULL
ORDER BY p.name ASC
LIMIT $1 OFFSET $2;

-- name: CountAllProjects :one
SELECT COUNT(*)
FROM projects p
WHERE p.deleted_at IS NULL;

-- name: ListUserProjectsFiltered :many
SELECT p.id, p.project_key, p.name, p.description, p.icon_id, p.cover_id, p.created_by, p.created_at, p.updated_at, p.deleted_at, pm.role
FROM projects p
JOIN project_members pm ON p.id = pm.project_id
WHERE pm.user_id = $1 AND p.deleted_at IS NULL
  AND (sqlc.narg('search')::text IS NULL OR p.name ILIKE '%' || sqlc.narg('search') || '%' OR p.project_key ILIKE '%' || sqlc.narg('search') || '%' OR p.description ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY p.name ASC
LIMIT $2 OFFSET $3;

-- name: CountUserProjectsFiltered :one
SELECT COUNT(*)
FROM projects p
JOIN project_members pm ON p.id = pm.project_id
WHERE pm.user_id = $1 AND p.deleted_at IS NULL
  AND (sqlc.narg('search')::text IS NULL OR p.name ILIKE '%' || sqlc.narg('search') || '%' OR p.project_key ILIKE '%' || sqlc.narg('search') || '%' OR p.description ILIKE '%' || sqlc.narg('search') || '%');

-- name: ListAllProjectsFiltered :many
SELECT p.id, p.project_key, p.name, p.description, p.icon_id, p.cover_id, p.created_by, p.created_at, p.updated_at, p.deleted_at, 'admin' AS role
FROM projects p
WHERE p.deleted_at IS NULL
  AND (sqlc.narg('search')::text IS NULL OR p.name ILIKE '%' || sqlc.narg('search') || '%' OR p.project_key ILIKE '%' || sqlc.narg('search') || '%' OR p.description ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY p.name ASC
LIMIT $1 OFFSET $2;

-- name: CountAllProjectsFiltered :one
SELECT COUNT(*)
FROM projects p
WHERE p.deleted_at IS NULL
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
       u.username, u.email, u.first_name, u.last_name
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
       u.username, u.email, u.first_name, u.last_name
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
INSERT INTO tasks (project_id, task_number, title, description, state_id, priority, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, project_id, task_number, title, description, state_id, priority, created_by, created_at, updated_at, deleted_at;

-- name: GetNextTaskNumber :one
SELECT COALESCE(MAX(task_number), 0) + 1 AS next_number
FROM tasks
WHERE project_id = $1;

-- name: GetTaskByID :one
SELECT t.id, t.project_id, t.task_number, t.title, t.description, t.state_id, t.priority, t.created_by, t.created_at, t.updated_at, t.deleted_at,
       p.project_key,
       ps.name as state_name, ps.state_type, ps.color as state_color,
       u.username as creator_username, u.first_name as creator_first_name, u.last_name as creator_last_name
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
JOIN users u ON t.created_by = u.id
WHERE t.id = $1 AND t.deleted_at IS NULL;

-- name: GetTaskByProjectAndNumber :one
SELECT t.id, t.project_id, t.task_number, t.title, t.description, t.state_id, t.priority, t.created_by, t.created_at, t.updated_at, t.deleted_at,
       p.project_key,
       ps.name as state_name, ps.state_type, ps.color as state_color,
       u.username as creator_username, u.first_name as creator_first_name, u.last_name as creator_last_name
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
JOIN users u ON t.created_by = u.id
WHERE t.project_id = $1 AND t.task_number = $2 AND t.deleted_at IS NULL;

-- name: UpdateTask :one
UPDATE tasks
SET title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    state_id = COALESCE(sqlc.narg('state_id'), state_id),
    priority = COALESCE(sqlc.narg('priority'), priority),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, project_id, task_number, title, description, state_id, priority, created_by, created_at, updated_at, deleted_at;

-- name: SoftDeleteTask :exec
UPDATE tasks
SET deleted_at = NOW(), updated_at = NOW()
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

-- name: ListProjectTasksFiltered :many
SELECT t.id, t.project_id, t.task_number, t.title, t.description, t.state_id, t.priority, t.created_by, t.created_at, t.updated_at, t.deleted_at,
       p.project_key,
       ps.name as state_name, ps.state_type, ps.color as state_color,
       u.username as creator_username, u.first_name as creator_first_name, u.last_name as creator_last_name,
       (SELECT COUNT(*) FROM comments c WHERE c.task_id = t.id AND c.deleted_at IS NULL)::bigint as comment_count
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
JOIN users u ON t.created_by = u.id
WHERE t.project_id = $1
  AND t.deleted_at IS NULL
  AND (sqlc.narg('state_id')::uuid IS NULL OR t.state_id = sqlc.narg('state_id'))
  AND (sqlc.narg('state_type')::state_type IS NULL OR ps.state_type = sqlc.narg('state_type'))
  AND (sqlc.narg('priority')::int IS NULL OR t.priority = sqlc.narg('priority'))
  AND (sqlc.narg('created_by')::uuid IS NULL OR t.created_by = sqlc.narg('created_by'))
  AND (sqlc.narg('assigned_to')::uuid IS NULL OR EXISTS (SELECT 1 FROM task_assignees ta WHERE ta.task_id = t.id AND ta.user_id = sqlc.narg('assigned_to')))
  AND (sqlc.narg('search')::text IS NULL OR t.title ILIKE '%' || sqlc.narg('search') || '%' OR t.description ILIKE '%' || sqlc.narg('search') || '%')
  AND (sqlc.narg('from_date')::timestamptz IS NULL OR t.created_at >= sqlc.narg('from_date'))
  AND (sqlc.narg('to_date')::timestamptz IS NULL OR t.created_at <= sqlc.narg('to_date'))
ORDER BY t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountProjectTasks :one
SELECT COUNT(*)
FROM tasks
WHERE project_id = $1 AND deleted_at IS NULL;

-- name: CountProjectTasksFiltered :one
SELECT COUNT(*)
FROM tasks t
JOIN project_states ps ON t.state_id = ps.id
WHERE t.project_id = $1
  AND t.deleted_at IS NULL
  AND (sqlc.narg('state_id')::uuid IS NULL OR t.state_id = sqlc.narg('state_id'))
  AND (sqlc.narg('state_type')::state_type IS NULL OR ps.state_type = sqlc.narg('state_type'))
  AND (sqlc.narg('priority')::int IS NULL OR t.priority = sqlc.narg('priority'))
  AND (sqlc.narg('created_by')::uuid IS NULL OR t.created_by = sqlc.narg('created_by'))
  AND (sqlc.narg('assigned_to')::uuid IS NULL OR EXISTS (SELECT 1 FROM task_assignees ta WHERE ta.task_id = t.id AND ta.user_id = sqlc.narg('assigned_to')))
  AND (sqlc.narg('search')::text IS NULL OR t.title ILIKE '%' || sqlc.narg('search') || '%' OR t.description ILIKE '%' || sqlc.narg('search') || '%')
  AND (sqlc.narg('from_date')::timestamptz IS NULL OR t.created_at >= sqlc.narg('from_date'))
  AND (sqlc.narg('to_date')::timestamptz IS NULL OR t.created_at <= sqlc.narg('to_date'));

-- name: ListTasksByAssignee :many
SELECT t.id, t.project_id, t.task_number, t.title, t.state_id, t.priority,
       p.project_key, ps.name as state_name, ps.state_type, ps.color as state_color
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
JOIN task_assignees ta ON t.id = ta.task_id
WHERE ta.user_id = $1 AND t.deleted_at IS NULL AND p.deleted_at IS NULL
  AND ps.state_type NOT IN ('completed', 'cancelled')
ORDER BY t.updated_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTasksByAssignee :one
SELECT COUNT(*)
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
JOIN task_assignees ta ON t.id = ta.task_id
WHERE ta.user_id = $1 AND t.deleted_at IS NULL AND p.deleted_at IS NULL
  AND ps.state_type NOT IN ('completed', 'cancelled');

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
       u.username, u.email, u.first_name, u.last_name
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
       u.username, u.first_name, u.last_name
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
       u.username, u.first_name, u.last_name
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
       u.username, u.first_name, u.last_name
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
