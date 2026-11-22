# DX3FV-0075: Publish v1.0.4 artifact and fixed installer mapping

- Priority: P1
- Product release: `v1.0.4`
- State: FIXED_IN_DX3FV

## Problem

The immutable `v1.0.2` through `v1.0.4` releases were published and verified,
but the one-command installer still defaulted to `v1.0.1` and rejected those
newer fixed versions.

## Fix

- Keep every historical fixed-version mapping unchanged.
- Add exact release filenames and SHA256 values for `v1.0.2`, `v1.0.3`, and
  `v1.0.4`.
- Set the installer default and installer version to `v1.0.4`.
- Keep `latest` unsupported so an installation cannot follow an unversioned
  development branch.
- Store the verified immutable release archives in `release/artifacts/` for the
  offline installer contract.

This publication-only change does not modify Panel, Agent, Runtime, database
schema, or the already published `v1.0.4` assets.

## Rollback

Revert the publication commit. Existing installations and immutable product
release tags remain unchanged.
