# DX3FV-0011: Reproducible native release

- Priority: P1 deployability and recovery
- Status: IMPLEMENTED, TEST_PANEL_ROLLBACK_PENDING
- Baseline: `v0.1.13-feature-route-edit-migration`
- Release tag: `v0.1.14-feature-native-release`
- Schema: 11 (unchanged)

## Problem

DX-3FV requires a container-free release that keeps the Panel and Agent
artifacts reproducible, verifies package integrity, and supports atomic Panel
upgrade and rollback without losing the pre-migration SQLite database.

## Scope

- Build the Panel, Agent, and frontend from the pinned repository state.
- Refuse a dirty worktree, a lightweight tag, or a tag not pointing at HEAD.
- Package native systemd and Caddy templates without credentials.
- Record source identity, dependency manifests, and SHA256 values.
- Install and upgrade the Panel through a stable `current` symlink.
- Install an Agent only with an explicit, pre-verified node configuration.
- Restore the database and binary release together after a failed upgrade.

## Verification

- `bash -n` for all packaged shell scripts.
- Actual release build from the annotated tag.
- Package and archive SHA256 verification.
- Archive extraction and package verification in an isolated temporary path.
- Secret-pattern scan of the final artifact.

## Rollback

Use `rollback-native.sh PREVIOUS_RELEASE MATCHING_DATABASE_BACKUP`. The script
stops the Panel before changing either resource and restores the previous
release automatically if its health check fails.
