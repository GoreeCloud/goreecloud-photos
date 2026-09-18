BEGIN;

ALTER TABLE upload_sessions
    DROP COLUMN IF EXISTS capture_time_zone,
    DROP COLUMN IF EXISTS capture_time,
    DROP COLUMN IF EXISTS device_id,
    DROP COLUMN IF EXISTS media_type,
    DROP COLUMN IF EXISTS original_filename;

COMMIT;
