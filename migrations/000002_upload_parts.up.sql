BEGIN;

ALTER TABLE upload_sessions
    ADD COLUMN original_filename TEXT,
    ADD COLUMN media_type TEXT,
    ADD COLUMN device_id TEXT;

ALTER TABLE upload_sessions
    ADD CONSTRAINT upload_sessions_original_filename_length
        CHECK (original_filename IS NULL OR char_length(original_filename) BETWEEN 1 AND 1024),
    ADD CONSTRAINT upload_sessions_media_type_length
        CHECK (media_type IS NULL OR char_length(media_type) BETWEEN 1 AND 255),
    ADD CONSTRAINT upload_sessions_device_id_length
        CHECK (device_id IS NULL OR char_length(device_id) BETWEEN 1 AND 256);

CREATE TABLE upload_parts (
    upload_id UUID NOT NULL REFERENCES upload_sessions(upload_id) ON DELETE CASCADE,
    part_number INTEGER NOT NULL CHECK (part_number >= 1),
    byte_offset BIGINT NOT NULL CHECK (byte_offset >= 0),
    byte_size BIGINT NOT NULL CHECK (byte_size > 0),
    content_sha256 CHAR(64) NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    storage_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (upload_id, part_number),
    UNIQUE (upload_id, byte_offset)
);

CREATE INDEX upload_parts_upload_offset_idx ON upload_parts (upload_id, byte_offset);

COMMIT;
