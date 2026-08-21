# GoreeCloud Photos product boundary

## Role

GoreeCloud Photos is the self-hosted photo and video platform for backup, synchronization, organization, search, memories, editing, sharing, migration, preservation, recovery, and access across devices.

It is intended to become a privacy-first alternative to Google Photos and Immich while keeping authoritative original media under GoreeCloud-controlled storage and maintaining practical export and recovery paths.

## Keepsake feature identity

**Keepsake** is the official first-party feature umbrella for GoreeCloud Photos.

The preferred product expression is **GoreeCloud Photos — powered by Keepsake**. Keepsake names the coherent family of GoreeCloud-owned photo and video capabilities rather than a single subsystem. It represents the product's emphasis on preserving meaningful personal and family media while keeping originals private, portable, recoverable, and under user control.

The existing capability domains remain useful architectural names beneath the Keepsake umbrella:

- Keepsake Core — users, libraries, assets, metadata, albums, favorites, archive state, Trash, permissions, and primary Photos behavior;
- Keepsake Vault — authoritative original-media storage, storage abstraction, integrity, placement, lifecycle handling, and durable-file access;
- Keepsake Sync — device backup, upload queues, resumable transfers, duplicate detection, synchronization state, retry behavior, and ingestion;
- Keepsake Memories — anniversaries, on-this-day experiences, trips, contextual collections, people-oriented memories, and rediscovery;
- Keepsake Search — metadata, filename, date, location, indexed, semantic, and visual discovery;
- Keepsake Vision — local facial recognition, object recognition, embeddings, OCR, classification, similarity analysis, and other privacy-sensitive media intelligence;
- Keepsake Studio — non-destructive editing, adjustments, transformations, video-editing foundations, edit history, and original restoration;
- Keepsake Share — albums, family and partner sharing, controlled links, expiration, download permissions, and sharing-policy enforcement;
- Keepsake Places — maps, geographic search, locations, trips, clusters, and location-oriented organization;
- Keepsake Import — Google Photos Takeout, Ente, Immich, filesystem, camera, removable-media, phone-media, archive, and portable-export ingestion;
- Keepsake Export — portable original-media and metadata export independent of continued Photos operation;
- Keepsake Recovery — integrity validation, rebuild tooling, backup verification, database/index reconstruction support, and restoration workflows;
- Keepsake Clients — web, Android, Linux desktop, command-line tooling, and future iOS access.

Keepsake does not replace the GoreeCloud Photos product name, Glaze UI, Wardveil Security, Privacy Shield, or Everkeep. Photos remains the application; Keepsake is its feature family; Glaze UI is its design language; Wardveil Security and Privacy Shield provide shared security and privacy foundations; Everkeep remains the broader GoreeCloud preservation and recovery identity where applicable.

## Relationship to GoreeCloud Gallery

GoreeCloud Photos and GoreeCloud Gallery are separate, complementary applications.

### GoreeCloud Gallery

Repository: `GoreeCloud/goreecloud-gallery`

Gallery is the offline-first Android application for browsing and managing media stored locally on a device. Its useful core behavior must not require a Photos server, network access, cloud account, or synchronized remote library.

Gallery must not be deleted, overwritten, or repurposed to create Photos.

### GoreeCloud Photos

Repository: `GoreeCloud/goreecloud-photos`

Photos owns server-backed capabilities such as:

- automatic device backup;
- synchronized photo and video libraries;
- albums and shared libraries;
- memories and rediscovery;
- people, object, OCR, and semantic discovery using local processing where practical;
- map and place browsing;
- cross-device access;
- server-side media processing;
- migration and import workflows;
- portable export and recovery.

## Optional interoperability

Gallery and Photos may integrate through explicit APIs and Android platform mechanisms without collapsing their product boundaries. Examples include:

- backup-state indicators in Gallery;
- `Back up to Photos` actions;
- `View in Photos` links;
- downloading Photos assets for offline access;
- opening downloaded media in Gallery;
- saving a Photos original to device-local storage.

Gallery's local-media authority and offline behavior must remain intact when Photos is unavailable.

## Product principles

1. Original media belongs to the user and family, not to the application database.
2. Originals must remain portable and independently recoverable.
3. Privacy-sensitive media analysis should run locally whenever practical.
4. Thumbnails, caches, derived media, search indexes, and ML embeddings should be rebuildable where practical.
5. Import and export are first-class capabilities.
6. Personal libraries remain private unless intentionally shared.
7. No essential capability should permanently depend on a commercial cloud provider.
8. GoreeCloud-owned interfaces should progressively isolate and replace inherited Immich internals.
9. User-facing GoreeCloud-controlled surfaces should follow Glaze UI.
10. Security, privacy, resilience, and recovery requirements should align with Wardveil Security, Privacy Shield, and Everkeep.
