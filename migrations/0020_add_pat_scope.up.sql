ALTER TABLE personal_access_tokens
    ADD COLUMN scope TEXT NOT NULL DEFAULT 'read_write'
        CHECK (scope IN ('read_only', 'read_write'));
