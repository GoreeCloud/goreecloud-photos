# GoreeCloud Photos initial CI audit

## Scope

This audit records the first inspect-only review of the inherited GitHub Actions configuration at the GoreeCloud Photos fork baseline `caea849fee509909d42220e327c855cc7de21b2a`.

No inherited workflow is modified by this audit.

## Initial findings

### Large upstream workflow surface

The fork inherits a substantial `.github/workflows` collection covering mobile builds, Docker images, OpenAPI checks, CLI work, documentation, CodeQL, automation, and other upstream project processes. This is expected for Immich but means GoreeCloud should not assume every inherited workflow is immediately suitable for the fork.

### Upstream GitHub App secrets

The inherited CodeQL workflow uses `immich-app/devtools/actions/create-workflow-token` and expects the upstream secrets `PUSH_O_MATIC_APP_CLIENT_ID` and `PUSH_O_MATIC_APP_KEY`. Those credentials are upstream-specific and are not part of the GoreeCloud fork foundation.

The inherited Docker workflow also uses the same upstream token action and secrets in its pre-job. Its build/release path additionally contains upstream-oriented image naming and reusable workflow dependencies.

Therefore, a failing inherited workflow on the fork must not automatically be interpreted as a GoreeCloud source defect until its upstream secret and infrastructure assumptions are separated from actual code validation.

### Third-party action pinning

The inspected CodeQL and Docker workflows pin important GitHub Actions and Immich devtools references to commit SHAs rather than floating tags. This is a positive supply-chain characteristic and should be preserved or improved when GoreeCloud creates replacement workflows.

### Workflow permissions

The inspected workflows use explicit permission blocks, including top-level `permissions: {}` followed by job-specific permissions. GoreeCloud-specific workflows should follow the same least-privilege direction.

### Container publication assumptions

The inherited Docker workflow targets Immich-oriented image names such as `immich-server` and `immich-machine-learning` and includes upstream Docker Hub/GHCR assumptions. GoreeCloud must not start publishing images merely by inheriting this workflow. Image identity, registry ownership, package permissions, versioning, and release authorization require a separate GoreeCloud decision.

## Immediate GoreeCloud CI policy

During the fork-foundation phase:

1. Do not broadly rewrite or disable inherited CI until each relevant workflow has been classified.
2. Add a small GoreeCloud-owned foundation workflow that validates GoreeCloud governance files without requiring upstream secrets.
3. Treat inherited upstream release/publish workflows as unapproved for GoreeCloud production publication until explicitly adapted and validated.
4. Preserve exact-revision validation and least-privilege permissions.
5. Do not add reusable credentials to source, workflow files, documentation, or PR text.
6. Do not represent an inherited workflow failure caused by missing upstream-only secrets as an application regression.

## Follow-up audit categories

The next CI review should classify inherited workflows into:

- safe validation workflows that can be retained with little or no modification;
- workflows that require GoreeCloud-owned secret or token replacement;
- upstream automation that does not belong in the fork;
- release/publication workflows requiring explicit GoreeCloud registry and signing policy;
- workflows that should eventually be replaced by smaller GoreeCloud-owned equivalents.

The classification should be completed before a Stable GoreeCloud Photos release line is created.