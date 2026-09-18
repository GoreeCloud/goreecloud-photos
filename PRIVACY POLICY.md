# GoreeCloud Photos — Privacy Policy

**Status:** Experimental privacy boundary  
**Runtime status:** Experimental engineering service exists; no supported photo-library runtime or Privacy Shield acceptance is verified.

## Current data handling

The current Experimental service exposes bounded health/readiness endpoints and includes a filesystem storage adapter exercised by tests. It does not expose a media-upload API, authentication, indexing, intelligence, synchronization, sharing, or user-facing library workflow.

Accordingly, this repository does not establish a current claim that GoreeCloud Photos is operating a user photo/video library or that Privacy Shield integration has been accepted.

## Planned privacy principles

Future Photos implementations must be private by default and must minimize access to:

- photographs and videos;
- face and person-recognition data;
- location;
- captions and descriptions;
- camera and file metadata;
- search history and indexes;
- generated intelligence;
- sharing relationships;
- device media libraries.

## Privacy Shield

Privacy Shield is the intended GoreeCloud privacy authority for Photos. Photos-specific Privacy Shield runtime integration and acceptance are not currently implemented or verified.

Privacy-sensitive operations should be able to explain:

- what information is processed;
- why processing is needed;
- where processing occurs;
- whether information is retained;
- who can access it;
- how authorization can be revoked;
- what derived information has been created.

Where required authorization cannot be established, sensitive processing should fail closed.

## Intelligence

Face recognition, person clustering, OCR, semantic search, object recognition, scene recognition, landmark recognition, and AI assistance must remain optional where specified.

Disabling optional intelligence must not remove ordinary access to stored original media.

## Sharing

New media, albums, identities, location, and derived intelligence remain private unless deliberately shared. External sharing should support metadata reduction where applicable, revocation, and explicit download permissions.

## Export and deletion

Future implementations must provide documented export and deletion behavior. Deleting a derived index or intelligence record should not require deleting the original media unless the user explicitly requests deletion of the media itself.

## Changes

This policy must continue to be revised as implementation adds real media-processing paths so it describes verified behavior rather than planned behavior.
