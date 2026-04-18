-- Modules: reusable sub-projects grouping tasks. Unlike cycles, modules have
-- optional dates, explicit status + lead + member list, and a task may belong
-- to many modules at once.
CREATE TYPE module_status AS ENUM (
    'backlog',
    'planned',
    'in_progress',
    'paused',
    'completed',
    'cancelled'
);

CREATE TABLE modules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status module_status NOT NULL DEFAULT 'backlog',
    start_date DATE,
    end_date DATE,
    lead_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT modules_dates_valid
        CHECK (start_date IS NULL OR end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_modules_project_id ON modules(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_modules_lead_id ON modules(lead_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_modules_status ON modules(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_modules_deleted_at ON modules(deleted_at) WHERE deleted_at IS NULL;

-- Module membership. Members must also be project members; enforced in handler.
CREATE TABLE module_members (
    module_id UUID NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    user_id   UUID NOT NULL REFERENCES users(id)   ON DELETE CASCADE,
    added_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    added_by  UUID NOT NULL REFERENCES users(id)   ON DELETE RESTRICT,
    PRIMARY KEY (module_id, user_id)
);

CREATE INDEX idx_module_members_user_id ON module_members(user_id);

-- Module-task links. Intentionally NO UNIQUE(task_id): a task can belong to
-- many modules simultaneously (unlike cycles, where the cycle_tasks table
-- enforces UNIQUE(task_id) for the one-cycle-per-task rule).
CREATE TABLE module_tasks (
    module_id UUID NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    task_id   UUID NOT NULL REFERENCES tasks(id)   ON DELETE CASCADE,
    added_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    added_by  UUID NOT NULL REFERENCES users(id)   ON DELETE RESTRICT,
    PRIMARY KEY (module_id, task_id)
);

CREATE INDEX idx_module_tasks_task_id ON module_tasks(task_id);
