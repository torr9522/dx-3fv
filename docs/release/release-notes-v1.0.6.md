# DX-3FV v1.0.6

DX-3FV v1.0.6 is a stable maintenance release.

## Fixed

- Serialize updates to the same Forward so a failed older Agent apply cannot
  roll back a newer successful target change.
- Preserve independent concurrency between different Forward IDs and shared
  Tunnel sibling businesses.

## Compatibility

- Database schema remains 12.
- Panel, Agent, GOST configuration, API payloads, and business data are
  unchanged.
- Existing v1.0.5 native installations can use the normal native upgrade and
  rollback workflow.

See `docs/issues/DX3FV-0077-forward-update-concurrency.md` for the deterministic
reproduction and acceptance scope.
