# GoreeCloud Photos — PostgreSQL Migrations

These migrations are the Experimental schema baseline.

They are not evidence that a production database has been deployed, backed up, or restored.

## Current migrations

- `000001_core` — libraries, memberships, original objects, assets, device asset state, upload sessions, synchronization changes, jobs, and idempotency records.
- `000002_upload_parts` — durable per-part upload receipt/checksum evidence keyed by upload session and one-based part number.

Rules:

- application code supplies UUIDv7 identifiers;
- migrations do not require a PostgreSQL extension merely to generate identifiers;
- original media remains outside PostgreSQL and is referenced by immutable object identity;
- the content checksum is not globally unique, so the schema does not silently introduce cross-user content deduplication;
- synchronization changes are ordered per database sequence and exposed to clients through opaque cursors;
- upload part rows are durable receipt evidence, not proof that media bytes have completed ingestion or backup;
- down migrations are development/recovery aids and do not substitute for production rollback or backup.

Exact-head CI validates migration ordering, readiness before/after required migrations, upload-session persistence behavior, and up/down execution against pinned PostgreSQL 18.6.
