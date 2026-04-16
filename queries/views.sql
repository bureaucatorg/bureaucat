-- ==================== PROJECT VIEWS ====================

-- name: CreateProjectView :one
INSERT INTO project_views (
    project_id, slug, name, description, visibility, owner_id,
    filter_tree, group_by, sort_by, sort_dir, position
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, project_id, slug, name, description, visibility, owner_id,
          filter_tree, group_by, sort_by, sort_dir, position,
          created_at, updated_at, deleted_at;

-- name: GetProjectViewBySlug :one
SELECT id, project_id, slug, name, description, visibility, owner_id,
       filter_tree, group_by, sort_by, sort_dir, position,
       created_at, updated_at, deleted_at
FROM project_views
WHERE project_id = $1 AND slug = $2 AND deleted_at IS NULL;

-- name: GetProjectViewByID :one
SELECT id, project_id, slug, name, description, visibility, owner_id,
       filter_tree, group_by, sort_by, sort_dir, position,
       created_at, updated_at, deleted_at
FROM project_views
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListProjectViews :many
-- Returns every shared view in the project plus private views owned by the caller.
SELECT id, project_id, slug, name, description, visibility, owner_id,
       filter_tree, group_by, sort_by, sort_dir, position,
       created_at, updated_at, deleted_at
FROM project_views
WHERE project_id = $1
  AND deleted_at IS NULL
  AND (visibility = 'shared' OR owner_id = $2)
ORDER BY position ASC, created_at ASC;

-- name: ProjectViewSlugExists :one
SELECT EXISTS (
    SELECT 1 FROM project_views
    WHERE project_id = $1 AND slug = $2 AND deleted_at IS NULL
) AS exists;

-- name: UpdateProjectView :one
UPDATE project_views
SET name        = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    visibility  = COALESCE(sqlc.narg('visibility'), visibility),
    filter_tree = COALESCE(sqlc.narg('filter_tree'), filter_tree),
    group_by    = COALESCE(sqlc.narg('group_by'), group_by),
    sort_by     = COALESCE(sqlc.narg('sort_by'), sort_by),
    sort_dir    = COALESCE(sqlc.narg('sort_dir'), sort_dir),
    position    = COALESCE(sqlc.narg('position'), position),
    updated_at  = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, project_id, slug, name, description, visibility, owner_id,
          filter_tree, group_by, sort_by, sort_dir, position,
          created_at, updated_at, deleted_at;

-- name: SoftDeleteProjectView :exec
UPDATE project_views
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ReorderProjectViews :exec
-- Pass a JSON array of {id, new_position} objects.
UPDATE project_views pv
SET position = u.new_position, updated_at = NOW()
FROM jsonb_to_recordset(@items::jsonb) AS u(id uuid, new_position int)
WHERE pv.id = u.id AND pv.project_id = $1 AND pv.deleted_at IS NULL;

-- name: TransferOwnedSharedViews :exec
-- When a member leaves a project, keep their shared views alive by reassigning ownership.
UPDATE project_views
SET owner_id = $3, updated_at = NOW()
WHERE project_id = $1 AND owner_id = $2 AND visibility = 'shared' AND deleted_at IS NULL;

-- name: SoftDeleteOwnedPrivateViews :exec
-- When a member leaves a project, their private views are soft-deleted.
UPDATE project_views
SET deleted_at = NOW(), updated_at = NOW()
WHERE project_id = $1 AND owner_id = $2 AND visibility = 'private' AND deleted_at IS NULL;
