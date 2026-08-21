# GoreeCloud Photos architecture direction

## Development model

GoreeCloud Photos begins from an Immich foundation and immediately enters a controlled fork-to-native transition. The objective is not a permanent visual rebrand. GoreeCloud will introduce first-party boundaries so individual inherited capabilities can be replaced safely over time.

## Keepsake capability umbrella

**Keepsake** is the first-party capability umbrella for GoreeCloud Photos. The preferred expression is **GoreeCloud Photos — powered by Keepsake**.

Keepsake is an architectural and product feature-family identity, not a replacement for the GoreeCloud Photos application name. New GoreeCloud-owned feature surfaces and domain boundaries should use Keepsake naming where appropriate while inherited Immich internals remain explicitly transitional.

## Capability domains

The target architecture separates the product into independently evolvable domains beneath Keepsake:

- **Keepsake Core** — users, libraries, assets, metadata, albums, favorites, archive, Trash, permissions, and primary application behavior.
- **Keepsake Vault** — authoritative original-media storage, storage abstraction, integrity state, media placement, and durable-file access.
- **Keepsake Sync** — device backup, upload queues, resumable transfers, duplicate detection, synchronization state, retry behavior, and bandwidth controls.
- **Keepsake Memories** — anniversaries, on-this-day experiences, trips, contextual collections, and rediscovery.
- **Keepsake Search** — metadata, filename, date, place, indexed, semantic, and visual search.
- **Keepsake Vision** — local facial recognition, object recognition, OCR, embeddings, classification, and similarity processing.
- **Keepsake Studio** — non-destructive editing, crop, rotation, adjustments, transforms, video-editing foundations, and restoration of originals.
- **Keepsake Share** — albums, family sharing, partner sharing, controlled links, expiration, download permissions, and sharing policy.
- **Keepsake Places** — map browsing, locations, geographic search, trips, and clusters.
- **Keepsake Import** — Google Photos Takeout, Ente exports, Immich data, filesystem trees, camera media, device media directories, supported archives, and future Photos exports.
- **Keepsake Export** — portable export of originals and associated authoritative metadata.
- **Keepsake Recovery** — integrity validation, rebuild tooling, backup verification, restoration, and reconstruction support.
- **Keepsake Clients** — web, Android, Linux desktop, CLI, and future iOS clients.

## Data authority

The application database describes and indexes media but does not conceptually own original files.

The target storage model separates authoritative originals from replaceable application state. Examples of durable media roots include:

```text
personal/<user>/photos
personal/<user>/videos
family/photos
family/videos
archives/
```

PostgreSQL metadata, thumbnails, transcoded derivatives, caches, search indexes, processing workspaces, and ML embeddings must be classified according to whether they are authoritative, reconstructable, or temporary.

A loss of the Photos application runtime must not make original photos and videos inaccessible or unintelligible.

## Native-boundary rule

New GoreeCloud-specific features should prefer GoreeCloud-owned contracts rather than direct coupling to inherited implementation details. Where practical, callers should depend on a GoreeCloud domain interface whose implementation can initially delegate to Immich-derived code and later be replaced by a native implementation.

The first boundaries to formalize are:

1. Keepsake Vault storage and original-media authority;
2. Keepsake Import and migration;
3. Keepsake Export and portable transfer;
4. Keepsake Sync and backup state;
5. authorization and personal/shared library boundaries;
6. Keepsake Vision and local metadata-processing policy.

## Migration requirements

Import workflows should be repeatable and idempotent. Re-running the same source import should not silently duplicate every asset.

Where source information is available, migrations should preserve original files, capture dates, timezone information, GPS, descriptions, filenames, EXIF metadata, motion-photo relationships, sidecars, album membership, favorites, edit provenance, and source provenance.

## Privacy and AI

Facial recognition, object recognition, OCR, semantic embeddings, similarity processing, and related privacy-sensitive analysis should run locally whenever practical. External AI services must not become a mandatory dependency for ordinary search or organization.

## Deployment boundary

The intended long-term deployment is a family-facing Photos application runtime with durable media on GoreeCloud-controlled storage. Backend application ports should not be exposed directly to the public Internet merely to provide ordinary family access.

## Transition phases

### Phase 1 — Maintained-fork foundation

Record provenance, validate the inherited build, establish GoreeCloud governance, document architecture, define security/privacy boundaries, and prepare controlled product identity work.

### Phase 2 — GoreeCloud domain contracts

Introduce first-party Keepsake storage, import/export, authorization, synchronization, API, and other domain contracts.

### Phase 3 — Native server replacement

Replace selected inherited server capabilities one domain at a time while preserving migrations and compatibility.

### Phase 4 — Native web replacement

Progressively replace inherited web architecture and components with GoreeCloud-owned implementations.

### Phase 5 — Native mobile replacement

Replace inherited mobile components while maintaining dependable background backup, browsing, sharing, and offline behavior.

### Phase 6 — Immich independence

Remove final Immich-specific internal dependencies once equivalent GoreeCloud-owned implementations are validated.
