-- ==================== MODULES ====================

-- name: CreateModule :one
-- The handler is responsible for defaulting `status` to 'backlog' when the
-- caller doesn't supply one, so the SQL can keep the arg non-nullable. This
-- sidesteps sqlc's handling of nullable enum args under a string override.
INSERT INTO modules (project_id, title, description, status, start_date, end_date, lead_id, created_by)
VALUES (
    sqlc.arg('project_id'),
    sqlc.arg('title'),
    sqlc.arg('description'),
    sqlc.arg('status')::module_status,
    sqlc.arg('start_date'),
    sqlc.arg('end_date'),
    sqlc.arg('lead_id'),
    sqlc.arg('created_by')
)
RETURNING id, project_id, title, description, status, start_date, end_date,
          lead_id, created_by, created_at, updated_at, deleted_at;

-- name: GetModuleByID :one
-- COALESCE the joined user fields because LEFT JOIN on nullable lead_id would
-- otherwise trip sqlc's (column-nullability-based) assumption that the fields
-- are non-null.
SELECT m.id, m.project_id, m.title, m.description, m.status,
       m.start_date, m.end_date, m.lead_id,
       m.created_by, m.created_at, m.updated_at, m.deleted_at,
       p.project_key, p.name AS project_name,
       COALESCE(stats.total_tasks, 0)::int     AS total_tasks,
       COALESCE(stats.completed_tasks, 0)::int AS completed_tasks,
       COALESCE(lu.username,   '')::text AS lead_username,
       COALESCE(lu.first_name, '')::text AS lead_first_name,
       COALESCE(lu.last_name,  '')::text AS lead_last_name,
       lu.avatar_url                     AS lead_avatar_url,
       COALESCE(lu.email,      '')::text AS lead_email
FROM modules m
JOIN projects p ON m.project_id = p.id
LEFT JOIN users lu ON m.lead_id = lu.id
LEFT JOIN LATERAL (
    SELECT COUNT(*)::int                                         AS total_tasks,
           COUNT(*) FILTER (WHERE ps.state_type = 'completed')::int AS completed_tasks
    FROM module_tasks mt
    JOIN tasks t ON mt.task_id = t.id AND t.deleted_at IS NULL
    JOIN project_states ps ON t.state_id = ps.id
    WHERE mt.module_id = m.id
) stats ON TRUE
WHERE m.id = $1 AND m.deleted_at IS NULL;

-- name: UpdateModule :one
-- `status` is passed as plain text; when empty string, no change. Avoids narg
-- around the enum type under the string override.
UPDATE modules
SET title       = COALESCE(sqlc.narg('title'),       title),
    description = COALESCE(sqlc.narg('description'), description),
    status      = CASE
                    WHEN sqlc.arg('status')::text = '' THEN status
                    ELSE sqlc.arg('status')::module_status
                  END,
    start_date  = CASE
                    WHEN sqlc.arg('clear_start_date')::bool THEN NULL
                    ELSE COALESCE(sqlc.narg('start_date'), start_date)
                  END,
    end_date    = CASE
                    WHEN sqlc.arg('clear_end_date')::bool THEN NULL
                    ELSE COALESCE(sqlc.narg('end_date'), end_date)
                  END,
    lead_id     = CASE
                    WHEN sqlc.arg('clear_lead')::bool THEN NULL
                    ELSE COALESCE(sqlc.narg('lead_id'), lead_id)
                  END,
    updated_at  = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, project_id, title, description, status, start_date, end_date,
          lead_id, created_by, created_at, updated_at, deleted_at;

-- name: SoftDeleteModule :exec
UPDATE modules
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListProjectModules :many
SELECT m.id, m.project_id, m.title, m.description, m.status,
       m.start_date, m.end_date, m.lead_id,
       m.created_by, m.created_at, m.updated_at,
       COALESCE(stats.total_tasks, 0)::int     AS total_tasks,
       COALESCE(stats.completed_tasks, 0)::int AS completed_tasks,
       COALESCE(lu.username,   '')::text AS lead_username,
       COALESCE(lu.first_name, '')::text AS lead_first_name,
       COALESCE(lu.last_name,  '')::text AS lead_last_name,
       lu.avatar_url                     AS lead_avatar_url,
       COALESCE(lu.email,      '')::text AS lead_email
