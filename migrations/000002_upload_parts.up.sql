BEGIN;

CREATE TABLE upload_parts (
    upload_id UUID NOT NULL REFERENCES upload_sessions(upload_id) ON DELETE CASCADE,
    part_number INTEGER NOT NULL CHECK (part_number >= 1),
    byte_size BIGINT NOT NULL CHECK (byte_size > 0),
    content_sha256 CHAR(64) NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (upload_id, part_number)
);

CREATE INDEX upload_parts_upload_created_idx ON upload_parts (upload_id, created_at, part_number);

COMMIT;
