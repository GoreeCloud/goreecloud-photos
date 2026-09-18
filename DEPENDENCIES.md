# GoreeCloud Photos — Dependency and Implementation Stack Policy

**Status:** Experimental dependency baseline  
**Runtime dependency state:** Go 1.27.1 is pinned. `github.com/jackc/pgx/v5` is approved and pinned at v5.11.0 for the PostgreSQL runtime boundary. No media-processing or intelligence dependency is yet accepted.

## 1. Governing rule

GoreeCloud controls the product architecture. Foundational third-party components may be used when they are narrower than the Photos product, justified, replaceable where practical, license-compatible, maintainable, and isolated behind GoreeCloud-owned boundaries.

No complete third-party photo application is an approved product foundation.

## 2. Selected core technologies

### Go 1.27.1

Selected for the initial server service plane.

Reasons:

- current GoreeCloud usage and CI precedent;
- strong standard HTTP/concurrency/runtime support;
- simple deployable binaries;
- good fit for API, synchronization, upload coordination, and workers;
- low runtime operational overhead.

The toolchain is pinned in source and verified in CI.

### PostgreSQL

Selected as the authoritative relational database and initial durable job queue.

Owns:

- domain metadata;
- authorization/membership relationships;
- upload session state;
- synchronization change log;
- durable worker jobs;
- idempotency records.

Using one database initially reduces dependency count and transaction-boundary complexity.

### pgx v5.11.0

**Source:** `github.com/jackc/pgx/v5`  
**Pinned version:** `v5.11.0`  
**License:** MIT  
**Dependency class:** Required runtime library for the PostgreSQL boundary  
**Update source:** upstream tagged pgx releases and Go module metadata

Role:

- PostgreSQL connection pooling;
- connectivity/readiness probing;
- database-backed Photos domain operations as they are introduced;
- PostgreSQL integration validation.

Boundary:

- GoreeCloud Photos owns database semantics, schema, migrations, transaction boundaries, authorization, and domain contracts.
- pgx does not define Photos data meaning or become a persistent data format.
- connection strings and credentials are configuration secrets and must not be logged.
- the adapter returns bounded errors rather than embedding credentials or raw connection strings in public status output.

Replacement path:

- PostgreSQL-specific code remains behind the GoreeCloud-owned `internal/database` boundary;
- persistent schema/data remains standard PostgreSQL state;
- replacing the Go driver must not change stable Asset or API semantics.

Security/update considerations:

- pgx is maintained as the current stable v5 line;
- updates require exact-version review, CI, migration compatibility checks, and security review before adoption;
- database transport/authentication requirements remain deployment policy and must not be weakened by application defaults.

### Storage-driver abstraction

Selected instead of coupling Photos to one object-storage product.

Required initial implementations:

- controlled filesystem backend for development/small deployments;
- S3-compatible backend for scalable object storage.

Specific S3-compatible server software is not selected by this document.

### TypeScript + React

Selected for the first-class web client.

The web client remains a presentation client of Photos APIs, not the source for wrapped native clients.

### Kotlin + Jetpack Compose

Selected for Mobile A/Android, using Android-native lifecycle, media, background work, notification, permission, accessibility, and storage APIs.

Shared GoreeCloud Android platform components should be used where applicable.

### Rust + GTK4

Selected target for the initial Linux graphical desktop client.

The target preserves a native Linux application model with direct desktop/file/accessibility integration.

### Mobile B

No platform-specific language/framework is selected until the target operating system is named by authoritative product scope. It must use that platform's native application architecture by default.

## 3. Local databases

Platform clients may use SQLite or a platform abstraction backed by SQLite-compatible semantics for:

- cache metadata;
- offline asset subsets;
- pending mutation queue;
- sync cursor;
- download/offline availability state.

Local databases are not independent authoritative libraries.

## 4. Initial dependency exclusions

The core architecture does not initially require:

- Redis;
- Kafka;
- NATS;
- RabbitMQ;
- Elasticsearch/OpenSearch;
- a vector database;
- a proprietary cloud AI API;
- a third-party hosted identity provider outside GoreeCloud Identity;
- GoreeCloud Drive as the Photos storage model.

A later dependency must be justified by measured requirements and documented before adoption.

## 5. Candidate foundational media tools

These are candidates, not implemented dependencies:

### FFmpeg / ffprobe

Potential role: video metadata, playback-compatible derivatives, thumbnails, and bounded transcode operations.

### libvips

Potential role: efficient image decode, resize, thumbnail, and derivative generation.

### ExifTool or equivalent audited metadata library

Potential role: broad metadata extraction for formats not adequately covered by native libraries.

Before adoption each candidate requires:

- exact source/version;
- license review;
- provenance;
- security/update process;
- parser threat-model review;
- sandbox/resource-limit design;
- supported-format scope;
- replacement strategy.

## 6. Intelligence dependencies

Photo intelligence is optional.

Models/runtimes must be:

- GoreeCloud-controlled or explicitly authorized;
- replaceable;
- privacy-scoped;
- versioned;
- independently disableable;
- isolated from original-media availability.

Embeddings or model-specific output formats must not become the only representation of user organization.

## 7. Dependency classes

### Critical

Failure threatens authoritative data or core service operation.

Initial examples once implemented:

- PostgreSQL;
- original-media storage.

### Required

Needed for a supported capability but not the preservation of all authoritative data.

Initial example:

- pgx PostgreSQL runtime library.

Potential later examples:

- media processor for derivative generation;
- platform-native client runtime.

### Optional

Capability enhancement that may fail without destroying core storage/browsing.

Examples:

- semantic-search model;
- landmark model;
- AI captioning.

### Build-only

Needed to build/test but not operate the deployed service.

Current example:

- official PostgreSQL 17.11 Bookworm image pinned by digest for CI integration validation.

Every material dependency must document class, owner, version/pin, update source, license, recovery/replacement path, and security/privacy implications.

## 8. Locking and provenance

- Go modules use `go.mod`/`go.sum` with exact direct dependency versions and a pinned Go toolchain.
- Web dependencies use a committed lockfile.
- Android dependencies use Gradle version controls and dependency locking where supported.
- Rust uses Cargo.lock for application deliverables.
- Container images are pinned by immutable digest where used.
- GitHub Actions use immutable commit SHAs.
- Dependency reports are generated as validation evidence where appropriate.

## 9. Replacement principle

Persistent user data contracts must not require one dependency forever.

Replacing a processor, search engine, object-store product, model runtime, or client framework must not require abandoning original media or user-authored metadata.

The safest dependency is one that can be upgraded, restored, or replaced without changing what the user's library means.
