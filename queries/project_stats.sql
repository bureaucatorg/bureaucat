-- ==================== PROJECT STATS ====================

-- name: ProjectStatsTotals :one
-- Blocked/blocking only count links where neither task is completed or cancelled.
WITH pt AS (
    SELECT t.id, t.parent_task_id, ps.state_type
    FROM tasks t
    JOIN project_states ps ON ps.id = t.state_id
    WHERE t.project_id = sqlc.arg('project_id')::uuid AND t.deleted_at IS NULL
),
pa AS (
    SELECT a.upload_id
    FROM attachments a
    JOIN pt ON a.entity_type = 'task' AND a.entity_id = pt.id
    UNION ALL
    SELECT a.upload_id
    FROM attachments a
    JOIN comments c ON a.entity_type = 'comment' AND a.entity_id = c.id AND c.deleted_at IS NULL
    JOIN pt ON pt.id = c.task_id
),
open_blocks AS (
    SELECT tb.task_id, tb.blocker_task_id
    FROM task_blockers tb
    JOIN pt b ON b.id = tb.task_id AND b.state_type NOT IN ('completed', 'cancelled')
    JOIN pt k ON k.id = tb.blocker_task_id AND k.state_type NOT IN ('completed', 'cancelled')
)
SELECT
    (SELECT COUNT(*) FROM pt WHERE parent_task_id IS NULL)::bigint AS tasks,
    (SELECT COUNT(*) FROM pt WHERE parent_task_id IS NOT NULL)::bigint AS subtasks,
    (SELECT COUNT(DISTINCT task_id) FROM open_blocks)::bigint AS blocked_tasks,
    (SELECT COUNT(DISTINCT blocker_task_id) FROM open_blocks)::bigint AS blocking_tasks,
    (SELECT COUNT(*) FROM pt
        WHERE state_type NOT IN ('completed', 'cancelled')
          AND NOT EXISTS (SELECT 1 FROM task_assignees ta WHERE ta.task_id = pt.id))::bigint AS unassigned_open,
    (SELECT COUNT(*) FROM pages WHERE project_id = sqlc.arg('project_id')::uuid AND deleted_at IS NULL)::bigint AS pages,
    (SELECT COUNT(*) FROM project_members WHERE project_id = sqlc.arg('project_id')::uuid)::bigint AS members,
    (SELECT COUNT(*) FROM cycles WHERE project_id = sqlc.arg('project_id')::uuid AND deleted_at IS NULL)::bigint AS cycles,
    (SELECT COUNT(*) FROM modules WHERE project_id = sqlc.arg('project_id')::uuid AND deleted_at IS NULL)::bigint AS modules,
    (SELECT COUNT(*) FROM pa)::bigint AS attachments,
    (SELECT COALESCE(SUM(u.size_bytes), 0) FROM pa JOIN uploads u ON u.id = pa.upload_id)::bigint AS attachments_bytes;

-- name: ProjectTasksByState :many
-- Top-level tasks per project state, in the project's state order.
SELECT ps.name, ps.state_type, ps.color, COUNT(t.id)::int AS count
FROM project_states ps
LEFT JOIN tasks t ON t.state_id = ps.id AND t.deleted_at IS NULL AND t.parent_task_id IS NULL
WHERE ps.project_id = sqlc.arg('project_id')::uuid
GROUP BY ps.id, ps.name, ps.state_type, ps.color, ps.position
ORDER BY ps.position ASC, ps.name ASC;

-- name: ProjectTasksByPriority :many
SELECT t.priority AS priority, COUNT(*)::int AS count
FROM tasks t
WHERE t.project_id = sqlc.arg('project_id')::uuid AND t.deleted_at IS NULL AND t.parent_task_id IS NULL
GROUP BY t.priority
ORDER BY t.priority ASC;

-- name: ProjectWorkload :many
-- Open tasks and subtasks per assignee.
SELECT u.id AS user_id, u.username, u.first_name, u.last_name, u.avatar_url, COUNT(*)::int AS open_tasks
FROM tasks t
JOIN project_states ps ON ps.id = t.state_id AND ps.state_type NOT IN ('completed', 'cancelled')
JOIN task_assignees ta ON ta.task_id = t.id
JOIN users u ON u.id = ta.user_id
WHERE t.project_id = sqlc.arg('project_id')::uuid AND t.deleted_at IS NULL
GROUP BY u.id, u.username, u.first_name, u.last_name, u.avatar_url
ORDER BY open_tasks DESC, u.first_name ASC, u.last_name ASC;

