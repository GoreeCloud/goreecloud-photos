# GoreeCloud Photos — Data Model

**Status:** Phase 0 domain-model baseline  
**Implementation status:** Contract defined; persistent implementation not yet established.

## 1. Identity and identifier rules

Every durable domain object receives a stable identifier independent from filenames, storage paths, titles, or mutable metadata.

The initial identifier format is UUIDv7 for newly created GoreeCloud Photos domain records.

Identifiers must not encode:

- a username;
- an email address;
- a storage path;
- a device serial number;
- a geographic location;
- a media filename.

The server is authoritative for server-domain identifiers.

## 2. Library

A Library is the top-level Photos ownership and visibility boundary.

Fields include:

- library_id
- library_type: personal or family
- owner_subject_id
- created_at
- updated_at
- revision
- lifecycle_state

A personal library has one controlling owner. A family library may have multiple members, but individual asset ownership remains explicit.

## 3. Library Membership

LibraryMembership records who may access a shared/family library and with which role.

Roles initially include:

- owner
- manager
- contributor
- viewer

Identity authenticates the actor. Photos evaluates library and asset-domain authorization using authenticated Identity inputs plus Photos membership/grant state and applicable Policy decisions.

## 4. Asset

Asset is the stable user-facing media identity.

Core fields:

- asset_id
- library_id
- owner_subject_id
- original_object_id
- media_type
- original_filename
- content_sha256
- byte_size
- capture_time
- capture_time_zone
- imported_at
- updated_at
- revision
- favorite
- lifecycle_state
- protection_state
- processing_state
- server_storage_state

Asset IDs never depend on file paths.

### lifecycle_state

- active
- archived
- trash
- purged

Purged is a terminal logical state. Physical destruction may follow governed retention and recovery rules.

### protection_state

- standard
- protected

Additional protected-media controls remain a later feature and must not be inferred from this field alone.

### processing_state

- pending
- ready
- degraded
- failed

### server_storage_state

- staging
- verified
- unavailable
- corrupt

A verified server-storage state means the primary server copy has passed the Photos ingestion invariant. It does not by itself prove independent backup or complete Everkeep recovery readiness.

## 5. Original Object

OriginalObject is an immutable media blob record.

Fields include:

- original_object_id
- storage_backend
- storage_key
- content_sha256
- byte_size
- media_type
- created_at
- integrity_verified_at
- storage_generation where supported

The object payload is never edited in place.

Exact duplicate storage reuse may occur only inside an authorized ownership/storage domain after exact checksum verification. Cross-user deduplication is not an initial architecture feature and must not expose whether another user possesses the same content.

## 6. Device Asset State

DeviceAssetState represents device-specific presence and backup workflow without polluting the global Asset with one device's local state.

Fields include:

- device_id
- asset_id
- local_presence
- local_quality
- upload_state
- last_verified_server_state
- last_sync_cursor
- updated_at

upload_state may include:

- device_only
- waiting
- uploading
- verifying
- verified_remote
- paused
- failed

## 7. Metadata

AssetMetadata separates technical extraction from user-authored information.

Authoritative user-authored fields include:

- caption
- description
- rating where supported
- corrected_capture_time
- corrected_time_zone
- corrected_location
- hidden_location preference

Extracted technical metadata may include:

- dimensions
- duration
- orientation
- camera make/model
- lens information
- exposure information
- embedded capture timestamp
- embedded location

The original embedded metadata remains preserved in the original object even when a user overrides a displayed value.

## 8. Album

Album fields include:

- album_id
- library_id
- owner_subject_id
- album_type
- title
- description
- cover_asset_id
- sort_mode
- created_at
- updated_at
- revision

AlbumMembership contains asset_id, album_id, ordering position, and membership metadata.

Removing AlbumMembership never deletes the Asset.

## 9. Edit Revision

Photo/video editing is non-destructive.

EditRevision fields include:

