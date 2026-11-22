# DX-3FV Issue Ledger

Each issue file records:

- stable issue ID and priority;
- source evidence and affected versions;
- reproduction and root cause;
- scoped fix and changed files;
- automated and real-operation tests;
- focused commit and annotated tag;
- deployment impact and rollback procedure.

Allowed upstream-audit states are `FIXED_UPSTREAM`, `PRESENT_IN_RC16`,
`FIXED_IN_DX3FV`, `REPRODUCED`, `NOT_REPRODUCED`, `UNKNOWN`, `NOT_RELEVANT`, and
`DEFERRED`.

Unrelated problems are never combined in one fix. A blocked issue does not
block independent work.

Latest release-preparation issues:

- `DX3FV-0045-native-release-portability.md`
- `DX3FV-0046-release-candidate-publication-assets.md`
- `DX3FV-0047-relay-cleanup-contract-race.md`
- `DX3FV-0048-release-candidate-audit-disclosure.md`
- `DX3FV-0049-portable-go-module-sbom.md`
- `DX3FV-0050-go-sbom-json-stream.md`
- `DX3FV-0051-publication-knowledge-base.md`
- `DX3FV-0052-sanitized-publication-history.md`
- `DX3FV-0053-publication-secret-scan-baseline.md`
- `DX3FV-0054-publication-history-example-allowlist.md`
- `DX3FV-0055-debian-one-command-installer.md`
- `DX3FV-0056-publication-generated-grpc-source.md`
- `DX3FV-0057-multiport-create-scope.md`
- `DX3FV-0058-udp-session-expiry-race.md`
- `DX3FV-0059-react-router-7.md`
- `DX3FV-0060-hour-scale-traffic.md`
- `DX3FV-0061-high-concurrency.md`
- `DX3FV-0062-fault-recovery-matrix.md`
- `DX3FV-0063-rc4-source-candidate.md`
- `DX3FV-0064-publish-rc4-installer.md`
- `DX3FV-0065-stable-v1.0.0-release.md`
- `DX3FV-0066-publish-v1.0.0-installer.md`
- `DX3FV-0067-entry-address-copy.md`
- `DX3FV-0068-update-policy-and-source.md`
- `DX3FV-0069-semantic-update-order.md`
- `DX3FV-0070-stable-v1.0.1-release.md`
- `DX3FV-0071-publish-v1.0.1-installer.md`
- `DX3FV-0072-multiport-landing-copy.md`
- `DX3FV-0073-entry-auto-port-policy.md`
- `DX3FV-0074-unified-entry-auto-port-policy.md`
