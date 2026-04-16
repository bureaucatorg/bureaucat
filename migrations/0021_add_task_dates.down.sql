ALTER TABLE tasks
    DROP COLUMN IF EXISTS start_date,
    DROP COLUMN IF EXISTS due_date;
