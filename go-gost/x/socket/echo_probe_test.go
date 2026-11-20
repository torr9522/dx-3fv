package socket

import (
	"io"
	"net"
	"testing"
	"time"
)

func startEchoListener(t *testing.T, protocol string) (string, func()) {
	t.Helper()
	if protocol == "tcp" {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			defer conn.Close()
			_, _ = io.Copy(conn, conn)
		}()
		return listener.Addr().String(), func() { _ = listener.Close() }
	}
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		buffer := make([]byte, 512)
		n, peer, err := conn.ReadFrom(buffer)
		if err == nil {
			_, _ = conn.WriteTo(buffer[:n], peer)
		}
	}()
	return conn.LocalAddr().String(), func() { _ = conn.Close() }
}

func TestEchoProbeRequiresExactPayloadRoundTrip(t *testing.T) {
	reporter := &WebSocketReporter{}
	for _, protocol := range []string{"tcp", "udp"} {
		t.Run(protocol, func(t *testing.T) {
			address, stop := startEchoListener(t, protocol)
			defer stop()
			host, portText, err := net.SplitHostPort(address)
			if err != nil {
				t.Fatal(err)
			}
			port, err := net.LookupPort(protocol, portText)
			if err != nil {
				t.Fatal(err)
			}
			result, err := reporter.handleEchoProbe(map[string]interface{}{
				"protocol": protocol,
				"ip":       host,
				"port":     port,
				"payload":  "dx3fv-echo-nonce-0123456789",
				"timeout":  int((2 * time.Second) / time.Millisecond),
			})
			if err != nil {
				t.Fatal(err)
			}
			if !result.Success || result.Bytes != len("dx3fv-echo-nonce-0123456789") {
				t.Fatalf("echo proof failed: %+v", result)
			}
		})
	}
}

func TestEchoProbeDoesNotTreatUDPDialAsPayloadSuccess(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	host, portText, _ := net.SplitHostPort(conn.LocalAddr().String())
	port, _ := net.LookupPort("udp", portText)
	result, err := (&WebSocketReporter{}).handleEchoProbe(map[string]interface{}{
		"protocol": "udp", "ip": host, "port": port,
		"payload": "dx3fv-no-echo-0123456789", "timeout": 50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success || result.ErrorMessage == "" {
		t.Fatalf("UDP connect was misreported as payload success: %+v", result)
	}
}
