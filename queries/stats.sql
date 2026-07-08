-- name: CountWorkspaces :one
SELECT COUNT(*) FROM workspaces WHERE deleted_at IS NULL;

-- name: CountProjects :one
SELECT COUNT(*) FROM projects WHERE deleted_at IS NULL;

-- name: CountTopLevelTasks :one
SELECT COUNT(*) FROM tasks WHERE deleted_at IS NULL AND parent_task_id IS NULL;

-- name: CountSubtasks :one
SELECT COUNT(*) FROM tasks WHERE deleted_at IS NULL AND parent_task_id IS NOT NULL;

-- name: CountPages :one
SELECT COUNT(*) FROM pages WHERE deleted_at IS NULL;

-- name: TasksByStateType :many
SELECT ps.state_type AS state_type, COUNT(t.id)::int AS count
FROM tasks t
JOIN project_states ps ON t.state_id = ps.id
WHERE t.deleted_at IS NULL AND t.parent_task_id IS NULL
GROUP BY ps.state_type
ORDER BY count DESC;

-- name: TasksByPriority :many
SELECT t.priority AS priority, COUNT(*)::int AS count
FROM tasks t
WHERE t.deleted_at IS NULL AND t.parent_task_id IS NULL
GROUP BY t.priority
ORDER BY t.priority ASC;

-- name: TopProjectsByTaskCount :many
SELECT p.id AS project_id, p.name, p.project_key, COUNT(t.id)::int AS task_count
FROM projects p
LEFT JOIN tasks t ON t.project_id = p.id AND t.deleted_at IS NULL
WHERE p.deleted_at IS NULL
GROUP BY p.id, p.name, p.project_key
ORDER BY task_count DESC
LIMIT 10;

-- name: ProjectsPerWorkspace :many
SELECT w.id AS workspace_id, w.name, w.workspace_key, COUNT(p.id)::int AS project_count
FROM workspaces w
LEFT JOIN projects p ON p.workspace_id = w.id AND p.deleted_at IS NULL
WHERE w.deleted_at IS NULL
GROUP BY w.id, w.name, w.workspace_key
ORDER BY project_count DESC;

-- name: TasksCreatedPerDay :many
SELECT d::date AS day, COUNT(t.id)::int AS count
FROM generate_series(
    sqlc.arg('from_date')::date,
    sqlc.arg('to_date')::date,
    INTERVAL '1 day'
) d
LEFT JOIN tasks t
    ON t.created_at::date = d::date
    AND t.deleted_at IS NULL
    AND t.parent_task_id IS NULL
GROUP BY d
ORDER BY d ASC;

-- name: SubtasksCreatedPerDay :many
SELECT d::date AS day, COUNT(t.id)::int AS count
FROM generate_series(
    sqlc.arg('from_date')::date,
    sqlc.arg('to_date')::date,
    INTERVAL '1 day'
) d
LEFT JOIN tasks t
    ON t.created_at::date = d::date
    AND t.deleted_at IS NULL
    AND t.parent_task_id IS NOT NULL
GROUP BY d
ORDER BY d ASC;

-- name: PagesCreatedPerDay :many
SELECT d::date AS day, COUNT(pg.id)::int AS count
FROM generate_series(
    sqlc.arg('from_date')::date,
    sqlc.arg('to_date')::date,
    INTERVAL '1 day'
) d
LEFT JOIN pages pg
    ON pg.created_at::date = d::date
    AND pg.deleted_at IS NULL
GROUP BY d
ORDER BY d ASC;
