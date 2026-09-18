# GoreeCloud Photos — PostgreSQL Migrations

These migrations are the Experimental schema baseline.

They are not evidence that a production database has been deployed or restored.

Current sequence:

1. `000001_core` — libraries, ownership, assets, original-object references, device state, upload sessions, synchronization changes, jobs, and idempotency records.
2. `000002_upload_parts` — resumable-upload request metadata and durable received-part metadata.

Rules:

- merged migration history is extended with a new migration rather than rewritten;
- application code supplies UUIDv7 identifiers;
- migrations do not require a PostgreSQL extension merely to generate identifiers;
- original media and staged part bytes remain outside PostgreSQL and are referenced by storage identity;
- the content checksum is not globally unique, so the schema does not silently introduce cross-user content deduplication;
- synchronization changes are ordered per database sequence and exposed to clients through opaque cursors;
- down migrations are development/recovery aids and do not substitute for production rollback or backup.
