# DX3FV-0022: Native Agent install accepts a missing CA trust store

- Priority: P1 Agent availability and transport security
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.25-fix-native-agent-runtime-bootstrap`
- Release tag: `v0.1.26-fail-closed-missing-agent-ca`
- Schema change: none

## Problem and evidence

GZ2 and Frontsea IX2 did not have the `ca-certificates` package or
`/etc/ssl/certs/ca-certificates.crt`. The repaired Agent process stayed
`active/running`, but every WSS and HTTPS management connection failed with:

```text
x509: certificate signed by unknown authority
```

The WS fallback correctly failed with HTTP 308 rather than bypassing TLS.
Blindbox IX had a current CA bundle and connected immediately with the same
Agent binary and Panel configuration.

## Root cause

The native installer verified its release files and local process state but did
not verify the OS TLS trust prerequisite. A live process was therefore reported
as a successful install even though it could never authenticate the Panel.

## Fix

Before copying files or starting systemd, the installer requires a non-empty CA
bundle at `SSL_CERT_FILE`, or at the standard Debian path when that variable is
unset. A missing bundle produces an explicit error and exits without creating a
new release directory or changing the current symlink.

The installer does not disable certificate verification, fall back to insecure
WS, or download packages from an IX node.

## Test environment remediation

The official Debian CA bundle already present on the development host was
transferred offline through the Guangzhou jump. Its SHA256 was verified before
installation on GZ2 and Frontsea IX2. No IX-origin apt, HTTP, ping, TCP probe,
UDP probe, or Landing probe was used. Both Agents established WSS afterward and
all three installed test nodes became ONLINE.

## Verification

- Native release contract checks that CA validation precedes systemd start.
- Shell syntax and full backend contract suite.
- Missing-CA reproduction on two clean nodes.
- Post-remediation WSS success and Panel ONLINE state.

## Rollback

Revert the commit pointed to by
`v0.1.26-fail-closed-missing-agent-ca`. Rollback would allow false-positive
native installs on hosts without a TLS trust store and is not recommended.
