# DX3FV-0020: Enforce IX probe policy at command dispatch

> Superseded by DX3FV-0023. Node-role dispatch blocking confused an operator
> access restriction with legitimate product diagnostics.

- Priority: P0 operational safety
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.23-fix-ix-service-monitor-probes`
- Release tag: `v0.1.24-enforce-probe-egress-policy`
- Schema change: none

## Problem and evidence

Page-specific and scheduler checks do not cover every internal caller. The
federation runtime and future control-plane code can call the common node
command dispatcher directly. Without a final policy guard, a caller could send
an active probe command to an IX node even though current UI paths are blocked.

## Fix

The common `sendNodeCommandWithTimeout` path now checks persisted node identity
before dispatching these active-probe commands:

- `TcpPing`;
- `UdpPing`;
- `EchoProbe`;
- `ServiceMonitorCheck`.

For an IX/middle/exit node it returns `NOT_PROBED` before local WebSocket or
remote federation transport is selected. Runtime lifecycle commands such as
`UpdateService`, normal GOST forwarding, metrics, flow reports, and Agent
management are unchanged.

This is defense in depth. The multiport, original Tunnel/Forward, automatic
quality, and service-monitor paths retain their more specific policy behavior.

## Verification

- `TestProbeDispatchPolicyBlocksRestrictedNodeBeforeAgentCommand` checks all
  four command types and proves the mock Agent receives zero commands.
- `TestProbeDispatchPolicyDoesNotBlockRuntimeCommands` proves an
  `UpdateService` for normal business Runtime still reaches the same restricted
  Agent.
- Existing GZ-to-IX diagnosis tests prove a full entry node remains eligible.

## Rollback

Revert the commit pointed to by `v0.1.24-enforce-probe-egress-policy`.
Rollback removes the final protocol-level protection and is prohibited under
the current operator policy.
