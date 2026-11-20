# FLVX rc16 upstream bug audit

Audit snapshot: 2026-11-19 (Asia/Shanghai)  
Audited baseline: `f357b30500ef99a6cf45d18f0c8a1e1881826b57`
(`3.0.1-rc16`)  
DX-3FV audit HEAD at scan time: `b83b12ed0902ced03a56b1ebb07bd3212ca639e3`

## Scope and method

This audit distinguishes an open report from a defect proven to remain in
rc16. Evidence was collected from the public Sagit-chu/flvx issue and pull
request API, local complete Git history, the checked-out rc16 source and its
tests, `govulncheck`, and `pnpm audit`. No production system was accessed. The
test forwarding nodes were unreachable, so no item below is described as an
E5/data-plane PASS.

For every merged fix below, inclusion was verified with
`git merge-base --is-ancestor <fix> f357b305...`; a closed issue or pull
request by itself was not accepted as proof.

Status meanings follow `docs/issues/README.md`. `REPRODUCED` for a dependency
finding means the scanner found a reachable symbol path; it does not mean an
exploit was executed against a running service.

## rc6 to rc16 fixes

rc16 is eleven commits after rc6. It contains rc6 and all commits in this
table.

| Area | Public evidence | Fix commit | rc16 status | Local evidence |
| --- | --- | --- | --- | --- |
| UDP wildcard reply source and destination-session isolation | PR #565 | `3549fe5` (rc6) | `FIXED_UPSTREAM` | `go-gost/x/listener/udp`, `docs/udp-wildcard-source.md` |
| Disabled users retained tunnel access | issue #562 / PR #568 | `46bfb86` | `FIXED_UPSTREAM` | permission and tunnel visibility contract tests |
| Relay endpoints lacked independent authentication and missing routes failed open | PR #569 | `fadc2b6` | `FIXED_UPSTREAM` | `internal/relaysecurity`, `handler/relay_auth.go`, `relay_auth_contract_test.go`, required-route tests in the embedded GOST fork |
| Agent upgrade verification and OpenRC lifecycle | issue #558 / PR #570 | `743bf6f` | `FIXED_UPSTREAM` | upgrade checksum and lifecycle tests in Agent socket package |
| Realtime monitor stale-request/state races | PR #571 | `a784350` | `FIXED_UPSTREAM` | monitor/realtime tests |
| Panel self-upgrade durability and observability | PR #572 | `7e81e1c` | `FIXED_UPSTREAM` | `internal/panelupgrade` job state and tests |
| SQLite relay-auth migration backfill | PR #573 | `3209ff8` | `FIXED_UPSTREAM` | `repository_relay_auth_migration_test.go` |
| PostgreSQL relay chain-role type mismatch | PR #579 | `8339f52` | `FIXED_UPSTREAM` | `repository_relay_auth_test.go` validates text binding |
| Cross-instance entry-port allocation race | PR #584 | `f357b30` (rc16) | `FIXED_UPSTREAM` | transactional reservation and port-allocation contract tests |

## Older high-value fixes already contained in rc16

| Area | Public evidence | Fix commit | rc16 status |
| --- | --- | --- | --- |
| Shared rules lost during flow reporting/recovery | PR #560 | `c952d2f` | `FIXED_UPSTREAM` |
| Federated Agent runtime update instability | PR #550 | `f260145` | `FIXED_UPSTREAM` |
| Replaced Tunnel sessions not retired | PR #541 | `0b23d6f` | `FIXED_UPSTREAM` |
| Unbounded Tunnel quality probes / connection growth | issue #532 / PR #533 | `e112d81` | `FIXED_UPSTREAM` |
| nftables state after Panel restart and recovery diagnostics | PRs #524/#525 | `0c2acfd`, `8e513a1` | `FIXED_UPSTREAM` |
| Proxy Protocol direction coupling | PR #522 | `82f6047` | `FIXED_UPSTREAM` |
| Concurrent WebSocket writes | PR #510 | `032b0f0` | `FIXED_UPSTREAM` |
| Auth/config/backup hardening | PR #503 | `2eee506` | `FIXED_UPSTREAM` |
| Unsafe Proxy Protocol rollout | PR #494 | `4ebd670` | `FIXED_UPSTREAM` |

PR #502 was closed without merge. Its API `merge_commit_sha` is not an
ancestor of rc16. The security work that can be proven present is PR #503 and
its tests; this audit does not claim that every statement in the unmerged PR
is independently present.

## Open reports at the rc16 boundary

