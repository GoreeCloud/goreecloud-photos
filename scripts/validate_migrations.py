#!/usr/bin/env python3
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

MIGRATIONS = [
    (ROOT / "migrations" / "000001_core.up.sql", ROOT / "migrations" / "000001_core.down.sql"),
    (ROOT / "migrations" / "000002_upload_parts.up.sql", ROOT / "migrations" / "000002_upload_parts.down.sql"),
]

for up_path, down_path in MIGRATIONS:
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

core_up = MIGRATIONS[0][0].read_text(encoding="utf-8")
core_down = MIGRATIONS[0][1].read_text(encoding="utf-8")
for table in ["libraries","library_memberships","original_objects","assets","device_asset_states","upload_sessions","sync_changes","jobs","idempotency_records"]:
    if f"CREATE TABLE {table} " not in core_up:
        raise SystemExit(f"core migration missing table: {table}")
    if f"DROP TABLE IF EXISTS {table};" not in core_down:
        raise SystemExit(f"core down migration missing table: {table}")

upload_up = MIGRATIONS[1][0].read_text(encoding="utf-8")
upload_down = MIGRATIONS[1][1].read_text(encoding="utf-8")
if "ALTER TABLE upload_sessions" not in upload_up:
    raise SystemExit("upload-parts migration must extend upload_sessions")
if "CREATE TABLE upload_parts " not in upload_up:
    raise SystemExit("upload-parts migration missing upload_parts")
if "DROP TABLE IF EXISTS upload_parts;" not in upload_down:
    raise SystemExit("upload-parts down migration missing upload_parts")

if "UNIQUE (content_sha256)" in core_up or "content_sha256 CHAR(64) NOT NULL UNIQUE" in core_up:
    raise SystemExit("core schema must not silently enable global checksum deduplication")

print("GoreeCloud Photos migration baseline validation passed.")