FROM modules m
LEFT JOIN users lu ON m.lead_id = lu.id
LEFT JOIN LATERAL (
    SELECT COUNT(*)::int                                         AS total_tasks,
           COUNT(*) FILTER (WHERE ps.state_type = 'completed')::int AS completed_tasks
    FROM module_tasks mt
    JOIN tasks t ON mt.task_id = t.id AND t.deleted_at IS NULL
    JOIN project_states ps ON t.state_id = ps.id
    WHERE mt.module_id = m.id
) stats ON TRUE
WHERE m.project_id = $1 AND m.deleted_at IS NULL
  AND (sqlc.arg('status')::text     = ''     OR m.status = sqlc.arg('status')::module_status)
  AND (sqlc.narg('lead_id')::uuid   IS NULL  OR m.lead_id = sqlc.narg('lead_id')::uuid)
  AND (sqlc.narg('start_after')::date IS NULL OR m.start_date >= sqlc.narg('start_after')::date)
  AND (sqlc.narg('end_before')::date  IS NULL OR m.end_date   <= sqlc.narg('end_before')::date)
ORDER BY
    CASE WHEN sqlc.arg('sort_by')::text = 'end_date' AND sqlc.arg('sort_dir')::text = 'asc'
         THEN m.end_date END ASC NULLS LAST,
    CASE WHEN sqlc.arg('sort_by')::text = 'end_date' AND sqlc.arg('sort_dir')::text = 'desc'
         THEN m.end_date END DESC NULLS LAST,
    CASE WHEN sqlc.arg('sort_by')::text = 'progress' AND sqlc.arg('sort_dir')::text = 'asc'
         THEN CASE WHEN COALESCE(stats.total_tasks, 0) = 0 THEN 0
                   ELSE (COALESCE(stats.completed_tasks, 0)::float / stats.total_tasks::float)
              END END ASC NULLS LAST,
    CASE WHEN sqlc.arg('sort_by')::text = 'progress' AND sqlc.arg('sort_dir')::text = 'desc'
         THEN CASE WHEN COALESCE(stats.total_tasks, 0) = 0 THEN 0
                   ELSE (COALESCE(stats.completed_tasks, 0)::float / stats.total_tasks::float)
              END END DESC NULLS LAST,
    m.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveModulesForUser :many
-- "Active" modules across all projects the user is a member of. Uses the
-- explicit status enum (modules don't derive status from dates the way cycles
-- do). in_progress is the natural "currently being worked on" state.
SELECT m.id, m.project_id, m.title, m.description, m.status,
       m.start_date, m.end_date, m.lead_id,
       m.created_by, m.created_at, m.updated_at,
       p.project_key, p.name AS project_name,
       COALESCE(stats.total_tasks, 0)::int     AS total_tasks,
       COALESCE(stats.completed_tasks, 0)::int AS completed_tasks,
       COALESCE(lu.username,   '')::text AS lead_username,
       COALESCE(lu.first_name, '')::text AS lead_first_name,
       COALESCE(lu.last_name,  '')::text AS lead_last_name,
       lu.avatar_url                     AS lead_avatar_url,
       COALESCE(lu.email,      '')::text AS lead_email
FROM modules m
JOIN projects p ON m.project_id = p.id AND p.deleted_at IS NULL
JOIN project_members pm ON p.id = pm.project_id AND pm.user_id = $1
LEFT JOIN users lu ON m.lead_id = lu.id
LEFT JOIN LATERAL (
    SELECT COUNT(*)::int                                         AS total_tasks,
           COUNT(*) FILTER (WHERE ps.state_type = 'completed')::int AS completed_tasks
    FROM module_tasks mt
    JOIN tasks t ON mt.task_id = t.id AND t.deleted_at IS NULL
    JOIN project_states ps ON t.state_id = ps.id
    WHERE mt.module_id = m.id
) stats ON TRUE
WHERE m.deleted_at IS NULL AND m.status = 'in_progress'
ORDER BY COALESCE(m.end_date, '9999-12-31'::date) ASC, p.name ASC;

-- name: CountProjectModules :one
SELECT COUNT(*)
FROM modules m
WHERE m.project_id = $1 AND m.deleted_at IS NULL
  AND (sqlc.arg('status')::text     = ''     OR m.status = sqlc.arg('status')::module_status)
  AND (sqlc.narg('lead_id')::uuid   IS NULL  OR m.lead_id = sqlc.narg('lead_id')::uuid)
  AND (sqlc.narg('start_after')::date IS NULL OR m.start_date >= sqlc.narg('start_after')::date)
  AND (sqlc.narg('end_before')::date  IS NULL OR m.end_date   <= sqlc.narg('end_before')::date);

-- ==================== MODULE MEMBERS ====================

-- name: AddModuleMember :exec
INSERT INTO module_members (module_id, user_id, added_by)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: AddModuleMembersBulk :exec
INSERT INTO module_members (module_id, user_id, added_by)
SELECT sqlc.arg('module_id')::uuid, uid, sqlc.arg('added_by')::uuid
FROM unnest(sqlc.arg('user_ids')::uuid[]) AS uid
ON CONFLICT DO NOTHING;

