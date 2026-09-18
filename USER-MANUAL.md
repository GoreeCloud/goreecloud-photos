# GoreeCloud Photos — User Manual

**Current product availability:** No supported user-facing release exists.

## What can be used today

Nothing in this repository currently provides a supported GoreeCloud Photos client or server.

The repository has an Experimental engineering server foundation, but there is still no supported GoreeCloud Photos client or ordinary photo-library service for users.

## Important backup warning

Do **not** delete photographs or videos from a device on the assumption that GoreeCloud Photos has backed them up. No verified user-facing Photos backup workflow exists. The Experimental server does not expose a media-upload API and is intentionally not ready for ordinary use.

Do not treat repository documentation, screenshots, planned status indicators, or future architecture as proof that any personal media is stored, protected, synchronized, encrypted, recoverable, or available elsewhere.

## Planned experience

When implemented, GoreeCloud Photos is intended to provide:

- automatic photo and video backup;
- timeline browsing;
- albums and favorites;
- offline media;
- search;
- sharing and family libraries;
- optional photo intelligence;
- editing;
- import/export;
- long-term preservation and recovery.

## Future backup-state model

The planned interface will distinguish states such as device only, waiting, uploading, processing, backed up, paused, and failed. A future cleanup feature must only remove local media after the server copy has been verified according to the applicable durability and recovery requirements.

## Privacy and protected media

Face analysis, person clustering, semantic search, location processing, protected media, and AI-assisted capabilities are planned privacy-sensitive features. They are not currently available and must not be represented as active Privacy Shield or Wardveil protection.

## Support and status

Use `README.md`, `FEATURES.md`, and `FEATURE-ROADMAP.md` for the current repository state. A future user manual revision must replace this availability notice only after a supported exact build and runtime have been independently verified.
