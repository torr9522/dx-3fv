# DX3FV-0001: Limiter rollback contract is intermittently failing

- Priority: P1 investigation
- Status: REPRODUCED
- Baseline: FLVX 3.0.1-rc16

## Evidence

The complete backend suite produced one failure in
`TestForwardCreateRollbackWhenLimiterDispatchFailsContract`. The exact test
then passed ten isolated repetitions. The sanitized reproduction log is stored
outside Git under the Phase 2 evidence directory.

## Current assessment

The failure is reproducibly flaky at suite scope, but no root cause or product
impact has yet been proven. Likely shared process state or cross-test pollution
must be confirmed from source and deterministic reproduction before a fix is
authorized.

## Required completion

Identify the shared state, create a deterministic regression test, implement a
focused correction, run the complete SQLite suite and applicable PostgreSQL
contract tests, then record the commit, annotated tag, and rollback command.
