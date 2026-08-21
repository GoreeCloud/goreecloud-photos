# GoreeCloud Photos — Inherited Workflow Classification

Status: Active transition record

## Purpose

This record classifies inherited Immich GitHub Actions before GoreeCloud treats any of them as release or production authorities.

The inherited workflows remain useful engineering references, but a workflow is not GoreeCloud-approved merely because it exists in the fork or previously worked upstream.

## Classification model

Each inherited workflow belongs to one of four transition classes:

- **Retain and adapt** — function is useful, but GoreeCloud ownership, names, credentials, permissions, artifacts, or release assumptions must be reviewed.
- **Reference only** — useful as an upstream implementation reference, but not an authoritative GoreeCloud workflow.
- **Replace** — a GoreeCloud-owned workflow should provide the required behavior.
- **Retire** — workflow serves upstream project/community operations that are outside GoreeCloud Photos scope.

No inherited workflow is authorized to publish a GoreeCloud Stable release until its dependencies, permissions, secrets, artifact destinations, provenance, rollback behavior, and release semantics are explicitly accepted.

## Initial classification

### Security and source validation

`codeql-analysis.yml` — **retain and adapt**.

Reason: CodeQL scanning is useful, but the inherited workflow requests an Immich-owned GitHub App token through `immich-app/devtools` and upstream-only secrets. GoreeCloud should retain least-privilege scanning while eliminating unnecessary upstream credential dependencies.

`check-openapi.yml` — **retain and adapt**.

Reason: API-contract validation is strategically useful during fork-to-native transition. The long-term authority must become GoreeCloud-controlled APIs and compatibility contracts.

### Build and package workflows

`docker.yml` — **replace incrementally**.

Reason: the inherited workflow contains Immich image names, upstream reusable workflows, upstream runner assumptions, and upstream release/tag semantics. GoreeCloud needs independently controlled image names, registry policy, build provenance, SBOM/vulnerability evidence, immutable release references, and release gates.

`build-mobile.yml` — **retain and adapt**.

Reason: Android/iOS build knowledge remains valuable while Photos Clients are transitioned. Signing, application identity, artifact names, credentials, release channels, and client acceptance must be GoreeCloud-owned.

`fdroid.yml` — **reference only until Android distribution is defined**.

Reason: F-Droid may remain an appropriate open-source distribution path, but the GoreeCloud Photos Android identity and release lifecycle are not yet accepted.

`cli.yml` — **retain and adapt**.

Reason: command-line migration, import, export, administration, and validation are first-class Photos capabilities. Long-term CLI behavior should become part of Photos Clients and GoreeCloud-controlled migration tooling.

### Documentation workflows

`docs-build.yml` — **retain and adapt**.

Reason: documentation build validation is useful.

`docs-deploy.yml` and `docs-destroy.yml` — **reference only / replace before use**.

Reason: publication and teardown destinations are upstream deployment concerns until GoreeCloud explicitly defines its own documentation publication architecture.

### Upstream community-management automation

`auto-close.yml` — **retire from GoreeCloud product authority**.

`close-duplicates.yml` — **retire or replace with GoreeCloud-specific repository governance**.

Reason: these automate upstream community triage behavior and should not silently govern GoreeCloud issues or contributions.

### Repository maintenance automation

`cache-cleanup.yml` — **retain and adapt if still justified**.

`fix-format.yml` — **retain and adapt only if it can operate with least privilege and GoreeCloud-controlled credentials**.

Reason: maintenance automation is useful only when it does not introduce hidden write authority or upstream-only authentication assumptions.

## External action policy

An external GitHub Action is not automatically prohibited. Before long-term use, GoreeCloud should verify:

1. exact immutable revision pinning;
2. minimum required permissions;
3. whether the action sends repository or build information to an external service;
4. whether secrets are exposed to it;
5. whether a GoreeCloud-owned or simpler first-party alternative is practical;
6. maintenance and vulnerability status;
7. rollback and replacement path.

`immich-app/devtools` is therefore considered a **transitional upstream dependency**, not a permanent GoreeCloud build authority.

## Release boundary

Until replacement/adaptation work is complete:

- inherited successful workflow runs are engineering evidence, not GoreeCloud production acceptance;
- inherited image names and release tags are not GoreeCloud Stable release identities;
- upstream GitHub App secrets must not be recreated merely to make inherited automation pass;
- GoreeCloud-owned validation may run independently in parallel;
- runtime behavior must remain unchanged unless a separate tested implementation change explicitly requires it.

## Next transition targets

Priority order:

1. GoreeCloud source/security validation that requires no Immich-only credentials;
2. independent container build and artifact provenance;
3. Photos API contract validation;
4. Photos Clients build and acceptance;
5. controlled documentation publication;
6. removal of upstream community/release automation that no longer has a GoreeCloud role.
