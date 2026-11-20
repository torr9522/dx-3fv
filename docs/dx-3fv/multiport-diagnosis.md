# Multiport diagnosis contract

Each result is evidence graded:

- `PASS`: the stated method succeeded.
- `FAIL`: the stated method ran and failed, timed out, or its Agent was offline.
- `UNVERIFIED`: the available method cannot prove the requested property.

Default TCP segment checks use TCP Connect. They do not prove application
payload. Default UDP results are `UNVERIFIED` because the legacy `UdpPing`
implementation only creates a connected UDP socket.

The controlled Echo option is valid only when the configured Landing is an
authorized echo fixture. The Agent sends a random nonce and requires the exact
nonce to return. This proves payload round trip for that probe, not arbitrary
application protocol compatibility.

The displayed GZ-to-IX UDP nonce check enters the IX route listener and reaches
Landing before the echo returns. Its method is explicitly
`udp_nonce_echo_via_ix_route`. It must not be added to the IX-to-Landing value
or described as a pure first-hop RTT. All leg measurements retain independent
timestamps and methods.
