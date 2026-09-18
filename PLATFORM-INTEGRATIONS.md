# GoreeCloud Photos — Integral Platform System Boundaries

**Status:** Phase 0 integration architecture  
**Current conformance:** All nine application-specific integrations remain applicable-blocked pending runtime implementation and evidence.

## 1. GoreeCloud Manager

### Responsibility

Manager may administer:

- quotas;
- storage consumption;
- processing queues;
- failed jobs;
- search/index health;
- derivative rebuild operations;
- intelligence enablement policy;
- sharing/public-link policy;
- retention policy;
- backup/recovery readiness;
- service version and health.

### Boundary

Manager receives aggregate operational information by default. It does not receive raw personal photos or videos merely to perform administration.

Manager actions that mutate Photos state require explicit authorization, audit, Policy evaluation where applicable, and Photos-domain enforcement.

## 2. Privacy Shield

Privacy Shield governs privacy-sensitive processing for:

- device-library access;
- photos/videos;
- faces and person clusters;
- location;
- metadata;
- search/indexing;
- OCR;
- intelligence;
- sharing;
- AI handoff;
- telemetry/diagnostics;
- retention and deletion.

Photos must expose purpose, processing location, retention, derived-data creation, sharing scope, and revocation behavior.

Sensitive processing fails closed when required privacy authorization cannot be established.

Derived intelligence is separable so revocation can remove or disable it without deleting ordinary original-media access unless the user requests deletion.

## 3. Wardveil Security

Wardveil supplies/normalizes security evidence and security-control integration for applicable boundaries.

Photos remains responsible for runtime enforcement of:

- upload admission;
- file/media validation;
- authentication handoff;
- authorization;
- session use;
- share-link access;
- privileged administration;
- worker isolation.

Wardveil branding never substitutes for enforcement.

Untrusted uploads remain in staging/quarantine until the required validation outcome permits promotion.

## 4. Everkeep

Everkeep governs Photos continuity requirements.

Covered authoritative scope includes:

- originals;
- relational metadata;
- ownership;
- albums;
- user metadata;
- edit instructions;
- required sharing state;
- authorized recovery material;
- restore procedures;
- integrity evidence;
- migration/export;
- succession requirements.

Thumbnails and ordinary indexes are rebuildable and should not consume irreplaceable-backup treatment unless a measured operational reason justifies it.

## 5. Glaze UI

Every controlled graphical surface uses Glaze UI 1.5.1 as the current Stable target.

Glaze presentation must adapt to platform-native behavior for web, Android, Linux, and any future Mobile B target.

Media remains the visual focus. Accessibility requirements include keyboard operation, screen-reader labeling, focus visibility, large text, high contrast, reduced motion, reduced translucency/effects, non-color-only state communication, and accessible video controls.

No conformance claim is made until exact client implementations are validated.

## 6. GoreeCloud Mesh

Mesh is used for ecosystem capability discovery and controlled cross-application coordination.

Potential Photos capabilities include:

- import from Drive;
- export/handoff to Drive;
- receive capture handoff from Camera;
- provide media selection to authorized applications;
- optional AI capability routing;
- recovery/status evidence routing.

Mesh events should contain stable identifiers and minimal metadata, not raw media, unless a capability explicitly requires a scoped media transfer.

Core Photos storage/browsing must not fail merely because Mesh is unavailable.

## 7. GoreeCloud Identity

Identity is authoritative for:

- users and subjects;
- authentication;
- sessions;
- devices;
- credential/trust inputs.

Photos is authoritative for:

- personal/family library membership;
- asset ownership;
- album authorization;
- Photos share grants;
- Photos domain permissions.

Identity subject IDs are referenced; Photos must not duplicate credential stores.

Detected persons in media never automatically map to Identity subjects.

## 8. GoreeCloud Policy

Policy provides shared rule evaluation for configurable governance decisions.

Initial Photos policy decision points include:

- upload limits;
- allowed media classes;
- storage quotas;
- retention;
- public-link availability;
- external-download permission;
- intelligence availability;
- location-sharing rules;
- administrative actions;
- deletion/purge requirements.

Photos remains the enforcement point for Photos operations. Policy decisions do not bypass Photos authorization.

Missing required policy evidence for a governed operation fails closed.

## 9. GoreeCloud Observability

Observability receives bounded operational evidence such as:

- request rate/latency/error class;
- upload bytes and failure classes;
- worker queue depth;
- processing duration/failure class;
- storage capacity;
- database health;
- index health;
- backup/recovery readiness references;
- version/build identity.

By default it does not receive:

- photo/video contents;
- captions;
- OCR text;
- face embeddings;
- exact locations;
- reusable share secrets;
- authentication secrets.

Operational identifiers should be minimized or pseudonymized where practical.

Unknown evidence must remain unknown; it must not become a healthy status.

## 10. GoreeCloud Sync boundary

GoreeCloud Sync is not a tenth Integral Platform System.

Photos owns its domain synchronization protocol and incremental change model. If a shared GoreeCloud Sync service later provides reusable transport or orchestration, Photos may integrate behind the synchronization boundary without surrendering Photos domain authority or data-model ownership.

## 11. Evidence rule

For every platform system, repository declarations must separately record:

- applicability;
- contract version where applicable;
- implementation/adaptor;
- runtime behavior;
- tests;
- exact-revision evidence;
- limitations;
- blockers.

Architecture documentation is not acceptance evidence.
