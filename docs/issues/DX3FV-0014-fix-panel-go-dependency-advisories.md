# DX3FV-0014: Remediate reachable Panel Go dependency advisories

- Priority: P0/P1 security and availability
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.17-fix-go-toolchain-http-race`
- Target release: `v0.1.18-fix-panel-go-advisories`
- Schema: 11 (unchanged)

## Problem and evidence

With the standard-library issue removed, `govulncheck ./...` still reports
nine reachable Panel findings in two modules:

- `golang.org/x/crypto` v0.42.0: SSH deadlock, infinite-loop, parser and
  callback findings used through the nftables SSH runner. The highest fixed
  version required by the reported set is v0.56.0.
- `golang.org/x/text` v0.36.0: crafted-input panic/infinite-loop findings
  reached through PostgreSQL/GORM dependencies. The highest fixed version
  required by the reported set is v0.41.0.

No exploit was run. Reachability is based on Go symbol analysis against the
actual Panel module.

## Scoped fix

Upgrade only those Panel module dependencies and the minimal transitive
versions selected by Go's module solver. Do not change Agent dependencies,
the embedded GOST fork, application behavior, schema, API, or UI in this
issue.

## Required verification

1. `govulncheck ./...` returns no reachable Panel findings.
2. Full backend unit and contract suite passes, except that the independently
   tracked `DX3FV-0001` suite-scope flake must be reported rather than hidden.
3. The limiter rollback test passes ten isolated repetitions.
4. Native release builds from a clean annotated tag with Go 1.26.11.
5. Test Panel upgrade, public desktop/mobile acceptance, health, rollback,
   and return to the fixed release pass without affecting the reference site.

## Rollback

Use the native rollback script with the matching pre-upgrade SQLite backup.
Schema 11 is unchanged. Rolling back also restores the vulnerable dependency
versions and is only an emergency compatibility action.

## Verification result

- Go 1.26.11 `govulncheck ./...`: zero reachable vulnerabilities in the
  Panel module.
- Go 1.26.11 `go test ./...`: PASS, including Repository and contract suites.
- Limiter rollback contract: PASS for ten isolated repetitions.
- The module solver additionally raised only `x/sync` to v0.22.0 and `x/sys`
  to v0.47.0; `go mod tidy` removed obsolete checksums.
