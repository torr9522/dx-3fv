# Native release and rollback

DX-3FV does not use Docker. Build a release from a clean tagged commit:

```bash
./scripts/build-native-release.sh v0.1.14-feature-native-release /root/dx-3fv-releases
```

The build refuses dirty worktrees, lightweight tags, or a tag that does not
resolve to the current `HEAD`. The recorded Commit and Tree therefore identify
all packaged source and release logic.

The archive contains statically linked Panel and Agent binaries, frontend
assets, systemd units, Caddy configuration, environment examples, dependency
manifests, per-file SHA256 values, and native install/upgrade/rollback scripts.
No credential is generated until installation on the target machine.

Panel releases live under `/opt/dx-3fv/releases`; `/opt/dx-3fv/current` is the
only systemd execution path. Upgrade stops the Panel, makes a consistent SQLite
backup, switches the symlink, and starts the new version. A failed local health
check restores both the previous symlink and the pre-upgrade database. Manual
rollback accepts the matching database backup as its second argument when the
target schema is not backward compatible.

The Agent uses `/opt/dx-3fv-agent/current` and a node-local `config.json` with
mode 0600. Agent deployment must be performed only after Node identity, Secret,
Panel address, and existing Runtime ownership are verified.

Both Panel and Agent installers verify every packaged file against
`SHA256SUMS`. The Agent installer requires an already prepared node-local
configuration file; it never generates or discovers a Node Secret.