| Report | Classification | Evidence and decision |
| --- | --- | --- |
| #585 SOCKS5 traffic through a Tunnel fails while Shadowsocks works | `UNKNOWN` | Filed after rc16. The embedded fork contains SOCKS5 TCP and UDP handlers/connectors, but the report has no versioned config, logs, or minimal reproducer. Test nodes are unreachable. Do not patch by inference. Build an isolated SS/SOCKS5 fixture and capture both generated GOST configs before deciding whether this is Panel generation, Relay routing, or protocol behavior. |
| #518 backup exit does not switch when the exit Agent is offline | `UNKNOWN` | Report is still open and has no reproducible topology or logs. Source includes selector/probe recovery fixes, but no authorized data plane was available to reproduce this exact failure. |
| #529 Tunnel type cannot be changed during edit | `NOT_RELEVANT` as a bug; `DEFERRED` as a feature | Changing a live Tunnel type changes runtime topology and cleanup semantics. The current UI restriction is observable product behavior, not proof of corruption. DX-3FV does not change original Tunnel semantics. |
| #513 accept Proxy Protocol on a managed entry | `DEFERRED` | The report requests a new topology/configuration option. PR #522 fixed send/accept direction handling, but the requested UI/config surface is not part of the Phase 2 multiport scope. |
| #348 node name truncation at 125% desktop scaling | `UNKNOWN` | Public browser acceptance passed at 1440 px and 390 px, but 125% OS/browser scaling was not separately captured. It is not marked fixed. |
| #512 separate TCP-only and UDP-only ordinary forwarding | `NOT_RELEVANT` | Feature request. DX-3FV protocol selection applies only to its independent multiport module and does not rewrite ordinary Forward defaults. |
| #559 bulk API import/export and #531 rule clone | `NOT_RELEVANT` | Feature requests outside the authorized product scope. |

## Dependency and runtime security findings

These findings were discovered after the rc16 freeze and are not erased by
the upstream-fix table above.

### Go standard library

`govulncheck ./...` found reachable `GO-2026-6859` paths in both Panel and
Agent binaries built by the native v0.1.15 release. The release manifest says
Go 1.27.3; the official fixed releases are Go 1.27.4 and 1.26.11.

Classification: `PRESENT_IN_RC16`, `REPRODUCED` by symbol analysis, not an
executed exploit. A separate toolchain pin/rebuild/redeploy issue is required.

### Panel Go dependencies

The backend scan found reachable paths in:

- `golang.org/x/crypto` v0.42.0: SSH denial-of-service and parser/callback
  advisories; fixed baseline v0.56.0 for all reported paths.
- `golang.org/x/text` v0.36.0: invalid-input panic/infinite-loop advisories;
  fixed baseline v0.41.0.

Classification: `PRESENT_IN_RC16`, `REPRODUCED` by symbol analysis. Upgrades
must be isolated and followed by SQLite/PostgreSQL, nftables SSH, auth, backup,
and full backend regression tests.

### Agent / embedded GOST dependencies

The Agent scan reported 24 reachable findings in seven modules plus the Go
standard library. The affected dependency groups include:

- `golang.org/x/net` v0.53.0 (HTTP/2 and IDNA findings; remediated at v0.60.0);
- `golang.org/x/crypto` v0.50.0 (SSH findings; remediated at v0.57.0 because
  of fixed-module constraints);
- `golang.org/x/text` v0.36.0 (remediated at v0.42.0 because of
  fixed-module constraints);
- `github.com/pion/dtls/v3` v3.0.11 (fixed v3.1.4);
- `github.com/quic-go/webtransport-go` v0.10.0 (fixed v0.11.1);
- `github.com/quic-go/quic-go` v0.59.0 (remediated at v0.60.0 because
  WebTransport v0.11.1 requires it);
- `google.golang.org/grpc` v1.80.0 (remediated at v1.83.2 after the full
  advisory set was rescanned).

Classification: `PRESENT_IN_RC16`, `REPRODUCED` by symbol analysis. This does
not authorize replacing the embedded `go-gost/x` fork with upstream. Each
dependency remediation needs Agent protocol, TCP/UDP, Relay, WebSocket,
QUIC/HTTP2/DTLS where applicable, restart, and rollback validation.

### Frontend dependency graph

`pnpm audit --prod --json` reported 50 findings (1 critical, 24 high, 24
moderate, and 1 low). Direct runtime dependencies include axios 1.15.2 and
react-router-dom 6.30.3; affected transitive/build dependencies include
follow-redirects, form-data, tar, PostCSS, nanoid, source-map-js,
`@remix-run/router`, and react-router.

The critical `tar` path is under Tailwind's build tooling and is not proof of
a server-side runtime exploit in the static frontend. Axios and router paths
are shipped browser code and require a separate compatibility upgrade plus
full route/login/browser acceptance. Classification: `PRESENT_IN_RC16`; no
blanket upgrade was performed during this audit.

## Required follow-up order

1. Pin a fixed Go toolchain and rebuild both binaries in an isolated issue.
2. Remediate backend Go modules and rerun full backend/contract scans.
3. Remediate Agent transport dependencies without replacing the embedded
   fork; run protocol and data-plane regression before deployment.
4. Remediate direct frontend runtime packages, then build tooling, and repeat
   desktop/mobile browser acceptance against the FLVX reference site.
5. Reproduce issue #585 only after an authorized Agent data plane is
   reachable; do not infer a fix from the issue title.

Until those remediations and real-node tests complete, the Phase 2 result
remains `PARTIAL`.
