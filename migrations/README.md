# GoreeCloud Photos — PostgreSQL Migrations

These migrations are the initial Experimental schema baseline.

They are not evidence that a production database has been deployed or restored.

Rules:

- application code supplies UUIDv7 identifiers;
- migrations do not require a PostgreSQL extension merely to generate identifiers;
- original media remains outside PostgreSQL and is referenced by immutable object identity;
- the content checksum is not globally unique, so the schema does not silently introduce cross-user content deduplication;
- synchronization changes are ordered per database sequence and exposed to clients through opaque cursors;
- down migrations are development/recovery aids and do not substitute for production rollback or backup.
