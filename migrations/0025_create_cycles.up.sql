-- Cycles table (time-boxed task containers, like sprints)
CREATE TABLE cycles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT cycles_dates_valid CHECK (end_date >= start_date)
);

CREATE INDEX idx_cycles_project_id ON cycles(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_cycles_dates ON cycles(start_date, end_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_cycles_deleted_at ON cycles(deleted_at) WHERE deleted_at IS NULL;

-- Cycle-task membership. A task belongs to at most one cycle (UNIQUE on task_id).
CREATE TABLE cycle_tasks (
    cycle_id UUID NOT NULL REFERENCES cycles(id) ON DELETE CASCADE,
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    added_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    PRIMARY KEY (cycle_id, task_id),
    UNIQUE (task_id)
);

CREATE INDEX idx_cycle_tasks_task_id ON cycle_tasks(task_id);
