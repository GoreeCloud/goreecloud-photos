#!/usr/bin/env python3
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
UP = ROOT / "migrations" / "000001_core.up.sql"
DOWN = ROOT / "migrations" / "000001_core.down.sql"

for path in (UP, DOWN):
    if not path.is_file():
        raise SystemExit(f"missing migration: {path.relative_to(ROOT)}")

up = UP.read_text(encoding="utf-8")
down = DOWN.read_text(encoding="utf-8")

if not up.startswith("BEGIN;") or not up.rstrip().endswith("COMMIT;"):
    raise SystemExit("up migration must be transaction-bounded")
if not down.startswith("BEGIN;") or not down.rstrip().endswith("COMMIT;"):
    raise SystemExit("down migration must be transaction-bounded")
if "CREATE EXTENSION" in up.upper():
    raise SystemExit("initial migration must not add an undeclared PostgreSQL extension")

tables = [
    "libraries",
    "library_memberships",
    "original_objects",
    "assets",
    "device_asset_states",
    "upload_sessions",
    "sync_changes",
    "jobs",
    "idempotency_records",
]

for table in tables:
    if f"CREATE TABLE {table} " not in up:
        raise SystemExit(f"up migration missing table: {table}")
    if f"DROP TABLE IF EXISTS {table};" not in down:
        raise SystemExit(f"down migration missing table: {table}")

if "UNIQUE (content_sha256)" in up or "content_sha256 CHAR(64) NOT NULL UNIQUE" in up:
    raise SystemExit("initial schema must not silently enable global checksum deduplication")

print("GoreeCloud Photos migration baseline validation passed.")
