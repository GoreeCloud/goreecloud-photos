#!/usr/bin/env python3
from __future__ import annotations

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

REQUIRED_FILES = [
    "README.md",
    "SPECIFICATIONS.md",
    "FEATURES.md",
    "FEATURE-ROADMAP.md",
    "BENEFITS.md",
    "COMPETITIVE-OBJECTIVES.md",
    "BRANDING.md",
    "USER-MANUAL.md",
    "PRIVACY POLICY.md",
    "NOTES.md",
    "SECURITY.md",
    "ARCHITECTURE.md",
    "DATA-MODEL.md",
    "API.md",
    "RECOVERY.md",
    "PLATFORM-INTEGRATIONS.md",
    "DEPENDENCIES.md",
    "GLAZE-UI-ACCEPTANCE.md",
    "LICENSE",
    ".gitignore",
    ".editorconfig",
    "goreecloud.platform.yaml",
    ".github/PULL_REQUEST_TEMPLATE.md",
    "contracts/README.md",
    "contracts/asset.v1.schema.json",
    "contracts/sync-change.v1.schema.json",
    "contracts/upload-session.v1.schema.json",
]

PLATFORM_SYSTEM_KEYS = [
    "manager:",
    "privacy_shield:",
    "wardveil_security:",
    "everkeep:",
    "glaze_ui:",
    "mesh:",
    "identity:",
    "policy:",
    "observability:",
]

SCHEMA_REQUIRED = {
    "contracts/asset.v1.schema.json": {
        "asset_id",
        "library_id",
        "owner_subject_id",
        "media_type",
        "content_sha256",
        "byte_size",
        "revision",
        "lifecycle_state",
        "server_storage_state",
    },
    "contracts/sync-change.v1.schema.json": {
        "change_id",
        "library_id",
        "resource_type",
        "resource_id",
        "operation",
        "revision",
        "changed_at",
    },
    "contracts/upload-session.v1.schema.json": {
        "upload_id",
        "library_id",
        "expected_size",
        "received_bytes",
        "part_size",
        "state",
        "expires_at",
    },
}


def fail(message: str) -> None:
    raise SystemExit(message)


for relative in REQUIRED_FILES:
    path = ROOT / relative
    if not path.is_file():
        fail(f"missing required repository file: {relative}")

readme = (ROOT / "README.md").read_text(encoding="utf-8")
if "Concept / Planning" not in readme:
    fail("README must preserve Concept / Planning lifecycle truth")

features = (ROOT / "FEATURES.md").read_text(encoding="utf-8")
if "Current implemented product functionality" not in features or "None." not in features:
    fail("FEATURES.md must explicitly preserve the no-runtime implementation state")

platform = (ROOT / "goreecloud.platform.yaml").read_text(encoding="utf-8")
for required in ['schema_version: "0.4"', "lifecycle: concept", 'glaze_ui_required: "1.5.1"']:
    if required not in platform:
        fail(f"platform declaration missing required baseline: {required}")
for key in PLATFORM_SYSTEM_KEYS:
    if key not in platform:
        fail(f"platform declaration missing Integral Platform System: {key}")

for relative, expected_required in SCHEMA_REQUIRED.items():
    path = ROOT / relative
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        fail(f"invalid JSON in {relative}: {exc}")

    if data.get("$schema") != "https://json-schema.org/draft/2020-12/schema":
        fail(f"{relative} must use JSON Schema 2020-12")
    if data.get("type") != "object":
        fail(f"{relative} must define an object")
    required = set(data.get("required", []))
    missing = expected_required - required
    if missing:
        fail(f"{relative} missing required contract fields: {sorted(missing)}")
    if not isinstance(data.get("properties"), dict):
        fail(f"{relative} missing properties object")

architecture = (ROOT / "ARCHITECTURE.md").read_text(encoding="utf-8")
for marker in ["Go 1.27.1", "PostgreSQL", "S3-compatible", "Kotlin + Jetpack Compose", "Rust + GTK4"]:
    if marker not in architecture:
        fail(f"ARCHITECTURE.md missing selected stack marker: {marker}")

glaze = (ROOT / "GLAZE-UI-ACCEPTANCE.md").read_text(encoding="utf-8")
for marker in ["Glaze UI V1.5 / 1.5.1", "keyboard", "screen-reader", "reduced motion", "exact-revision"]:
    if marker not in glaze:
        fail(f"GLAZE-UI-ACCEPTANCE.md missing acceptance marker: {marker}")

integrations = (ROOT / "PLATFORM-INTEGRATIONS.md").read_text(encoding="utf-8")
for name in [
    "GoreeCloud Manager",
    "Privacy Shield",
    "Wardveil Security",
    "Everkeep",
    "Glaze UI",
    "GoreeCloud Mesh",
    "GoreeCloud Identity",
    "GoreeCloud Policy",
    "GoreeCloud Observability",
]:
    if name not in integrations:
        fail(f"PLATFORM-INTEGRATIONS.md missing: {name}")

print("GoreeCloud Photos repository baseline validation passed.")
