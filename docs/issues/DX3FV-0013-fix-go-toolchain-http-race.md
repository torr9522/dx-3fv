# DX3FV-0013: Pin a fixed Go release toolchain

- Priority: P0 availability / release security
- Status: FIXED_IN_DX3FV
- Affected release: `v0.1.15-fix-portable-release-checksum`
- Target release: `v0.1.17-fix-go-toolchain-http-race`
- Schema: 11 (unchanged)

## Problem and evidence

The v0.1.15 native manifest records Go 1.27.3. `govulncheck` found reachable
Panel and Agent paths for `GO-2026-6859`, an HTTP/1 server crash caused by a
`MaxBytesReader` header race. The official fixed branches are Go 1.26.11 and
Go 1.27.4.

The modules declared Go 1.26.10, but the native release script called the
host's unqualified `go` binary. A host with Go 1.27.3 therefore produced
vulnerable release binaries despite the earlier minimum-version policy.

## Root cause

A module `go` directive is a language/toolchain floor, not an exact release
builder pin. A newer host toolchain wins unless `GOTOOLCHAIN` is constrained.

## Fix

- Raise all three module floors and the backend builder image to Go 1.26.11.
- Pin every native release build and module-SBOM command to
  `GOTOOLCHAIN=go1.26.11`.
- Record that exact toolchain in the release manifest.
- Update the documented safe patch levels; Go 1.27.3 is no longer accepted.

This issue does not change application code, database schema, API, Agent
protocol, GOST configuration, or Runtime behavior.

## Verification

1. Run all three module test suites with `GOTOOLCHAIN=go1.26.11`.
2. Run `govulncheck ./...` for Panel and Agent using Go 1.26.11 and confirm
   `GO-2026-6859` is absent.
3. Build the clean annotated release and inspect both binaries with
   `go version -m`; both must record `go1.26.11`.
4. Verify the release manifest and portable archive checksum.
5. Upgrade the test Panel, run health and browser acceptance, then exercise
   rollback and return to the fixed release.

## Rollback

Roll back to v0.1.15 with its matching database backup. Schema 11 is
unchanged. Rollback restores binaries built by Go 1.27.3 and therefore also
restores the known vulnerability; it is for emergency compatibility only.
