# GoreeCloud Photos — Glaze UI and Accessibility Acceptance Plan

**Status:** Phase 0 acceptance design  
**Target:** Glaze UI V1.5 / 1.5.1 current Stable  
**Implementation status:** Acceptance plan defined; no Photos client conformance is yet accepted.

## 1. Scope

This plan governs the GoreeCloud-controlled graphical Photos surfaces:

- web;
- Mobile A / Android;
- native Linux desktop;
- Mobile B once its platform is authoritatively assigned;
- authorized administrative Photos surfaces.

Headless workers and protocol-only components do not require visual conformance, but any UI they expose later becomes in scope.

## 2. Product visual hierarchy

Photos is media-first.

The visual hierarchy is:

1. photograph or video content;
2. information and selection state;
3. contextual actions;
4. navigation and transient Glaze surfaces.

Glaze materials frame media rather than obscure it.

Adaptive color or backdrop behavior must not reduce text contrast, hide state, or imply a stronger privacy/security/backup condition than authoritative evidence establishes.

## 3. Platform-native adaptation

### Web

Validate responsive layouts, keyboard-only operation, browser zoom/text scaling, pointer/touch input, and reduced-motion/reduced-transparency behavior where available.

### Android

Use Android-native application architecture and platform APIs. Validate TalkBack, large font/display scaling, touch targets, navigation modes, appearance, reduced animation, background-upload state, permission denial/revocation, offline state, rotation, and supported form-factor changes.

### Linux desktop

Use native Linux desktop behavior and GTK accessibility integration. Validate keyboard navigation, screen-reader exposure, window resizing, high contrast, reduced motion/effects, file-dialog and drag/drop workflows, focus visibility, and pointer/keyboard parity.

### Mobile B

Acceptance criteria must be specialized to the named platform before implementation. The platform is not inferred in this document.

## 4. Mandatory accessibility behavior

Every client must support, as applicable:

- full keyboard access to primary workflows;
- screen-reader labels and meaningful roles;
- logical focus order;
- visible focus indication;
- large text and scalable layout;
- high-contrast presentation;
- reduced motion;
- reduced translucency/effects;
- non-color-only status communication;
- accessible selection states;
- accessible image captions/descriptions;
- accessible video controls;
- adequate touch targets;
- state announcements for upload/sync/error transitions.

Generated descriptions may supplement but never overwrite or silently replace user-authored descriptions.

## 5. Truthful system-state presentation

The UI must visibly distinguish states whose operational meaning differs, including device only, waiting, uploading, verifying, verified remote, failed, server unavailable, offline copy available, protected media, shared, and link shared.

A favorable icon or Glaze treatment is not evidence. Wardveil, Privacy Shield, Everkeep, backup, synchronization, and availability status must come from the authoritative producer for the represented scope. Unknown or stale required evidence remains unknown/non-passing.

## 6. Media viewer requirements

The viewer must preserve readable controls over bright and dark media, predictable controls, an accessible way to reveal hidden chrome, zoom/pan behavior that does not trap keyboard or assistive-technology users, captions/descriptions independent from overlays, and visible media availability/failure state.

## 7. Timeline and grid requirements

Validate progressive loading without focus loss, stable selection state, keyboard range selection where supported, year/month/day navigation, density changes without information loss, virtualization/lazy loading with correct accessibility semantics, and non-color-only state presentation.

## 8. Forms and destructive operations

Sharing, deletion, purge, device cleanup, protected-media changes, and privacy-sensitive controls require explicit labels, clear current state, consequence-proportional confirmation, recoverability information where applicable, accessible error messages, and no misleading default that expands sharing or exposure.

## 9. Reduced-effects mode

Where blur, translucency, adaptive imagery, or animation would reduce readability, accessibility, or performance, the client must provide or respect a reduced-effects path. Reduced-effects presentation must preserve hierarchy and state.

## 10. Acceptance evidence

Each supported graphical target requires exact-revision evidence including build identity, target OS/runtime and version, Glaze UI target version, automated UI/accessibility checks where available, keyboard navigation results, screen-reader results, contrast/reduced-effects validation, representative responsive/form-factor evidence, critical state-presentation tests, and known limitations.

Source existence or component-library linkage alone does not establish acceptance.

## 11. Stable gate

No Photos client may claim Glaze UI 1.5.1 conformance or Stable visual/accessibility acceptance until the exact client revision is identified; applicable Glaze requirements are implemented; representative runtime evidence exists; accessibility acceptance passes; system-state presentation remains truthful under success, failure, stale, unknown, and offline conditions; and unresolved deviations are corrected or governed as explicit nonconformities.

The current repository remains Concept / Planning.
