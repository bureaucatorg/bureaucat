-- via: channel that caused the activity. Existing rows predate tracking and default to 'web'.
-- checksum_version: 1 = original formula, 2 = formula that also covers via.
ALTER TABLE activity_log
    ADD COLUMN via TEXT NOT NULL DEFAULT 'web' CHECK (via IN ('web', 'assistant', 'pat')),
    ADD COLUMN checksum_version SMALLINT NOT NULL DEFAULT 1;
