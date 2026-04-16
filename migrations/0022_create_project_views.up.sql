-- Project views: saved filter/sort/group-by combinations.
--
-- Visibility: private views are owner-only. Shared views are readable by any
-- project member and editable by the owner or a project admin.

CREATE TYPE view_visibility AS ENUM ('private', 'shared');

CREATE TABLE project_views (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    slug            VARCHAR(64) NOT NULL,
    name            VARCHAR(120) NOT NULL,
    description     TEXT,
    visibility      view_visibility NOT NULL DEFAULT 'private',
    owner_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filter_tree     JSONB NOT NULL,
    group_by        VARCHAR(32) NOT NULL DEFAULT 'state_type',
    sort_by         VARCHAR(32) NOT NULL DEFAULT 'created_at',
    sort_dir        VARCHAR(4)  NOT NULL DEFAULT 'desc' CHECK (sort_dir IN ('asc','desc')),
    position        INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_project_views_project_slug
    ON project_views(project_id, slug) WHERE deleted_at IS NULL;

CREATE INDEX idx_project_views_project_visibility
    ON project_views(project_id, visibility) WHERE deleted_at IS NULL;

CREATE INDEX idx_project_views_owner
    ON project_views(owner_id) WHERE deleted_at IS NULL;
