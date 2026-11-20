# DX3FV-0023: Restore Agent-based business-path diagnosis

- Priority: P1 diagnostics regression
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.26-fail-closed-missing-agent-ca`
- Release tag: `v0.1.27-restore-business-diagnosis`
- Schema change: none

## Problem and evidence

The operator rule prohibited Codex or the development machine from probing an
IX egress public address or bypassing the Guangzhou SSH jump. Releases
`v0.1.20` through `v0.1.24` incorrectly translated that operator boundary into
a product-wide ban on `TcpPing`, `UdpPing`, `EchoProbe`, and
`ServiceMonitorCheck` whenever the source node owned an IX port pool. This
disabled legitimate IX Agent to configured Landing checks, original Tunnel and
Forward diagnosis, automatic tunnel quality collection, and Service Monitor.

The archived ForwardX implementation was inspected as behavioral reference.
Its manual tunnel planner builds one probe per persisted topology edge, sends
the work to that edge's source Agent, and targets only the next node's selected
address and listen port. ForwardX is AGPL-3.0; no source was copied.

## Fix

- Restore the segmented multiport diagnosis from `v0.1.12`.
- Restore original Tunnel/Forward manual diagnosis, tunnel quality probes,
  Service Monitor checks, and the common Agent command dispatcher.
- Keep target construction bound to persisted topology and current route data.
- Fail before dispatch when a multiport route's active NAT resource no longer
  matches its IX node, pool, public port, or internal port.
- Keep UDP `UNVERIFIED` unless a controlled nonce echo returns successfully.
- Keep development-machine and SSH jump restrictions as operational policy;
  they are not enforced by disabling authenticated product Agent commands.

## Verification

- Unit coverage checks exact source Agent, address, and port for both
  GZ-to-IX and IX-to-Landing TCP legs.
- A corrupted NAT-resource association fails closed and dispatches no probe.
- Existing handler, contract, health, tunnel quality, frontend type, and build
  tests cover the five restored call paths.
- Real-browser and real-Agent evidence is stored outside Git under the Phase 2
  evidence directory.

## Rollback

Revert the commit tagged `v0.1.27-restore-business-diagnosis`. This restores
the over-broad v0.1.24 behavior and is only appropriate if all IX-origin
product diagnostics must again be disabled.
