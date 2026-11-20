# DX3FV-0026: Tunnel protocol capability audit

- Priority: P0 behavior and compatibility
- Baseline: `v0.1.29-fix-theme-switch-background-flash`
- Target release: `v0.1.30-tunnel-transport-default-tcp`
- Scope: FLVX 3.0.1-rc16 derived Tunnel control plane and embedded go-gost/x

## Problem

The Tunnel form called `tls`, `tcp`, `kcp`, and multiplexed variants a
"protocol". That label mixed two different concepts: the transport between
Tunnel nodes and the logical TCP or UDP business request carried by GOST
Relay. New hops also silently defaulted to TLS. The requested default is plain
TCP, but existing Tunnel transports must not be rewritten.

## Capability matrix

| Tunnel model | Runtime shape | Inter-node transports | Logical TCP | Logical UDP |
| --- | --- | --- | --- | --- |
| Type 1, port forwarding | Entry selection used by normal Forward services; no Type-2 Relay service is built | Not applicable to a Relay hop | Provided by the Forward service protocol | Provided by the Forward service protocol when that Forward mode supports UDP |
| Type 2, tunnel forwarding | Entry chain -> optional middle Relay service(s) -> exit Relay service | `tcp`, `tls`, `wss`, `mtls`, `mwss`, `mtcp`, `kcp` | Yes | Yes, through Relay's UDP logical network over the selected transport |

There is no user-selectable Tunnel Type 3. `chain_type=3` identifies an exit
node inside a Type-2 Tunnel. `chain_type=1` is entry and `chain_type=2` is a
middle hop.

Entry nodes own a chain/dialer for the next hop. Middle nodes own both a Relay
listener and a chain to the next hop. Exit nodes own a Relay listener and dial
the configured Landing through the Relay handler. The selected transport must
match on both sides of each hop.

## Source evidence and call chain

1. `go-backend/internal/store/model/model.go` stores `Tunnel.Type`, the summary
   `Tunnel.Protocol`, and a per-row `ChainTunnel.Protocol`.
2. `go-backend/internal/http/handler/mutations.go` parses the entry, middle,
   and exit rows in `prepareTunnelCreateState`.
3. `buildTunnelChainConfig` creates a GOST `relay` connector and passes the
   stored value to `buildTunnelDialerConfig`.
4. `buildTunnelChainServiceConfig` creates a GOST `relay` handler and passes
   the same value to `buildTunnelListenerConfig`.
5. The embedded fork registers all seven values in `go-gost/x/dialer/*` and
   `go-gost/x/listener/*`. TLS listens and dials over TCP; KCP uses its UDP
   packet transport; the multiplexed variants remain stream transports.
6. `go-gost/x/connector/relay/connector.go` sets `relay.FUDP` and the UDP
   network feature for logical UDP requests. The matching Relay handler
   dispatches that request as UDP. This is independent of whether the outer
   transport is TCP, TLS, WSS, multiplexed, or KCP.
7. `go-gost/x/chain/required_relay_transport_test.go` exercises logical TCP
   and UDP over `tcp`, `tls`, and `kcp`, proving that TCP+UDP does not require
   two Tunnel transport services.

## TLS meaning

TLS is an encrypted outer transport between two Tunnel nodes. It is not the
business protocol, is not a substitute name for TCP, and does not mean that a
Landing speaks TLS. MTLS is the multiplexed TLS transport. WSS/MWSS similarly
wrap WebSocket transport with TLS. A Guangzhou or other entry node only uses
TLS when the chosen next-hop transport is TLS; ordinary TCP forwarding does
not require it.

## Decision

- Rename the form label from "协议" to "隧道传输".
- Default newly created Type-2 hop and exit transports to `tcp`.
- Keep the seven transports that are actually registered by the embedded
  Runtime. Do not add `udp` or `tcp+udp` to this selector because those are
  logical business networks, not outer transports.
- Validate the transport server-side and reject unknown or business-protocol
  labels instead of passing arbitrary strings to GOST.
- Preserve every stored protocol during edit. The legacy fallback remains TLS
  only when an old row is missing a value; no migration or bulk rewrite runs.

## Tests and evidence

- Backend tests cover the TCP default, every registered transport, rejection
  of `udp`/`tcp+udp`, and matching listener/dialer generation.
- The embedded Relay transport matrix tests remain the Runtime proof for TCP
  and UDP logical traffic.
- Frontend production build verifies the typed form after the label/default
  change.
- Browser acceptance must verify that a new Type-2 Tunnel defaults to TCP and
  that editing the existing shared Relay preserves its stored transport.

## Compatibility and rollback

There is no schema migration. Existing Tunnel and ChainTunnel rows are not
changed. Roll back by switching the native release symlink to the preceding
`v0.1.29` release and restarting the Panel. Tunnels created with plain TCP
remain readable by v0.1.29 because TCP was already a supported transport.
