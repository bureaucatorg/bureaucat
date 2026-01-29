-- State type enum
CREATE TYPE state_type AS ENUM ('backlog', 'unstarted', 'started', 'completed', 'cancelled');

-- Project states table (workflow states per project)
CREATE TABLE project_states (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    state_type state_type NOT NULL,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) DEFAULT '#6B7280',  -- hex color
    position INT NOT NULL DEFAULT 0,     -- for ordering
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);

CREATE INDEX idx_project_states_project_id ON project_states(project_id);
