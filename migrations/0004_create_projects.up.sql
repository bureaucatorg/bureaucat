-- Projects table
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_key VARCHAR(10) NOT NULL UNIQUE,  -- e.g., "DEVOP"
    name VARCHAR(255) NOT NULL,
    description TEXT,
    icon_id UUID REFERENCES uploads(id) ON DELETE SET NULL,
    cover_id UUID REFERENCES uploads(id) ON DELETE SET NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  -- soft delete
);

CREATE INDEX idx_projects_project_key ON projects(project_key);
CREATE INDEX idx_projects_created_by ON projects(created_by);
CREATE INDEX idx_projects_deleted_at ON projects(deleted_at) WHERE deleted_at IS NULL;
