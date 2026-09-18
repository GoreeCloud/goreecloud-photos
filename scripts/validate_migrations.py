#!/usr/bin/env python3
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

MIGRATIONS = [
    (
        ROOT / "migrations" / "000001_core.up.sql",
        ROOT / "migrations" / "000001_core.down.sql",
        [
            "libraries",
            "library_memberships",
            "original_objects",
            "assets",
            "device_asset_states",
            "upload_sessions",
            "sync_changes",
            "jobs",
            "idempotency_records",
        ],
    ),
    (
        ROOT / "migrations" / "000002_upload_parts.up.sql",
        ROOT / "migrations" / "000002_upload_parts.down.sql",
        ["upload_parts"],
    ),
]

for up_path, down_path, tables in MIGRATIONS:
    for path in (up_path, down_path):
        if not path.is_file():
            raise SystemExit(f"missing migration: {path.relative_to(ROOT)}")

    up = up_path.read_text(encoding="utf-8")
    down = down_path.read_text(encoding="utf-8")

    if not up.startswith("BEGIN;") or not up.rstrip().endswith("COMMIT;"):
        raise SystemExit(f"{up_path.name} must be transaction-bounded")
    if not down.startswith("BEGIN;") or not down.rstrip().endswith("COMMIT;"):
        raise SystemExit(f"{down_path.name} must be transaction-bounded")
    if "CREATE EXTENSION" in up.upper():
        raise SystemExit(f"{up_path.name} must not add an undeclared PostgreSQL extension")

    for table in tables:
        if f"CREATE TABLE {table} " not in up:
            raise SystemExit(f"{up_path.name} missing table: {table}")
        if f"DROP TABLE IF EXISTS {table};" not in down:
            raise SystemExit(f"{down_path.name} missing table: {table}")

core_up = MIGRATIONS[0][0].read_text(encoding="utf-8")
if "UNIQUE (content_sha256)" in core_up or "content_sha256 CHAR(64) NOT NULL UNIQUE" in core_up:
    raise SystemExit("initial schema must not silently enable global checksum deduplication")

upload_parts_up = MIGRATIONS[1][0].read_text(encoding="utf-8")
for marker in [
    "PRIMARY KEY (upload_id, part_number)",
    "REFERENCES upload_sessions(upload_id) ON DELETE CASCADE",
    "content_sha256 CHAR(64) NOT NULL",
]:
    if marker not in upload_parts_up:
        raise SystemExit(f"upload-parts migration missing invariant: {marker}")

print("GoreeCloud Photos migration baseline validation passed.")
