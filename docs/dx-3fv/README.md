# DX-3FV Engineering Overview

DX-3FV is developed from the frozen FLVX `3.0.1-rc16` baseline at commit
`f357b30500ef99a6cf45d18f0c8a1e1881826b57` and tree
`025c47c85613f4eba341c1e1c5fbd1a1676ca43a`.

The original FLVX Forward, Tunnel, Node, user, monitoring, limit, sharing,
configuration, Agent, and embedded GOST semantics are compatibility contracts.
The new IX multiport workflow is a separate product module. It must not create
synthetic Tunnel rows and must not change the default behavior of ordinary
Forward records.

## Development rules

- Each independent defect or feature has one ledger entry, one focused commit,
  one annotated tag, explicit tests, and rollback notes.
- Publicly reported bugs are not assumed present until the affected version,
  fixing commit, rc16 source, and a reproducible condition have been checked.
- A database success response is not runtime success. Desired state, observed
  state, listeners, NAT resources, and leases are reconciled independently.
- Test evidence and credentials stay outside Git.
- Production systems are never accessed.

## Product boundaries

1. **Forward** remains the ordinary FLVX forwarding workflow.
2. **Tunnel** remains the real FLVX Tunnel and Relay workflow, including the
   QIX single-public-port Relay use case.
3. **Multiport** manages independent GZ -> IX -> Landing businesses.
4. **IX Ports** manages declared and verified NAT resources, leases, owners,
   exclusions, and observations.

See `port-resource-model.md`, `test-plan.md`, and `security-boundaries.md`.