-- name: RemoveModuleMember :exec
DELETE FROM module_members
WHERE module_id = $1 AND user_id = $2;

-- name: ListModuleMembers :many
SELECT mm.user_id, mm.added_at,
       u.username, u.email, u.first_name, u.last_name, u.avatar_url
FROM module_members mm
JOIN users u ON mm.user_id = u.id
WHERE mm.module_id = $1
ORDER BY mm.added_at ASC;

-- name: ListModuleMembersForModules :many
-- Used for hydrating the list view with member avatars. Returns up to 4 members
-- per module so the card can show 3 + "+N more".
SELECT mm.module_id, mm.user_id,
       u.username, u.first_name, u.last_name, u.avatar_url
FROM module_members mm
JOIN users u ON mm.user_id = u.id
WHERE mm.module_id = ANY(sqlc.arg('module_ids')::uuid[])
ORDER BY mm.module_id, mm.added_at ASC;

-- ==================== MODULE TASKS ====================

-- name: AddModuleTasksBulk :exec
INSERT INTO module_tasks (module_id, task_id, added_by)
SELECT sqlc.arg('module_id')::uuid, tid, sqlc.arg('added_by')::uuid
FROM unnest(sqlc.arg('task_ids')::uuid[]) AS tid
ON CONFLICT DO NOTHING;

-- name: RemoveModuleTask :exec
DELETE FROM module_tasks
WHERE module_id = $1 AND task_id = $2;

-- name: ListModuleTasks :many
SELECT t.id, t.project_id, t.task_number, t.title, t.description, t.state_id, t.priority,
       t.start_date, t.due_date, t.created_by, t.created_at, t.updated_at,
       p.project_key,
       ps.name AS state_name, ps.state_type, ps.color AS state_color
FROM module_tasks mt
JOIN tasks t ON mt.task_id = t.id AND t.deleted_at IS NULL
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
WHERE mt.module_id = $1
  AND (sqlc.narg('assignee_id')::uuid IS NULL OR EXISTS (
      SELECT 1 FROM task_assignees ta
      WHERE ta.task_id = t.id AND ta.user_id = sqlc.narg('assignee_id')::uuid
  ))
ORDER BY ps.position ASC, t.created_at DESC;

-- name: ListModuleTaskIDs :many
-- Used by the duplicate dialog to pre-populate the issues checklist.
SELECT task_id FROM module_tasks WHERE module_id = $1;

-- name: ListProjectTasksNotInModule :many
-- Picker source: project tasks that are NOT already in the given module. A task
-- can belong to many modules, so we only exclude by the target module.
SELECT t.id, t.project_id, t.task_number, t.title, t.state_id, t.priority,
       p.project_key, ps.name AS state_name, ps.state_type, ps.color AS state_color
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
WHERE t.project_id = $1 AND t.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM module_tasks mt
      WHERE mt.task_id = t.id AND mt.module_id = sqlc.arg('module_id')::uuid
  )
  AND (sqlc.narg('search')::text IS NULL
       OR t.title ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY t.created_at DESC
LIMIT $2;

-- name: GetModuleMetrics :one
SELECT
    COUNT(*)::int                                                         AS total,
    COUNT(*) FILTER (WHERE ps.state_type = 'completed')::int             AS completed,
    COUNT(*) FILTER (WHERE ps.state_type = 'started')::int               AS in_progress,
    COUNT(*) FILTER (WHERE ps.state_type IN ('backlog', 'unstarted'))::int AS todo,
    COUNT(*) FILTER (WHERE ps.state_type = 'cancelled')::int             AS cancelled
FROM module_tasks mt
JOIN tasks t ON mt.task_id = t.id AND t.deleted_at IS NULL
JOIN project_states ps ON t.state_id = ps.id
WHERE mt.module_id = $1;

-- name: GetModuleStateBreakdown :many
SELECT ps.id AS state_id, ps.name AS state_name, ps.color AS state_color,
       ps.state_type, ps.position,
       COUNT(t.id)::int AS task_count
FROM module_tasks mt
JOIN tasks t ON mt.task_id = t.id AND t.deleted_at IS NULL
JOIN project_states ps ON t.state_id = ps.id
WHERE mt.module_id = $1
GROUP BY ps.id, ps.name, ps.color, ps.state_type, ps.position
ORDER BY ps.position ASC;

-- name: GetTaskAssigneesForSeeding :many
-- Given a set of task IDs, return the distinct user IDs of their assignees.
-- Used to auto-seed module members when tasks are linked to a module.
SELECT DISTINCT ta.user_id
FROM task_assignees ta
WHERE ta.task_id = ANY(sqlc.arg('task_ids')::uuid[]);
