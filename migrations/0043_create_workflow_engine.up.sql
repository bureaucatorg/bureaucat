-- Opt-in workflow engine. A row exists once an admin has saved a workflow;
-- enabled toggles enforcement without discarding the configuration.
CREATE TABLE project_workflows (
    project_id UUID PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    enabled    BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE workflow_states (
    state_id   UUID PRIMARY KEY REFERENCES project_states(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    is_entry   BOOLEAN NOT NULL DEFAULT FALSE,
    pos_x      DOUBLE PRECISION NOT NULL DEFAULT 0,
    pos_y      DOUBLE PRECISION NOT NULL DEFAULT 0
);

CREATE INDEX idx_workflow_states_project_id ON workflow_states(project_id);

CREATE TABLE workflow_transitions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    from_state_id   UUID NOT NULL REFERENCES project_states(id) ON DELETE CASCADE,
    to_state_id     UUID NOT NULL REFERENCES project_states(id) ON DELETE CASCADE,
    allow_assignees BOOLEAN NOT NULL DEFAULT FALSE,
    allow_creator   BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE (from_state_id, to_state_id),
    CHECK (from_state_id <> to_state_id)
);

CREATE INDEX idx_workflow_transitions_project_id ON workflow_transitions(project_id);

CREATE TABLE workflow_transition_movers (
    transition_id UUID NOT NULL REFERENCES workflow_transitions(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (transition_id, user_id)
);

CREATE TABLE workflow_state_actions (
    state_id UUID NOT NULL REFERENCES workflow_states(state_id) ON DELETE CASCADE,
    action   VARCHAR(32) NOT NULL CHECK (action IN ('add_assignee', 'remove_assignee')),
    user_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (state_id, action, user_id)
);
