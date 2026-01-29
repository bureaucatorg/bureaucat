-- Tasks table
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    task_number INT NOT NULL,  -- Sequential per project
    title VARCHAR(500) NOT NULL,
    description TEXT,
    state_id UUID NOT NULL REFERENCES project_states(id) ON DELETE RESTRICT,
    priority INT NOT NULL DEFAULT 0,  -- 0=none, 1=low, 2=medium, 3=high, 4=urgent
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,  -- soft delete
    UNIQUE(project_id, task_number)
);

-- Composite index for task display ID lookup (e.g., "DEVOP-780")
CREATE INDEX idx_tasks_project_task_number ON tasks(project_id, task_number);
CREATE INDEX idx_tasks_state_id ON tasks(state_id);
CREATE INDEX idx_tasks_created_by ON tasks(created_by);
CREATE INDEX idx_tasks_deleted_at ON tasks(deleted_at) WHERE deleted_at IS NULL;
