package handler

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateIPv4OnlyHostWithLookup(t *testing.T) {
	lookup := func(_ context.Context, host string) ([]net.IP, error) {
		switch host {
		case "ipv4.example":
			return []net.IP{net.ParseIP("192.0.2.10")}, nil
		case "aaaa-only.example":
			return []net.IP{net.ParseIP("2001:db8::10")}, nil
		default:
			return nil, errors.New("not found")
		}
	}

	for _, host := range []string{"192.0.2.1", "ipv4.example"} {
		if err := validateIPv4OnlyHostWithLookup(context.Background(), host, lookup); err != nil {
			t.Fatalf("expected %q to be accepted: %v", host, err)
		}
	}
	for _, host := range []string{"2001:db8::1", "aaaa-only.example", "missing.example"} {
		err := validateIPv4OnlyHostWithLookup(context.Background(), host, lookup)
		if err == nil {
			t.Fatalf("expected %q to be rejected", host)
		}
		if host == "2001:db8::1" && !strings.Contains(err.Error(), "IPv6") {
			t.Fatalf("expected explicit IPv6 error, got %v", err)
		}
	}
}

func TestValidateIPv4OnlyListener(t *testing.T) {
	for _, value := range []string{"0.0.0.0", "127.0.0.1", "192.0.2.5"} {
		if err := validateIPv4OnlyListener(value); err != nil {
			t.Fatalf("expected %q to be accepted: %v", value, err)
		}
	}
	for _, value := range []string{"::", "[::]", "2001:db8::1", "example.com"} {
		if err := validateIPv4OnlyListener(value); err == nil {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
}

func TestValidateIPv4OnlyRemoteTargetsRejectsIPv6(t *testing.T) {
	if err := validateIPv4OnlyRemoteTargets(context.Background(), "192.0.2.1:443,198.51.100.2:8443"); err != nil {
		t.Fatalf("expected IPv4 targets to be accepted: %v", err)
	}
	if err := validateIPv4OnlyRemoteTargets(context.Background(), "[2001:db8::1]:443"); err == nil {
		t.Fatal("expected IPv6 target to be rejected")
	}
}

func TestSelectTunnelDialHostV4PreferenceDoesNotFallbackToIPv6(t *testing.T) {
	from := v6OnlyNode("from", "2001:db8::1")
	to := v6OnlyNode("to", "2001:db8::2")
	if _, err := selectTunnelDialHost(from, to, "v4", ""); err == nil {
		t.Fatal("IPv4-only preference must not silently fall back to IPv6")
	}
}

func TestIPv4OnlyAPIsRejectNewIPv6Targets(t *testing.T) {
	h, repo := newIXPortHandlerTest(t)

	nodeResponse := httptest.NewRecorder()
	h.nodeCreate(nodeResponse, ixPortRequest(t, http.MethodPost, "/api/v1/node/create", map[string]interface{}{
		"name": "ipv6-node", "serverIp": "2001:db8::10", "serverIpV6": "2001:db8::10",
	}, 0))
	if code := ixPortResponseCode(t, nodeResponse); code == 0 {
		t.Fatal("new IPv6 node must be rejected")
	}

	forwardResponse := httptest.NewRecorder()
	h.forwardCreate(forwardResponse, ixPortRequest(t, http.MethodPost, "/api/v1/forward/create", map[string]interface{}{
		"name": "ipv6-forward", "remoteAddr": "[2001:db8::20]:443",
	}, 0))
	if code := ixPortResponseCode(t, forwardResponse); code == 0 {
		t.Fatal("new IPv6 Forward target must be rejected")
	}

	ixResponse := httptest.NewRecorder()
	h.ixRouteCreate(ixResponse, ixPortRequest(t, http.MethodPost, "/api/v1/ix-route/create", map[string]interface{}{
		"pathProfileId": 1, "name": "ipv6-route", "landingHost": "2001:db8::30", "landingPort": 443, "protocolMode": "tcp",
	}, 0))
	if code := ixPortResponseCode(t, ixResponse); code == 0 {
		t.Fatal("new IPv6 multiport Landing must be rejected")
	}

	var count int64
	if err := repo.DB().Table("node").Where("name = ?", "ipv6-node").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("rejected IPv6 node must not be persisted")
	}
}

func TestNodeCreateDefaultsListenersToIPv4Wildcard(t *testing.T) {
	h, repo := newIXPortHandlerTest(t)
	response := httptest.NewRecorder()
	h.nodeCreate(response, ixPortRequest(t, http.MethodPost, "/api/v1/node/create", map[string]interface{}{
		"name": "ipv4-node", "serverIp": "192.0.2.10", "serverIpV4": "192.0.2.10",
	}, 0))
	if code := ixPortResponseCode(t, response); code != 0 {
		t.Fatalf("create IPv4 node failed: %s", response.Body.String())
	}
	var tcpAddr, udpAddr string
	if err := repo.DB().Table("node").Select("tcp_listen_addr, udp_listen_addr").Where("name = ?", "ipv4-node").Row().Scan(&tcpAddr, &udpAddr); err != nil {
		t.Fatal(err)
	}
	if tcpAddr != "0.0.0.0" || udpAddr != "0.0.0.0" {
		t.Fatalf("expected IPv4 wildcard listeners, got tcp=%q udp=%q", tcpAddr, udpAddr)
	}
}
