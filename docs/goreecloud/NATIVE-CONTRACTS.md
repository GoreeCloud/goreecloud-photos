# GoreeCloud Photos — Native Contracts

Status: Initial first-party contract foundation

## Purpose

This record establishes the first GoreeCloud-owned behavioral contracts that may be implemented behind the Immich-derived runtime during the controlled fork-to-native transition.

These contracts describe GoreeCloud requirements rather than copying upstream implementation details.

## Contract 1 — Media authority

**Photos Vault owns the media-authority contract.**

An original asset is authoritative user-owned media. The application may create derived representations, but it must distinguish originals from reconstructable state.

Required properties:

- stable internal asset identity;
- recorded owner identity;
- original-file location abstraction rather than hard-coded storage assumptions;
- media checksum or equivalent integrity evidence;
- original byte length;
- media type;
- captured/imported timestamps when known;
- provenance describing how the asset entered Photos;
- deletion state separate from physical destruction where recovery semantics require it;
- exportability without requiring proprietary GoreeCloud-only formats.

The contract must permit storage backends to evolve without redefining the user-visible concept of an asset.

## Contract 2 — Derived-state separation

The following are supporting or reconstructable state unless explicitly promoted by a future requirement:

- thumbnails;
- transcoded playback copies;
- preview images;
- machine-learning embeddings;
- face/object indexes;
- OCR indexes;
- search indexes;
- caches;
- generated memories;
- temporary import staging.

Loss of reconstructable state should not be equivalent to loss of the authoritative original.

## Contract 3 — Import provenance

**Photos Import owns import provenance.**

Every importer should be able to record a normalized provenance envelope containing, where applicable:

- source type;
- importer version;
- original source path/name;
- source-side identifier;
- source-side timestamps;
- source-side album or grouping context;
- source metadata sidecars;
- checksum evidence;
- import time;
- warnings or normalization decisions.

Initial source families include Google Photos Takeout, Ente exports, Immich-compatible exports, filesystem media, camera/removable-media sources, and phone media directories.

Import provenance must not require preserving source-provider credentials.

## Contract 4 — Portable export

**Photos Export owns portability.**

A portable export must be understandable outside a running GoreeCloud Photos instance.

At minimum, export design must support:

- original media;
- portable metadata;
- ownership/context needed to reconstruct the library;
- album/group relationships where selected;
- timestamps and provenance where selected;
- checksums or equivalent integrity evidence;
- a documented manifest format;
- deterministic or clearly documented path naming;
- explicit handling of unsupported or missing metadata.

A database dump alone is not considered a complete portable user export.

## Contract 5 — Sync state

**Photos Sync owns synchronization state, not the original asset itself.**

Sync state should be representable independently from the media authority record. A client may track:

- server asset identity;
- client-local identity;
- checksum/version evidence;
- upload/download state;
- last successful reconciliation;
- conflict state;
- requested deletion/tombstone state;
- retry/error state;
- bandwidth/power policy state.

A synchronization failure must not silently redefine asset ownership or authorize destructive conflict resolution.

## Contract 6 — Authorization

Every media operation must be attributable to an authorization context.

Authorization should be capable of answering:

- who is acting;
- which asset/library/share is targeted;
- which action is requested;
- whether the actor owns or is delegated access;
- whether the operation is administrative;
- whether the action crosses a sharing/publication boundary;
- whether destructive action is permitted;
- whether background processing is operating under an appropriately bounded service authority.

Authentication implementation and authorization policy are separate concerns.

## Contract 7 — Deletion and recovery

Deletion is a lifecycle operation, not merely a file-system call.

The long-term contract should distinguish:

1. user removal request;
2. recoverable Trash/tombstone state;
3. synchronization propagation;
4. retention/recovery boundary;
5. final application-level purge;
6. independent backup/recovery copies that follow their own retention policy.

A purge must not imply that separately governed backups are rewritten outside their documented retention rules.

## Contract 8 — Gallery interoperability

GoreeCloud Gallery remains independently useful without Photos.

Optional interoperability may include:

- backup-state queries;
- explicit upload/backup requests;
- deep links or View in Photos actions;
- album handoff;
- download-original requests;
- save-to-device actions.

No Photos integration may require Gallery to surrender its offline-first role or direct Android local-media authority.

## Contract versioning

These contracts begin at **Photos native contract set 0.1**.

Changes that alter ownership, destructive behavior, portability, authorization, or migration semantics require explicit review and migration consideration.

## Implementation rule

The first implementation should use adapters around inherited behavior where that reduces risk. The contract is the GoreeCloud authority; the inherited implementation is a transitional provider behind that authority until it is independently replaced.
