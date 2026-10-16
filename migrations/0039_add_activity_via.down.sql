-- Rows written with checksum_version 2 will no longer verify after this rollback.
ALTER TABLE activity_log
    DROP COLUMN IF EXISTS checksum_version,
    DROP COLUMN IF EXISTS via;
