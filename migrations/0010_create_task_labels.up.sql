-- Task labels table (many-to-many)
CREATE TABLE task_labels (
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES project_labels(id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    added_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    PRIMARY KEY (task_id, label_id)
);

CREATE INDEX idx_task_labels_label_id ON task_labels(label_id);
