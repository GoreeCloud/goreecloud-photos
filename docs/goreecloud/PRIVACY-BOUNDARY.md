# GoreeCloud Photos — Privacy and External-Service Boundary

Status: Active architecture record

## Purpose

GoreeCloud Photos is a privacy-first, self-hosted photo and video platform. This record defines the default boundary for media, metadata, derived intelligence, network access, telemetry, and third-party services while the Immich-derived fork is transitioned toward first-party GoreeCloud architecture.

## Default rule

Photo and video content, thumbnails, metadata, face/person information, semantic embeddings, OCR output, object classifications, location history, album membership, sharing relationships, search history, and other library-derived information remain inside GoreeCloud-controlled infrastructure unless a user explicitly invokes a documented feature that requires a different destination.

No external service becomes an implicit authority merely because the inherited Immich implementation supports or references it.

## Local-first processing

The preferred architecture processes privacy-sensitive media intelligence locally, including:

- face detection and recognition;
- person clustering;
- object recognition;
- semantic embeddings and similarity search;
- OCR;
- duplicate and near-duplicate analysis;
- media classification;
- thumbnail and preview generation;
- metadata extraction;
- memories generation;
- library indexing.

A future optional external processor must be separately documented, disabled by default where practical, explicit about transmitted data, and incapable of silently becoming required for core library operation.

## Telemetry and analytics

GoreeCloud Photos must not require product analytics, advertising identifiers, behavioral tracking, marketing telemetry, or third-party usage profiling for normal operation.

Operational observability must be purpose-limited. Logs, metrics, traces, crash evidence, and audit records should minimize personal media information and avoid recording reusable credentials or unnecessary user content.

Inherited analytics, feedback, hosted-service, sponsorship, marketing, or upstream project reporting features are candidates for retirement rather than automatic recreation.

## External network access

Outbound network access must have a documented role. Examples that may be justified after review include:

- explicitly requested software/update metadata;
- standards-based time or certificate functions provided by the operating environment;
- user-requested remote imports;
- an explicitly configured geocoding or map provider when no approved local alternative is active;
- federated or sharing functions intentionally enabled by the administrator;
- package/build dependency retrieval in development and CI.

Core library browsing, original-media access, local search, album organization, and local AI should not require a commercial cloud-photo provider.

## Location and maps

Location information is sensitive media metadata. Photos Places should prefer GoreeCloud-controlled processing and caching.

If an external map tile, reverse-geocoding, or geocoding provider is used during transition, the integration must be treated as a replaceable adapter with documented data disclosure, caching behavior, rate limits, privacy implications, and an eventual local/self-hosted option where technically practical.

## Sharing

Sharing must be explicit and authorization-controlled. A share link or federated exchange must not grant broader library access than intended.

Public sharing is a publication action, not the default storage state. Private Photos deployment does not imply that every generated share endpoint should be publicly reachable.

## Originals and derived data

Original media is authoritative user-owned information. Application databases, indexes, thumbnails, embeddings, caches, and generated previews are supporting state.

The architecture must preserve the ability to recover and export originals independently of reconstructable application state wherever practical.

Destructive operations require clear authority boundaries so deletion of application metadata cannot unintentionally destroy independently protected originals or recovery copies.

## Authentication and authorization

The inherited authentication implementation may remain during transition, but it is not the long-term architectural authority solely because it came from Immich.

GoreeCloud Photos will progressively define its own documented authorization contracts and may integrate with the approved GoreeCloud Identity architecture when that integration is ready and justified.

User isolation, album/share permissions, administrative capabilities, media ownership, API authorization, and background-worker access must be independently testable.

## Secrets

API keys, signing material, tokens, database credentials, storage credentials, encryption material, OAuth secrets, and external-service credentials must not be stored in ordinary repository documentation or source files.

Repository examples use placeholders only.

## Fork-transition rule

Inherited functionality is classified as one of:

- retained temporarily for compatibility;
- retained as a justified mature dependency;
- replaced with a first-party GoreeCloud capability;
- retired because it conflicts with Photos scope or privacy goals.

Visible feature parity is not sufficient. A first-party replacement must preserve required security, authorization, migration, data-integrity, recovery, and privacy behavior.

## Production boundary

This record does not claim that the inherited Immich codebase already satisfies every GoreeCloud privacy requirement. It establishes the target architecture and review criteria.

No production migration or runtime cutover is authorized by this document.
