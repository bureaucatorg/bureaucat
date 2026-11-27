ALTER TABLE projects DROP COLUMN IF EXISTS require_template;

ALTER TABLE task_templates
    DROP COLUMN IF EXISTS default_assignees,
    DROP COLUMN IF EXISTS default_labels,
    DROP COLUMN IF EXISTS default_priority,
    DROP COLUMN IF EXISTS fields;