-- name: ProjectStatsSeries :many
-- One row per day in the range. "completed" counts top-level tasks moved into a completed state that day.
WITH days AS (
    SELECT d::date AS day
    FROM generate_series(sqlc.arg('from_date')::date, sqlc.arg('to_date')::date, INTERVAL '1 day') d
),
pt AS (
    SELECT id, parent_task_id, created_at
    FROM tasks
    WHERE project_id = sqlc.arg('project_id')::uuid AND deleted_at IS NULL
),
task_days AS (
    SELECT created_at::date AS day,
           COUNT(*) FILTER (WHERE parent_task_id IS NULL) AS tasks,
           COUNT(*) FILTER (WHERE parent_task_id IS NOT NULL) AS subtasks
    FROM pt
    GROUP BY 1
),
completed_days AS (
    SELECT al.created_at::date AS day, COUNT(DISTINCT al.task_id) AS n
    FROM activity_log al
    JOIN pt ON pt.id = al.task_id AND pt.parent_task_id IS NULL
    JOIN project_states ps ON ps.id = (al.new_value->>'state_id')::uuid AND ps.state_type = 'completed'
    WHERE al.activity_type = 'state_changed'
    GROUP BY 1
),
activity_days AS (
    SELECT al.created_at::date AS day, COUNT(*) AS n
    FROM activity_log al
    JOIN pt ON pt.id = al.task_id
    WHERE al.activity_type NOT IN ('comment_created', 'comment_updated', 'comment_deleted')
    GROUP BY 1
),
comment_days AS (
    SELECT c.created_at::date AS day, COUNT(*) AS n
    FROM comments c
    JOIN pt ON pt.id = c.task_id
    WHERE c.deleted_at IS NULL
    GROUP BY 1
),
attachment_days AS (
    SELECT a.created_at::date AS day, COUNT(*) AS n
    FROM attachments a
    LEFT JOIN comments c ON a.entity_type = 'comment' AND c.id = a.entity_id AND c.deleted_at IS NULL
    JOIN pt ON pt.id = CASE WHEN a.entity_type = 'task' THEN a.entity_id ELSE c.task_id END
    GROUP BY 1
),
page_days AS (
    SELECT created_at::date AS day, COUNT(*) AS n
    FROM pages
    WHERE project_id = sqlc.arg('project_id')::uuid AND deleted_at IS NULL
    GROUP BY 1
),
cycle_days AS (
    SELECT created_at::date AS day, COUNT(*) AS n
    FROM cycles
    WHERE project_id = sqlc.arg('project_id')::uuid AND deleted_at IS NULL
    GROUP BY 1
),
module_days AS (
    SELECT created_at::date AS day, COUNT(*) AS n
    FROM modules
    WHERE project_id = sqlc.arg('project_id')::uuid AND deleted_at IS NULL
    GROUP BY 1
)
SELECT days.day,
       COALESCE(td.tasks, 0)::int AS tasks,
       COALESCE(td.subtasks, 0)::int AS subtasks,
       COALESCE(cd.n, 0)::int AS completed,
       COALESCE(ad.n, 0)::int AS activity,
       COALESCE(cm.n, 0)::int AS comments,
       COALESCE(att.n, 0)::int AS attachments,
       COALESCE(pg.n, 0)::int AS pages,
       COALESCE(cy.n, 0)::int AS cycles,
       COALESCE(md.n, 0)::int AS modules
FROM days
LEFT JOIN task_days td ON td.day = days.day
LEFT JOIN completed_days cd ON cd.day = days.day
LEFT JOIN activity_days ad ON ad.day = days.day
LEFT JOIN comment_days cm ON cm.day = days.day
LEFT JOIN attachment_days att ON att.day = days.day
LEFT JOIN page_days pg ON pg.day = days.day
LEFT JOIN cycle_days cy ON cy.day = days.day
LEFT JOIN module_days md ON md.day = days.day
ORDER BY days.day ASC;
