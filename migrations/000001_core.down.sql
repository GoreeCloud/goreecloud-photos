BEGIN;

DROP TABLE IF EXISTS idempotency_records;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS sync_changes;
DROP TABLE IF EXISTS upload_sessions;
DROP TABLE IF EXISTS device_asset_states;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS original_objects;
DROP TABLE IF EXISTS library_memberships;
DROP TABLE IF EXISTS libraries;

COMMIT;
