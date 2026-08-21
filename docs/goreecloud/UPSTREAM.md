# GoreeCloud Photos upstream provenance

## Project identity

GoreeCloud Photos is a GoreeCloud-maintained fork of Immich. The fork is an initial implementation foundation, not the permanent architectural end state. GoreeCloud intends to preserve upstream licensing and attribution while progressively introducing GoreeCloud-owned interfaces and replacing inherited components when that improves independence, security, maintainability, portability, interoperability, or product quality.

## Recorded fork baseline

- GoreeCloud repository: `GoreeCloud/goreecloud-photos`
- Upstream project: Immich
- Upstream repository: `immich-app/immich`
- Upstream branch represented by the initial GitHub fork: `main`
- Recorded baseline commit: `caea849fee509909d42220e327c855cc7de21b2a`
- Baseline commit subject: `feat: asset file apis (#25900)`
- Baseline recorded: 2026-08-21
- GoreeCloud foundation branch: `agent/fork-foundation`

The baseline SHA is the authoritative starting point for the first GoreeCloud foundation work. Future upstream synchronization must be deliberate, reviewed, and recorded rather than treated as an implicit product upgrade.

## Licensing

The inherited project is licensed under the GNU Affero General Public License version 3. GoreeCloud must preserve the applicable license, copyright notices, attribution, source-availability obligations, and other legal notices for inherited and modified covered work.

GoreeCloud-specific documentation in this directory does not erase or supersede upstream provenance.

## Upstream synchronization policy

Before an upstream synchronization is accepted into a GoreeCloud release line, the change should be reviewed for:

- security and vulnerability impact;
- database and migration changes;
- storage-layout or media-integrity changes;
- API and client compatibility;
- mobile background-backup behavior;
- authentication and authorization changes;
- machine-learning and metadata-processing changes;
- privacy, telemetry, and external-service behavior;
- licensing or dependency changes;
- conflicts with GoreeCloud-owned domain boundaries;
- deployment and recovery implications.

The GoreeCloud development history should remain auditable so that an administrator can identify which behavior is inherited from Immich and which behavior is implemented or modified by GoreeCloud.