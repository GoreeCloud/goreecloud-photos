BEGIN;

ALTER TABLE upload_sessions
    ADD COLUMN original_filename TEXT,
    ADD COLUMN media_type TEXT,
    ADD COLUMN device_id TEXT,
    ADD COLUMN capture_time TIMESTAMPTZ,
    ADD COLUMN capture_time_zone TEXT;

COMMIT;
