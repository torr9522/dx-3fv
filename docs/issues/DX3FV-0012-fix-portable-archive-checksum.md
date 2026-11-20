# DX3FV-0012: Make the release archive checksum portable

- Priority: P1 release integrity
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.14-feature-native-release`
- Release tag: `v0.1.15-fix-portable-release-checksum`
- Schema: 11 (unchanged)

## Problem

The native release builder passed an absolute archive path to `sha256sum`.
The resulting sidecar embedded the development-host path, so
`sha256sum -c` failed after copying both files to the test Panel even though
the archive bytes matched the expected digest.

## Root cause

`sha256sum` records the input filename verbatim. The build ran it from the
repository while supplying `$output_root/$release_name.tar.gz`.

## Fix

Generate the sidecar from inside the output directory and pass only the archive
basename. The sidecar can then be copied with the archive and verified from any
directory without rewriting trusted checksum data.

## Verification

- Build from the clean annotated release tag.
- Confirm the sidecar contains only the archive basename.
- Copy the archive and sidecar to a different directory and run
  `sha256sum -c` there.
- Copy both files to the authorized test Panel and verify again there.

## Rollback

Revert the commit pointed to by `v0.1.15-fix-portable-release-checksum`. No
database, API, UI, Agent, or Runtime behavior changes in this release.
