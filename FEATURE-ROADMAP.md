# GoreeCloud Photos — Feature Roadmap

**Status:** Active roadmap control  
**Lifecycle:** Concept / Planning  
**As of:** 2026-09-18  
**Canonical specification:** `SPECIFICATIONS.md`  
**Repository:** `GoreeCloud/goreecloud-photos`  
**Implementation status:** Repository documentation foundation only; no supported Photos runtime is verified.

> Roadmap entries are planned work unless a later exact revision and evidence explicitly establish implementation and acceptance.

## Phase 0 — Product Foundation

**State:** In progress — documentation and governance baseline only.

Establish and validate:

- repository information architecture and required root controls;
- architectural boundaries for clients, services, storage, jobs, and background processing;
- stable Asset ID and ownership model;
- immutable-original and rebuildable-derivative rules;
- upload, synchronization, edit, sharing, deletion, and conflict contracts;
- GoreeCloud Identity model;
- Privacy Shield authorization model;
- Wardveil Security ingestion and access boundaries;
- Everkeep backup, restore, integrity, preservation, migration, and succession requirements;
- GoreeCloud Manager, Mesh, Policy, and Observability applicability;
- Glaze UI 1.5.1 application foundation;
- data classification, retention, export, and metadata rules;
- API versioning and compatibility policy;
- implementation stack and dependency decisions;
- CI, test, security, privacy, recovery, and representative-runtime evidence strategy.

**Exit condition:** approved architecture and contracts exist, repository baseline is complete, no planned integration is represented as implemented, and the first bounded implementation slice can be developed without inventing unresolved security, privacy, storage, or recovery behavior.

## Phase 1 — Core Photo Server

**State:** Planned.

Implement and verify:

- multi-user libraries;
- stable assets and ownership;
- immutable original media store;
- metadata store;
- resumable upload and transfer verification;
- ingestion validation;
- thumbnail and preview generation;
- basic authentication and authorization integration;
- trash and restore;
- basic Photos API;
- idempotent ingestion and job recovery.

## Phase 2 — Web Library

**State:** Planned.

Implement and verify:

- chronological timeline;
- asset viewer;
- upload;
- albums;
- favorites;
- download;
- basic date and metadata search;
- trash;
- responsive Glaze UI experience.

This is the first intended end-to-end server/client validation.

## Phase 3 — Mobile Backup

**State:** Planned.

Implement and verify:

- camera-library backup;
- background upload;
- selected-folder controls where supported;
- network, battery, charging, and roaming policies;
- resumable transfer recovery;
- backup-state truth;
- safe device-storage cleanup only after verified server durability;
- mobile timeline and offline/download behavior.

## Phase 4 — Desktop Library

**State:** Planned.

Implement and verify:

- folder imports;
- watch folders;
- large-library management;
- bulk operations;
- local cache;
- offline collections;
- large resumable imports;
- export and archival workflows;
- storage diagnostics.

## Phase 5 — Sharing and Family Libraries

**State:** Planned.

Implement and verify:

- direct user sharing;
- shared and collaborative albums;
- personal versus family ownership boundaries;
- household libraries;
- revocable and expiring links;
- password-protected and view/download-controlled links;
- metadata-removal options;
- explicit ownership preservation.

## Phase 6 — Optional Intelligence

**State:** Planned.

Implement as modular, privacy-controlled capabilities:

- object and scene recognition;
- OCR;
- document and screenshot detection;
- pet detection;
- face detection and person clustering;
- duplicate and near-duplicate analysis;
- semantic search;
- suggested albums and memories.

Disabling intelligence must not impair ordinary storage, browsing, export, or recovery.

## Phase 7 — Editing and Memories

**State:** Planned.

Implement and verify:

- non-destructive photo edits;
- edit history and synchronization;
- basic video trim/rotate/mute/color operations;
- memories;
- event clustering;
- generated collections;
- hide/dismiss controls for people, dates, and memory classes.

## Phase 8 — Advanced Privacy and Protected Media

**State:** Planned.

Develop and validate:

- protected media;
- reauthentication requirements;
- search, preview, widget, notification, and memory exclusion;
- sensitive metadata controls;
- secure sharing;
- Sealed Library design;
- client-controlled encryption only after the complete key, thumbnail, metadata, search, sharing, recovery, and succession lifecycle is solved.

## Phase 9 — Preservation, Portability, and Migration

**State:** Planned.

Complete and verify:

- selected and full-library export;
- portable machine-readable manifest;
- album and metadata reconstruction;
- migration from common photo-library exports;
- integrity auditing;
- tested clean-target restore;
- disaster recovery;
- rebuild procedures for derivatives and indexes;
- long-term migration;
- family succession.

## Phase 10 — Stable Acceptance

**State:** Planned.

Require exact evidence for:

- cross-platform behavior;
- accessibility;
- representative-device operation;
- performance and large-library behavior;
- security;
- privacy;
- recovery;
- upgrade and rollback;
- data portability;
- nine-system platform evaluation;
- Glaze UI 1.5.1 conformance;
- release artifact provenance;
- exact-release validation.

## MVP gate

The first usable MVP should include authentication, multi-user libraries, original storage, resumable upload, mobile camera backup, web/mobile timeline, desktop import, thumbnails, albums, favorites, basic metadata search, user-to-user sharing, trash/restore, backup-state indicators, offline/download support, import/export, Privacy Shield integration, Wardveil Security integration, and Everkeep recovery requirements.

Advanced intelligence is explicitly later than trusted storage, synchronization, privacy, security, and recovery foundations.
