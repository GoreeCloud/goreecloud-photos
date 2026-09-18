# GoreeCloud Photos — Contract Schemas

These files are Phase 0 machine-readable contract baselines. They define intended shapes and invariants; they do not prove a running API or persistence implementation.

Current schemas:

- asset.v1.schema.json — stable Asset representation.
- sync-change.v1.schema.json — ordered synchronization change representation.
- upload-session.v1.schema.json — resumable-upload session representation.

Contract evolution rules:

1. Existing required field meaning does not silently change.
2. Breaking changes require a new contract version or explicit migration.
3. Additive fields must preserve safe unknown-field handling.
4. Sensitive information must not be added to general-purpose contracts merely for diagnostics.
5. Implementation and validation evidence remain separate from schema existence.