- edit_revision_id
- asset_id
- parent_revision_id
- editor_subject_id
- instruction_format_version
- instructions
- created_at

Rendered results are derivatives. Reverting to original means selecting no active edit revision, not rewriting the original.

## 10. Derived Media

DerivedMedia describes rebuildable outputs:

- derivative_id
- asset_id
- derivative_type
- source_edit_revision_id
- processor_version
- storage_key
- byte_size
- checksum
- created_at

Derivative types may include thumbnail, preview, optimized, transcode, and rendered_edit.

## 11. Sharing

### Share Grant

A ShareGrant provides an authenticated GoreeCloud subject with explicit access.

Fields include:

- share_grant_id
- resource_type
- resource_id
- grantee_subject_id
- permission
- expires_at
- created_by
- revoked_at

### Share Link

A ShareLink represents an externally usable capability.

Fields include:

- share_link_id
- resource_type
- resource_id
- token_verifier
- password_verifier when enabled
- allow_download
- expires_at
- revoked_at
- created_by

Raw reusable link secrets must not be persisted in ordinary database records after issuance.

## 12. Trash and deletion

Moving an Asset to trash creates a synchronization-visible state transition and retention deadline.

Permanent deletion requires:

1. authorization;
2. retention/policy evaluation;
3. removal of active sharing grants;
4. logical purge/tombstone commit;
5. physical media cleanup according to reference and recovery policy;
6. deletion/revocation of derived intelligence and indexes where required.

Backups may retain deleted data for a documented period. User-facing deletion claims must distinguish active-store deletion from recovery retention.

## 13. Upload Session

UploadSession tracks resumable transfer:

- upload_id
- actor_subject_id
- library_id
- expected_size
- expected_sha256 when supplied
- received_bytes
- part_size
- state
- expires_at
- created_at
- updated_at

States:

- open
- receiving
- verifying
- complete
- failed
- expired
- cancelled

Completion creates an Asset only after ingestion validation and authoritative commit.

## 14. Sync Change

SyncChange is an ordered change record:

- sequence
- library_id
- resource_type
- resource_id
- operation
- revision
- changed_at

Operations:

- create
- update
- delete
- restore

The synchronization cursor is opaque to clients even if the server internally derives it from sequence state.

## 15. Job

Job is durable worker orchestration state.

Fields include:

- job_id
- job_type
- resource_id
- state
- attempt
- available_at
- lease_owner
- lease_expires_at
- last_error_code
- created_at
- updated_at

Job payloads must minimize user content. Workers retrieve authoritative data by scoped identifiers when needed.

## 16. Intelligence records

Intelligence-derived data is logically separate.

Examples:

- FaceObservation
- PersonCluster
- OCRResult
- ObjectTag
- SceneTag
- SimilarityVector
- QualityAssessment

User-assigned person names and merge/split decisions are distinct from raw model outputs because they are user-authored organization state.

Face embeddings, similarity vectors, recognized text, and location-derived intelligence are sensitive data subject to Privacy Shield controls and retention/deletion rules.

## 17. Conflict and revision model

Mutable resources carry a monotonically changing revision value.

Clients submit the revision they edited against. The server:

- accepts when the expected revision matches;
- performs explicit field-level merge only for classes documented as safely mergeable;
- otherwise returns a conflict requiring client resolution.

Body-like edit recipes, ownership changes, sharing changes, protected-media state, and destructive operations must not silently use last-write-wins.

## 18. Rebuildability classes

### Irreplaceable or authoritative

- original media
- ownership and library membership
- user captions/descriptions
- album relationships
- edit instructions
- explicit sharing state
- user-assigned person organization
- policy-required audit/evidence references
- required recovery material

### Rebuildable

- thumbnails
- previews
- optimized copies
- search indexes
- extracted technical metadata where recoverable from originals
- machine-generated intelligence where reprocessing is permitted

Rebuildable does not mean free to lose; rebuild cost and user impact still require operational planning.
