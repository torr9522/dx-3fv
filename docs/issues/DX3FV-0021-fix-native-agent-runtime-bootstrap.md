# DX3FV-0021: Native Agent release omits Runtime bootstrap config

- Priority: P1 Agent availability
- Status: FIXED_IN_DX3FV
- Baseline: `v0.1.24-enforce-probe-egress-policy`
- Release tag: `v0.1.25-fix-native-agent-runtime-bootstrap`
- Schema change: none

## Problem and evidence

The v0.1.24 native Agent package installed `config.json` but did not install
the GOST Runtime configuration file. On clean GZ2, Frontsea IX2, and Blindbox
IX test nodes, systemd repeatedly restarted the Agent with:

```text
Config File "gost" Not Found
```

All three nodes reproduced the same failure with the same verified Agent
binary. No Runtime listener or business traffic started.

FLVX rc16's official `install.sh` creates `gost.json` containing an empty
JSON object on a fresh installation. The embedded parser requires this file
before it starts the WebSocket reporter.

## Root cause

`scripts/build-native-release.sh` did not package an empty Runtime config and
`release/native/install-agent-native.sh` did not install one. The install
script's immediate `systemctl is-active` check also accepted the transient
`activating` state before the process failed, so remote verification must
wait for a stable active process.

## Fix

- Add `release/native/agent-gost.json` with the rc16-compatible empty object.
- Include that asset in every native release.
- Install it as `gost.json` with mode `0600` before systemd starts Agent.
- Add a release contract test that validates JSON, packaging, permissions, and
  start ordering.

The change does not replace go-gost/x, alter Runtime behavior, change the Agent
protocol, or add services. It only supplies the required empty desired state
for a fresh native installation.

## Verification

- `TestNativeAgentReleaseBootstrapsEmptyRuntimeConfig`.
- Full Agent tests and backend contract tests.
- Reproducible tagged native build.
- Clean-node deployment to GZ2, Frontsea IX2, and Blindbox IX.
- Stable systemd active state, Panel ONLINE state, and zero initial Runtime
  services.

## Rollback

Roll back the Agent release directory and systemd unit to the preceding tagged
release. Do not remove an existing `gost.json` during rollback because it is
Runtime state. The v0.1.24 native installer cannot bootstrap a clean node and
must not be used for a fresh deployment.
