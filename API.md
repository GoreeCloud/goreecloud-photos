# GoreeCloud Photos — API and Synchronization Contract

**Status:** Phase 0 API baseline  
**Base path:** /api/v1  
**Implementation status:** Contract defined; API runtime not yet implemented.

## 1. Principles

The API is:

- versioned;
- authenticated through GoreeCloud Identity inputs;
- authorized by Photos domain rules and applicable GoreeCloud Policy decisions;
- idempotent where retries are expected;
- cursor-paginated;
- conflict-aware;
- privacy-minimized;
- suitable for web and native clients.

Internal implementation details must not become permanent public contracts without need.

## 2. Authentication and authorization

Requests that require identity carry a GoreeCloud Identity-issued credential or session representation accepted by the server.

Identity establishes the actor. Photos remains authoritative for:

- library membership;
- asset ownership;
- album membership;
- share permissions;
- family-library contribution rules;
- resource-level authorization.

A detected person in an image is never treated as an authenticated user.

Authorization is evaluated again on every sensitive operation. Client-side visibility is not authorization.

## 3. Request identity and retries

Every response includes a request identifier.

Mutating requests that can be safely retried accept an Idempotency-Key.

The server stores the bounded result of an idempotent operation with a request fingerprint and rejects reuse of one key for a materially different request.

Expected retry-safe operations include:

- creating upload sessions;
- completing upload sessions;
- creating albums;
- applying metadata mutations;
- trash/restore transitions;
- share creation/revocation where the request identity is stable.

## 4. Revisions and conditional mutation

Mutable resources expose a revision.

Clients send the expected revision using If-Match or an equivalent contract field.

If the resource changed and the operation cannot be safely merged, the server returns HTTP 409 with:

- resource identifier;
- expected revision;
- current revision;
- conflict class;
- recovery action.

Last-write-wins is not the general conflict model.

## 5. Error envelope

Error responses use a stable structure:

- code
- message
- request_id
- details when safe
- retryable
- retry_after when applicable

Messages must not leak filesystem paths, secrets, share tokens, raw policy internals, or private media metadata not already authorized for the actor.

## 6. Pagination

Collection APIs use opaque cursor pagination.

A cursor is scoped to the actor, library, query, and ordering context as needed. Clients must treat it as opaque.

Offset pagination is not used for the primary large-library timeline because inserts and deletes make stable traversal difficult.

## 7. Resumable upload protocol

### Create session

POST /api/v1/libraries/{library_id}/uploads

Request declares:

- original filename
- media type when known
- expected byte size
- optional client SHA-256
- capture metadata hints
- device identifier

Response declares:

- upload_id
- negotiated part size
- expiration
- accepted transfer method
- current received state

### Upload part

PUT /api/v1/uploads/{upload_id}/parts/{part_number}

Each part is independently retryable. The server records received length and checksum evidence.

A repeated identical part is accepted idempotently. A conflicting replacement for an already committed part is rejected unless the protocol explicitly resets that part.

### Inspect session

GET /api/v1/uploads/{upload_id}

Returns received parts/ranges and session state without exposing unnecessary media content.

### Complete session

POST /api/v1/uploads/{upload_id}/complete

Completion:

1. verifies all required bytes are present;
2. calculates the server-authoritative SHA-256;
3. compares the client checksum when supplied;
4. performs required ingestion validation;
5. durably stores the immutable original;
6. atomically creates the Asset and required metadata;
7. records the synchronization change;
8. returns the new asset and revision.

The client must not show verified remote backup before successful completion.

### Cancel session

DELETE /api/v1/uploads/{upload_id}

Cancels incomplete state and schedules staging cleanup.

## 8. Core resource endpoints

Initial API families:

- /libraries
- /libraries/{library_id}/assets
- /assets/{asset_id}
- /albums
- /albums/{album_id}
- /shares
- /share-links
- /uploads
- /sync
- /devices
- /exports

Future People, Memories, Intelligence, and advanced Search APIs remain separate versioned capability surfaces.

## 9. Asset lifecycle operations

Archive, trash, restore, and purge are explicit transitions rather than ambiguous generic deletion.

Examples:

- POST /assets/{asset_id}/archive
- POST /assets/{asset_id}/unarchive
- POST /assets/{asset_id}/trash
- POST /assets/{asset_id}/restore
- DELETE /assets/{asset_id} for authorized permanent purge requests

Permanent deletion may become asynchronous. A successful request can mean deletion has been accepted, not that all backup retention has instantly expired.

## 10. Synchronization

### Change feed

GET /api/v1/libraries/{library_id}/sync/changes?cursor={cursor}

Returns an ordered page of changes plus next_cursor.

Each change includes:

- sequence or opaque change identity;
- resource_type;
- resource_id;
- operation;
- revision;
- changed_at.

Clients fetch current resource state as needed.

### Cursor behavior

- cursors are monotonic for one scoped change stream;
- a client persists the last committed cursor only after applying the page;
- expired/invalid cursors return an explicit resync-required error;
- resync uses paginated current state plus a new cursor rather than downloading one unbounded payload.

### Tombstones

Deletes and purges produce tombstones long enough for active clients to observe the transition according to retention policy.

## 11. Offline mutation queue

Native clients may queue authorized mutations offline.

Each queued mutation includes:

- client operation ID;
- resource ID;
- expected revision;
- operation type;
- bounded payload;
- created time.

On reconnect, the client submits in order where dependencies require it. The server applies normal idempotency and conflict rules.

## 12. Conflict rules

### Safely mergeable candidates

Only explicitly documented independent fields may be auto-merged, such as adding different assets to one album when ordering is not conflicting.

### Explicit conflict required

Examples:

- two edits changing the same caption from one base revision;
- conflicting edit-recipe changes;
- ownership changes;
- simultaneous protected-state changes;
- incompatible album ordering changes;
- share permission escalation;
- delete versus edit.

A conflict response must preserve both user intents long enough for resolution when practical.

## 13. Search

Initial structured search is server-side and uses authoritative metadata plus rebuildable indexes.

Search queries must remain scoped by authorization before result ranking.

Optional semantic search must be a separately authorized capability and must not be required for ordinary date, album, filename, metadata, or type search.

## 14. Export

Export requests are jobs.

POST /api/v1/exports creates an export with:

- scope;
- format;
- inclusion options;
- authorization snapshot reference;
- expiration policy.

Export artifacts are temporary, access-controlled, integrity-checked, and removable after expiry.

## 15. API compatibility

Within v1:

- additive response fields are permitted;
- clients must ignore unknown fields where safe;
- meaning of existing fields must not silently change;
- enum extension requires clients to handle unknown values safely;
- breaking semantic changes require a new API version or a migration mechanism.

## 16. Security and privacy

API logs must not contain raw media bodies, reusable share secrets, authentication credentials, face embeddings, or full sensitive metadata by default.

Request identifiers, bounded error codes, durations, byte counts, and non-sensitive operational fields are preferred for diagnostics.

Missing required Identity, Policy, Privacy Shield, or Wardveil evidence for a sensitive action produces a non-passing result rather than optimistic access.
