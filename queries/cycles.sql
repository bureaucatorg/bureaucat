-- ==================== CYCLES ====================

-- name: CreateCycle :one
INSERT INTO cycles (project_id, title, description, start_date, end_date, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, project_id, title, description, start_date, end_date, created_by, created_at, updated_at, deleted_at;

-- name: GetCycleByID :one
SELECT c.id, c.project_id, c.title, c.description, c.start_date, c.end_date,
       c.created_by, c.created_at, c.updated_at, c.deleted_at,
       p.project_key, p.name AS project_name
FROM cycles c
JOIN projects p ON c.project_id = p.id
WHERE c.id = $1 AND c.deleted_at IS NULL;

-- name: UpdateCycle :one
UPDATE cycles
SET title       = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    start_date  = COALESCE(sqlc.narg('start_date'), start_date),
    end_date    = COALESCE(sqlc.narg('end_date'), end_date),
    updated_at  = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, project_id, title, description, start_date, end_date, created_by, created_at, updated_at, deleted_at;

-- name: SoftDeleteCycle :exec
UPDATE cycles
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListProjectCycles :many
SELECT c.id, c.project_id, c.title, c.description, c.start_date, c.end_date,
       c.created_by, c.created_at, c.updated_at,
       COALESCE(stats.total_tasks, 0)::int     AS total_tasks,
       COALESCE(stats.completed_tasks, 0)::int AS completed_tasks
FROM cycles c
LEFT JOIN LATERAL (
    SELECT COUNT(*)                                                          AS total_tasks,
           COUNT(*) FILTER (WHERE ps.state_type = 'completed')              AS completed_tasks
    FROM cycle_tasks ct
    JOIN tasks t ON ct.task_id = t.id AND t.deleted_at IS NULL
    JOIN project_states ps ON t.state_id = ps.id
    WHERE ct.cycle_id = c.id
) stats ON TRUE
WHERE c.project_id = $1 AND c.deleted_at IS NULL
ORDER BY c.start_date DESC, c.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountProjectCycles :one
SELECT COUNT(*)
FROM cycles
WHERE project_id = $1 AND deleted_at IS NULL;

-- name: ListProjectCyclesAll :many
SELECT c.id, c.title, c.start_date, c.end_date
FROM cycles c
WHERE c.project_id = $1 AND c.deleted_at IS NULL
ORDER BY c.start_date DESC;

-- name: ListActiveCyclesForUser :many
SELECT c.id, c.project_id, c.title, c.description, c.start_date, c.end_date,
       c.created_by, c.created_at, c.updated_at,
       p.project_key, p.name AS project_name,
       COALESCE(stats.total_tasks, 0)::int     AS total_tasks,
       COALESCE(stats.completed_tasks, 0)::int AS completed_tasks
FROM cycles c
JOIN projects p ON c.project_id = p.id AND p.deleted_at IS NULL
JOIN project_members pm ON p.id = pm.project_id AND pm.user_id = $1
LEFT JOIN LATERAL (
    SELECT COUNT(*)                                                          AS total_tasks,
           COUNT(*) FILTER (WHERE ps.state_type = 'completed')              AS completed_tasks
    FROM cycle_tasks ct
    JOIN tasks t ON ct.task_id = t.id AND t.deleted_at IS NULL
    JOIN project_states ps ON t.state_id = ps.id
    WHERE ct.cycle_id = c.id
) stats ON TRUE
WHERE c.deleted_at IS NULL
  AND CURRENT_DATE BETWEEN c.start_date AND c.end_date
ORDER BY c.end_date ASC, p.name ASC;

-- name: CheckCycleOverlap :one
SELECT COUNT(*)::int AS overlap_count
FROM cycles
WHERE project_id = $1
  AND deleted_at IS NULL
  AND NOT (end_date < sqlc.arg('start_date')::date OR start_date > sqlc.arg('end_date')::date)
  AND (sqlc.narg('exclude_id')::uuid IS NULL OR id <> sqlc.narg('exclude_id')::uuid);

-- ==================== CYCLE TASKS ====================

-- name: AddTasksToCycle :exec
INSERT INTO cycle_tasks (cycle_id, task_id, added_by)
SELECT sqlc.arg('cycle_id')::uuid, tid, sqlc.arg('added_by')::uuid
FROM unnest(sqlc.arg('task_ids')::uuid[]) AS tid
ON CONFLICT DO NOTHING;

-- name: RemoveTaskFromCycle :exec
DELETE FROM cycle_tasks
WHERE cycle_id = $1 AND task_id = $2;

-- name: GetTaskCycleID :one
SELECT cycle_id FROM cycle_tasks WHERE task_id = $1;

-- name: ListCycleTasks :many
SELECT t.id, t.project_id, t.task_number, t.title, t.description, t.state_id, t.priority,
       t.start_date, t.due_date, t.created_by, t.created_at, t.updated_at,
       p.project_key,
       ps.name AS state_name, ps.state_type, ps.color AS state_color
FROM cycle_tasks ct
JOIN tasks t ON ct.task_id = t.id AND t.deleted_at IS NULL
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
WHERE ct.cycle_id = $1
  AND (sqlc.narg('assignee_id')::uuid IS NULL OR EXISTS (
      SELECT 1 FROM task_assignees ta
      WHERE ta.task_id = t.id AND ta.user_id = sqlc.narg('assignee_id')::uuid
  ))
ORDER BY ps.position ASC, t.created_at DESC;

-- name: ListUnassignedProjectTasks :many
SELECT t.id, t.project_id, t.task_number, t.title, t.state_id, t.priority,
       p.project_key, ps.name AS state_name, ps.state_type, ps.color AS state_color
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN project_states ps ON t.state_id = ps.id
WHERE t.project_id = $1 AND t.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM cycle_tasks ct WHERE ct.task_id = t.id)
  AND (sqlc.narg('search')::text IS NULL
       OR t.title ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY t.created_at DESC
LIMIT $2;

-- name: GetCycleMetrics :one
SELECT
    COUNT(*)::int                                                         AS total,
    COUNT(*) FILTER (WHERE ps.state_type = 'completed')::int             AS completed,
    COUNT(*) FILTER (WHERE ps.state_type = 'started')::int               AS in_progress,
    COUNT(*) FILTER (WHERE ps.state_type IN ('backlog', 'unstarted'))::int AS todo,
    COUNT(*) FILTER (WHERE ps.state_type = 'cancelled')::int             AS cancelled
FROM cycle_tasks ct
JOIN tasks t ON ct.task_id = t.id AND t.deleted_at IS NULL
JOIN project_states ps ON t.state_id = ps.id
WHERE ct.cycle_id = $1;

-- name: GetCycleStateBreakdown :many
SELECT ps.id AS state_id, ps.name AS state_name, ps.color AS state_color,
       ps.state_type, ps.position,
       COUNT(t.id)::int AS task_count
FROM cycle_tasks ct
JOIN tasks t ON ct.task_id = t.id AND t.deleted_at IS NULL
JOIN project_states ps ON t.state_id = ps.id
WHERE ct.cycle_id = $1
GROUP BY ps.id, ps.name, ps.color, ps.state_type, ps.position
ORDER BY ps.position ASC;

-- name: ListCycleAssignees :many
SELECT u.id AS user_id, u.username, u.first_name, u.last_name, u.avatar_url,
       COUNT(DISTINCT t.id)::int AS task_count
FROM cycle_tasks ct
JOIN tasks t ON ct.task_id = t.id AND t.deleted_at IS NULL
JOIN task_assignees ta ON t.id = ta.task_id
JOIN users u ON ta.user_id = u.id
WHERE ct.cycle_id = $1
GROUP BY u.id, u.username, u.first_name, u.last_name, u.avatar_url
ORDER BY task_count DESC, u.first_name ASC;
