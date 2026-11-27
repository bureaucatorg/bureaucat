-- Structured templates: typed fields rendered into the description, plus defaults applied on create.
ALTER TABLE task_templates
    ADD COLUMN fields JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN default_priority INT,
    ADD COLUMN default_labels UUID[] NOT NULL DEFAULT '{}',
    ADD COLUMN default_assignees UUID[] NOT NULL DEFAULT '{}';

ALTER TABLE projects ADD COLUMN require_template BOOLEAN NOT NULL DEFAULT FALSE;
