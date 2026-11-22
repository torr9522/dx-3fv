# DX3FV-0077: Prevent stale Forward rollback after concurrent updates

## Problem

`v1.0.5` allowed two updates of the same Forward to overlap across database
persistence, Agent runtime apply, and rollback. If the older request failed
during runtime apply after the newer request succeeded, the older rollback
could restore its stale snapshot over the newer database and runtime state.

The defect affects both direct single-port Forwards and Forwards attached to a
shared Tunnel. IX Routes use their separate durable migration transaction and
are outside this change.

## Reproduction

The deterministic contract test blocks the older request in Agent apply,
starts a newer update, then forces the older apply to fail. Before the fix the
final target reverted from `9.9.9.9:443` to `1.1.1.1:443` instead of retaining
the last successful update.

## Fix

Serialize one Forward's update lifecycle by Forward ID. The lock covers access
resolution, persistence, runtime apply, and rollback, and is removed after the
last waiter exits. Different Forward IDs remain independent, so sibling
businesses do not block each other.

## Verification

- `TestConcurrentForwardUpdateFailureCannotRollbackNewerSuccess`
- Handler, repository, and contract suites
- Real TCP/UDP A-to-B-to-A online edits for direct, Tunnel Forward, and IX Route
- Three-business sibling isolation on every forwarding model
- Unreachable and closed target recovery without changing entry ports

No schema, API payload, Agent protocol, GOST runtime model, or existing business
data is changed.
