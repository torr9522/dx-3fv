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
