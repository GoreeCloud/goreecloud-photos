BEGIN;

DROP TABLE IF EXISTS upload_parts;

ALTER TABLE upload_sessions
    DROP CONSTRAINT IF EXISTS upload_sessions_device_id_length,
    DROP CONSTRAINT IF EXISTS upload_sessions_media_type_length,
    DROP CONSTRAINT IF EXISTS upload_sessions_original_filename_length,
    DROP COLUMN IF EXISTS device_id,
    DROP COLUMN IF EXISTS media_type,
    DROP COLUMN IF EXISTS original_filename;

COMMIT;
