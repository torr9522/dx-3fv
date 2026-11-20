# Port Resource Model

## Identity boundary

Port ownership is scoped by immutable Node identity, not public IP. GZ2 and
Frontsea IX2 share `138.252.163.137` but are different NAT targets and must
have separate Agent identities, pools, resources, leases, runtime observations,
and SSH host-key aliases.

## GZ2 entry ports

GZ2 (`138.252.163.137:22`) has the full valid listener-port space. The UI must
not display a fabricated narrow range.

- Port 22 is permanently reserved for SSH.
- System listeners and explicit administrator reservations are unavailable.
- Empty input requests automatic allocation.
- A supplied port requests that exact port and fails on conflict.
- TCP and UDP capabilities are separate resources. TCP+UDP reserves both for
  the same numeric port atomically.
- Allocation requires a durable owner and lease plus a runtime availability
  observation. A numeric range check alone is insufficient.

The implementation may define a broad policy interval such as 1-65535, but
privileged, reserved, declared, leased, and observed-in-use ports must be
excluded. This policy interval is not presented as a provider-assigned pool.

## Frontsea IX2 NAT ports

Frontsea IX2 (`138.252.163.137:38000`) has a declared provider range of
38000-38099. Port 38000 is permanently reserved for SSH. Ports 38001-38099 are
only allocatable after the exact public-to-internal NAT mapping and protocol
capability have been registered and verified. The system never modifies the
provider NAT configuration.

## Allocation invariants

- A lease key includes Node identity, internal listener port, and protocol.
- A NAT resource records public address/port separately from internal
  address/listener port.
- Manual and automatic allocation use the same transaction and lifecycle.
- Uniqueness constraints, not an in-process mutex, arbitrate concurrency.
- Delete marks a lease pending release. Reuse is allowed only after all owned
  listeners are observed absent.
- Agent or Panel restart preserves ownership and resumes reconciliation.
- Invalid, ambiguous, unverified, or exhausted resources produce explicit
  errors and are never silently substituted.
# GZ2 direct-node policy

GZ2 uses a `full` pool whose eligible range is always `1-65535`. It is not an
IX NAT pool and must never inherit Frontsea IX2's `38000-38099` declaration.
Port 22 is permanently unavailable to GOST. The UI labels this as a full-port
policy and does not present a fictitious narrow range.

Eligibility does not prove availability. Allocation additionally checks the
persistent Node/port/protocol Lease, original FLVX Forward and Tunnel ownership,
peer-share ownership, excluded resources, Runtime observations, and the actual
Agent bind result. A manual bind conflict is returned to the user; an automatic
bind conflict records an observation and tries another candidate.

TCP and UDP are separate ownership dimensions. A TCP+UDP request reserves both
for the same number in one database transaction or reserves neither. Resource
identity is based on Node ID rather than public IP, so GZ2 and Frontsea IX2 do
not collide even though both are reached through `138.252.163.137`.
