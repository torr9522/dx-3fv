# DX3FV-0015: Remediate reachable Agent transport advisories

- Priority: P0/P1 Runtime security and availability
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.18-fix-panel-go-advisories`
- Target release: `v0.1.19-fix-agent-transport-advisories`
- Schema: 11 (unchanged)

## Problem and evidence

After pinning Go 1.26.11, `govulncheck ./...` reports 23 reachable findings
in the Agent dependency graph. The affected paths include HTTP/2, SSH, DTLS,
WebTransport, QUIC, gRPC, IDNA, and Unicode normalization. Findings reach
actual Agent and embedded GOST symbols; no exploit was executed.

The minimum fixed versions covering the reported set are:

- `golang.org/x/net` v0.60.0;
- `golang.org/x/crypto` v0.57.0 (required by the fixed `x/net` release);
- `golang.org/x/text` v0.42.0 (required by fixed `x/net`/`x/crypto`);
- `github.com/pion/dtls/v3` v3.1.4;
- `github.com/quic-go/webtransport-go` v0.11.1;
- `github.com/quic-go/quic-go` v0.60.0 (required by WebTransport v0.11.1);
- `google.golang.org/grpc` v1.83.2 (the first version covering the full
  reported gRPC set).

## Scope and non-goals

Update the explicit transport dependencies in both the Agent module and its
local `go-gost/x` module, then accept only the minimal transitive changes
selected by Go's module solver. Do not replace the local fork, regenerate
protobuf source, alter Panel APIs, change GOST service configuration, or add
new protocols.

## Required verification

1. `govulncheck ./...` reports zero reachable Agent findings.
2. Agent lifecycle and complete embedded-fork test suites pass.
3. Relay authentication, required-route, UDP wildcard/session-isolation,
   WebSocket reporting, service lifecycle, SOCKS5 config parsing, and
   TCP/UDP listener tests remain green.
4. Native release builds with Go 1.26.11 and passes checksums.
5. Test Panel upgrade, browser regression, rollback, and return pass.
6. Real TCP/UDP/Relay data-plane verification remains required after nodes
   become reachable; static and local tests are not an E5 PASS.

## Rollback

Use the native rollback script and matching pre-upgrade database backup.
Schema 11 and the Panel protocol are unchanged. Existing nodes must not be
upgraded until their identity is verified and management endpoint is
reachable.

## Verification result

- Go 1.26.11 `govulncheck ./...`: zero reachable vulnerabilities in the
  Agent module.
- Agent lifecycle suite: PASS.
- Complete local `go-gost/x` test suite: PASS, including Relay chain,
  required-route, UDP listener/session isolation, service lifecycle,
  WebSocket reporter, configuration parser, and session retirement packages.
- Real-node TCP/UDP/Relay verification: BLOCKED because every exact authorized
  node management endpoint timed out before SSH authentication.
