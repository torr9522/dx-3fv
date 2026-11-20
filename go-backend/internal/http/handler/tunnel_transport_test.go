package handler

import "testing"

func TestNormalizeTunnelTransportDefaultsToTCP(t *testing.T) {
	got, err := normalizeTunnelTransport("")
	if err != nil {
		t.Fatalf("normalize empty transport: %v", err)
	}
	if got != "tcp" {
		t.Fatalf("expected tcp, got %q", got)
	}
}

func TestNormalizeTunnelTransportAcceptsRuntimeRegistryValues(t *testing.T) {
	for _, value := range []string{"tcp", "tls", "wss", "mtls", "mwss", "mtcp", "kcp"} {
		t.Run(value, func(t *testing.T) {
			got, err := normalizeTunnelTransport("  " + value + "  ")
			if err != nil {
				t.Fatalf("normalize %q: %v", value, err)
			}
			if got != value {
				t.Fatalf("expected %q, got %q", value, got)
			}
		})
	}
}

func TestNormalizeTunnelTransportRejectsBusinessProtocolLabels(t *testing.T) {
	for _, value := range []string{"udp", "tcp+udp", "http3", "unknown"} {
		t.Run(value, func(t *testing.T) {
			if _, err := normalizeTunnelTransport(value); err == nil {
				t.Fatalf("expected %q to be rejected", value)
			}
		})
	}
}

func TestTunnelTransportBuildsMatchingDialerAndListener(t *testing.T) {
	for _, value := range []string{"tcp", "tls", "wss", "mtls", "mwss", "mtcp", "kcp"} {
		dialer := buildTunnelDialerConfig(value)
		listener := buildTunnelListenerConfig(value)
		if dialer["type"] != value {
			t.Fatalf("dialer type for %q: %#v", value, dialer["type"])
		}
		if listener["type"] != value {
			t.Fatalf("listener type for %q: %#v", value, listener["type"])
		}
	}
}
