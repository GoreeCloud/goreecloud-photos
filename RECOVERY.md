# GoreeCloud Photos — Recovery and Preservation Model

**Status:** Phase 0 recovery baseline  
**Implementation status:** Recovery requirements defined; backup and restore implementation not yet accepted.

## 1. Recovery objective

Photos protects information, not one deployment.

A successful backup job does not prove recovery. Recovery is established only when required information can be restored into an approved environment, validated, and returned to usable service.

## 2. Recovery classes

### Class A — Irreplaceable / authoritative

Must be backed up and restored:

- immutable original photos and videos;
- authoritative relational database state;
- ownership and library memberships;
- albums and memberships;
- user captions/descriptions and corrected metadata;
- non-destructive edit instructions;
- required sharing grants and revocation state;
- protected-media state;
- user-assigned person organization;
- required policy/security/privacy evidence references;
- authorized encryption/recovery material where such features exist.

### Class B — Valuable but reproducible

May be rebuilt when originals and authoritative metadata survive:

- extracted technical metadata;
- thumbnails;
- previews;
- optimized copies;
- transcodes;
- search indexes.

### Class C — Optional derived intelligence

Rebuild or omit according to authorization and policy:

- object/scene tags;
- OCR;
- face observations;
- embeddings;
- similarity indexes;
- quality assessments;
- generated memories.

User-authored names, exclusions, and merge/split decisions are not disposable merely because they relate to intelligence.

## 3. Recovery layers

The mature system should use complementary layers:

1. Primary authoritative storage.
2. Local/object-store integrity protection and versioning where supported.
3. Database-native backup and point-in-time recovery.
4. GoreeCloud Backup operational jobs.
5. Everkeep-governed off-environment recovery protection.
6. User export for portability and independent reconstruction.

Synchronization is not backup.

## 4. Original-media protection

Original media must have:

- content SHA-256;
- verified byte size;
- immutable storage semantics;
- periodic integrity sampling or full verification according to scale and policy;
- independent recovery protection;
- restore procedures that preserve the original payload.

A derivative is never an acceptable substitute for a lost original.

## 5. Database protection

The PostgreSQL recovery plan must include:

- consistent full backup;
- transaction/WAL protection sufficient for the approved recovery objective;
- schema migration history;
- version compatibility records;
- restoration into a clean database;
- relational integrity validation.

A database backup without matching original-media recovery is incomplete.

## 6. Restore order

Clean-target recovery order:

1. Establish trusted infrastructure and secrets access.
2. Restore or create PostgreSQL at a compatible version.
3. Apply the approved schema baseline/migrations as required.
4. Restore authoritative database backup.
5. Restore immutable original media.
6. Validate checksums and database-to-object references.
7. Restore required configuration and protected recovery material.
8. Start API in recovery-validation mode.
9. Rebuild derivatives, indexes, and optional intelligence according to policy.
10. Validate representative user/library/album/share relationships.
11. Run application-level recovery acceptance.
12. Only then declare the recovered service usable.

## 7. Integrity validation

Recovery validation must include, at minimum:

- asset count by scoped library;
- original-object reference completeness;
- checksum verification sample or full pass according to approved procedure;
- missing-object detection;
- orphan-object detection;
- album relationship validation;
- ownership and membership validation;
- share/revocation validation;
- edit-instruction readability;
- representative original download and decode;
- representative derivative rebuild;
- export generation from restored state.

## 8. Clean-target restore testing

A restore test must use a target that does not depend on hidden state from the live deployment.

The test record identifies:

- exact application revision;
- database version;
- backup/recovery point;
- media-store generation or backup identity;
- steps performed;
- validations performed;
- failures and remediation;
- final result.

No Production/Stable recovery claim is permitted without successful restore evidence appropriate to the exact release scope.

## 9. Deletion and backup retention

Deletion from active Photos storage does not imply immediate disappearance from all retained backups.

The UI and documentation must distinguish:

- active-library deletion;
- trash retention;
- primary-store purge;
- backup retention;
- recovery expiration.

Recovery systems must not silently resurrect a deleted item into ordinary active state during restore. Deletion tombstones or equivalent reconciliation data must be preserved for the required window.

## 10. Export is not backup

Portable export exists so the user can reconstruct and migrate data without Photos.

Full export should contain:

- original media;
- portable metadata;
- album relationships;
- user captions/descriptions;
- non-destructive edit information where portable;
- manifest with stable asset identity and checksums.

An export may complement recovery but does not replace operational backup of the running service.

## 11. Disaster recovery

Disaster recovery must account for failure of:

- the application host;
- PostgreSQL;
- primary media storage;
- credentials/secrets systems;
- network publication;
- worker system;
- an entire storage pool or administrative boundary.

A single reasonable failure must not destroy both the authoritative information and every required recovery path.

## 12. Sealed Library future boundary

Client-controlled encryption cannot be marked recoverable until the recovery lifecycle covers:

- keys;
- device loss;
- authorized recovery;
- encrypted previews;
- encrypted metadata;
- sharing;
- search;
- inheritance/succession;
- revocation.

Until then Sealed Library remains design work, not a recovery capability.
