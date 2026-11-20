# Security and Environment Boundaries

## Permanent prohibition

Production Panel, production domains, production nodes, logs, APIs, databases,
listeners, and network ports are outside the project scope and are never
accessed, including read-only access.

## Authorized systems

- Development and build work: `152.53.192.226` only.
- Dedicated test Panel: `185.126.158.33`.
- Test nodes only after exact endpoint and host identity verification: the
  Guangzhou jump host, GZ2, Frontsea IX2, QIX, and Blindbox IX.

No broad port scanning is permitted. Every IX management connection must use
the independently identified Guangzhou jump host `134.175.156.145:22` and a
target-specific host-key alias. The jump host and GZ2 are distinct identities.

## Diagnostic traffic boundary

The prohibition concerns operator and development-machine access to an IX
egress public address. The development machine must never ping, scan, or use an
IX egress address as an SSH fallback. IX management always follows the approved
Guangzhou jump path.

Authenticated Panel diagnostics are normal product traffic. They may ask the
GZ Agent to test the persisted IX business endpoint and the matching IX Agent
to test the exact Landing host and port stored on that business route. Tunnel,
Forward, and Service Monitor diagnostics use their persisted topology or
configured target. These paths must not expand one target into a port scan, and
an unverified UDP response must not be reported as success.

## Secrets

Credentials are loaded at runtime from the approved `/1x-ui/` files. They are
never copied into this repository, command output, reports, screenshots, build
artifacts, or release archives.
