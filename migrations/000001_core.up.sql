BEGIN;

CREATE TABLE libraries (
    library_id UUID PRIMARY KEY,
    library_type TEXT NOT NULL CHECK (library_type IN ('personal', 'family')),
    owner_subject_id TEXT NOT NULL,
    lifecycle_state TEXT NOT NULL DEFAULT 'active' CHECK (lifecycle_state IN ('active', 'disabled')),
    revision BIGINT NOT NULL DEFAULT 1 CHECK (revision >= 1),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE library_memberships (
    library_id UUID NOT NULL REFERENCES libraries(library_id) ON DELETE CASCADE,
    subject_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('owner', 'manager', 'contributor', 'viewer')),
    revision BIGINT NOT NULL DEFAULT 1 CHECK (revision >= 1),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (library_id, subject_id)
);

CREATE TABLE original_objects (
    original_object_id UUID PRIMARY KEY,
    storage_backend TEXT NOT NULL,
    storage_key TEXT NOT NULL UNIQUE,
    content_sha256 CHAR(64) NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    byte_size BIGINT NOT NULL CHECK (byte_size >= 0),
    media_type TEXT NOT NULL,
    integrity_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE assets (
    asset_id UUID PRIMARY KEY,
    library_id UUID NOT NULL REFERENCES libraries(library_id) ON DELETE RESTRICT,
    owner_subject_id TEXT NOT NULL,
    original_object_id UUID NOT NULL UNIQUE REFERENCES original_objects(original_object_id) ON DELETE RESTRICT,
    original_filename TEXT,
    media_type TEXT NOT NULL,
    content_sha256 CHAR(64) NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    byte_size BIGINT NOT NULL CHECK (byte_size >= 0),
    capture_time TIMESTAMPTZ,
    capture_time_zone TEXT,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revision BIGINT NOT NULL DEFAULT 1 CHECK (revision >= 1),
    favorite BOOLEAN NOT NULL DEFAULT FALSE,
    lifecycle_state TEXT NOT NULL DEFAULT 'active' CHECK (lifecycle_state IN ('active', 'archived', 'trash', 'purged')),
    protection_state TEXT NOT NULL DEFAULT 'standard' CHECK (protection_state IN ('standard', 'protected')),
    processing_state TEXT NOT NULL DEFAULT 'pending' CHECK (processing_state IN ('pending', 'ready', 'degraded', 'failed')),
    server_storage_state TEXT NOT NULL DEFAULT 'staging' CHECK (server_storage_state IN ('staging', 'verified', 'unavailable', 'corrupt'))
);

CREATE INDEX assets_library_imported_idx ON assets (library_id, imported_at DESC, asset_id);
CREATE INDEX assets_library_capture_idx ON assets (library_id, capture_time DESC NULLS LAST, asset_id);

CREATE TABLE device_asset_states (
    device_id TEXT NOT NULL,
    asset_id UUID NOT NULL REFERENCES assets(asset_id) ON DELETE CASCADE,
    local_presence TEXT NOT NULL CHECK (local_presence IN ('absent', 'preview', 'optimized', 'original')),
    upload_state TEXT NOT NULL CHECK (upload_state IN ('device_only', 'waiting', 'uploading', 'verifying', 'verified_remote', 'paused', 'failed')),
    last_verified_server_state TEXT,
    last_sync_cursor TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (device_id, asset_id)
);

CREATE TABLE upload_sessions (
    upload_id UUID PRIMARY KEY,
    actor_subject_id TEXT NOT NULL,
    library_id UUID NOT NULL REFERENCES libraries(library_id) ON DELETE CASCADE,
    expected_size BIGINT NOT NULL CHECK (expected_size >= 0),
    expected_sha256 CHAR(64) CHECK (expected_sha256 IS NULL OR expected_sha256 ~ '^[0-9a-f]{64}$'),
    received_bytes BIGINT NOT NULL DEFAULT 0 CHECK (received_bytes >= 0 AND received_bytes <= expected_size),
    part_size BIGINT NOT NULL CHECK (part_size > 0),
    state TEXT NOT NULL CHECK (state IN ('open', 'receiving', 'verifying', 'complete', 'failed', 'expired', 'cancelled')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX upload_sessions_expiry_idx ON upload_sessions (state, expires_at);

CREATE TABLE sync_changes (
    sequence BIGSERIAL PRIMARY KEY,
    change_id UUID NOT NULL UNIQUE,
    library_id UUID NOT NULL REFERENCES libraries(library_id) ON DELETE CASCADE,
    resource_type TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    operation TEXT NOT NULL CHECK (operation IN ('create', 'update', 'delete', 'restore')),
    revision BIGINT NOT NULL CHECK (revision >= 1),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX sync_changes_library_sequence_idx ON sync_changes (library_id, sequence);

CREATE TABLE jobs (
    job_id UUID PRIMARY KEY,
    job_type TEXT NOT NULL,
    resource_id TEXT,
    state TEXT NOT NULL CHECK (state IN ('pending', 'leased', 'complete', 'failed', 'cancelled')),
    attempt INTEGER NOT NULL DEFAULT 0 CHECK (attempt >= 0),
    available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    lease_owner TEXT,
    lease_expires_at TIMESTAMPTZ,
    last_error_code TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX jobs_ready_idx ON jobs (state, available_at) WHERE state IN ('pending', 'failed');

CREATE TABLE idempotency_records (
    actor_subject_id TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    request_sha256 CHAR(64) NOT NULL CHECK (request_sha256 ~ '^[0-9a-f]{64}$'),
    response_status INTEGER NOT NULL CHECK (response_status BETWEEN 100 AND 599),
    response_body JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (actor_subject_id, idempotency_key)
);

CREATE INDEX idempotency_records_expiry_idx ON idempotency_records (expires_at);

COMMIT;
