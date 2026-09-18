# GoreeCloud Photos — Security

**Security status:** No Photos runtime security acceptance exists.

## Current boundary

This repository now contains an Experimental Go service foundation and filesystem original-media storage adapter. The service exposes only bounded health/readiness endpoints; it does not expose a media-ingestion HTTP API, authentication service, Photos authorization, sharing service, user-facing client, or release artifact.

The filesystem adapter's local integrity/immutability tests are not Wardveil acceptance. No statement in this repository should be interpreted as proof that personal media is currently protected by Wardveil Security or by a production-ready Photos implementation.

## Planned security requirements

Future implementation must address, as applicable:

- authenticated and authorized access;
- device/session management;
- encrypted transport;
- protected secrets;
- per-user and family-library isolation;
- upload validation and malformed-file handling;
- safe media processing;
- share-link authorization and revocation;
- rate limiting and abuse controls;
- administrative-action protection;
- secure metadata handling;
- integrity verification;
- backup and recovery security;
- audit and security evidence.

## Wardveil Security

Wardveil Security is the planned security authority for applicable Photos ingestion and access boundaries. Application-specific Wardveil runtime integration and acceptance are currently blocked and unverified.

Do not display or document a “protected” state unless current evidence from the appropriate authoritative producer establishes it.

## Reporting security issues

Do not publish credentials, private media, private metadata, exploit details, tokens, keys, recovery material, or other sensitive information in a public issue.

Use the repository's approved private vulnerability-reporting path when available, or the established private GoreeCloud security process.

## Sealed Library

Sealed Library is a future design concept, not a current end-to-end encryption claim. No such claim may be made until client encryption, keys, metadata, previews, search, sharing, recovery, and succession have all been designed, implemented, tested, and accepted.
