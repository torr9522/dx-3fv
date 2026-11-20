package handler

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"go-backend/internal/store/model"
)

func TestChineseNodeRenamePreservesStableIdentity(t *testing.T) {
	h, repository := newIXPortHandlerTest(t)
	now := time.Now().UnixMilli()
	node := &model.Node{
		Name: "gz2", Secret: "stable-agent-secret", ServerIP: "192.0.2.10",
		ServerIPV4: sql.NullString{String: "192.0.2.10", Valid: true}, Port: "10000-20000",
		CreatedTime: now, Status: 0, TCPListenAddr: "0.0.0.0", UDPListenAddr: "0.0.0.0",
		ForwardMode: "agent",
	}
	if err := repository.DB().Create(node).Error; err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	h.nodeUpdate(response, ixPortRequest(t, http.MethodPost, "/api/v1/node/update", map[string]interface{}{
		"id": node.ID, "name": "广州2号", "serverIp": "192.0.2.10", "serverIpV4": "192.0.2.10",
		"port": "10000-20000", "forwardMode": "agent", "tcpListenAddr": "0.0.0.0", "udpListenAddr": "0.0.0.0",
	}, 0))
	if code := ixPortResponseCode(t, response); code != 0 {
		t.Fatalf("rename failed: %s", response.Body.String())
	}

	var stored model.Node
	if err := repository.DB().First(&stored, node.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.ID != node.ID || stored.Secret != "stable-agent-secret" || stored.Name != "广州2号" {
		t.Fatalf("rename changed stable identity: %+v", stored)
	}
	if got := nodeDisplayName(&nodeRecord{ID: stored.ID, Name: stored.Name}); got != "广州2号" {
		t.Fatalf("diagnostic display name = %q", got)
	}
}

func TestChineseResourceNamesRoundTripWithoutEnteringRuntimeIDs(t *testing.T) {
	_, repository := newIXPortHandlerTest(t)
	now := time.Now().UnixMilli()
	node := model.Node{Name: "前海IX2", Secret: "node-secret", ServerIP: "192.0.2.20", Port: "38001-38099", CreatedTime: now, Status: 1, TCPListenAddr: "0.0.0.0", UDPListenAddr: "0.0.0.0", ForwardMode: "agent"}
	if err := repository.DB().Create(&node).Error; err != nil {
		t.Fatal(err)
	}
	tunnel := model.Tunnel{Name: "广州2号 → QIX 共享 Relay", Type: 2, Protocol: "tcp", Flow: 1, TrafficRatio: 1, CreatedTime: now, UpdatedTime: now, Status: 1, Inx: 1, IPPreference: "v4"}
	if err := repository.DB().Create(&tunnel).Error; err != nil {
		t.Fatal(err)
	}
	forward := model.Forward{UserID: 1, UserName: "admin", Name: "美国 TRI", TunnelID: tunnel.ID, Protocol: "tcp", RemoteAddr: "192.0.2.30:443", Strategy: "fifo", CreatedTime: now, UpdatedTime: now, Status: 1, Inx: 1}
	if err := repository.DB().Create(&forward).Error; err != nil {
		t.Fatal(err)
	}
	profile := model.IXPathProfile{Name: "广州2号 → 前海IX2", EntryNodeID: node.ID, IXNodeID: node.ID + 1, EntryPoolID: 1, IXPoolID: 2, Status: 1, CreatedTime: now, UpdatedTime: now}
	if err := repository.DB().Create(&profile).Error; err != nil {
		t.Fatal(err)
	}
	route := model.IXRoute{PathProfileID: profile.ID, Name: "香港落地业务", LandingHost: "192.0.2.40", LandingPort: 443, ProtocolMode: "tcp", DesiredState: "active", ObservedState: "active", CreatedTime: now, UpdatedTime: now}
	if err := repository.DB().Create(&route).Error; err != nil {
		t.Fatal(err)
	}

	if got := buildForwardServiceBase(forward.ID, forward.UserID, 0); got != "1_1_0" {
		t.Fatalf("Forward Runtime identity unexpectedly contains display name: %q", got)
	}
	if routeRuntime := "dx3fv_ix_" + strconv.FormatInt(route.ID, 10) + "_entry_tcp"; routeRuntime != "dx3fv_ix_1_entry_tcp" {
		t.Fatalf("IX Runtime identity unexpectedly contains display name: %q", routeRuntime)
	}

	var storedNode model.Node
	var storedTunnel model.Tunnel
	var storedForward model.Forward
	var storedProfile model.IXPathProfile
	var storedRoute model.IXRoute
	for _, query := range []struct {
		value interface{}
		id    int64
	}{{&storedNode, node.ID}, {&storedTunnel, tunnel.ID}, {&storedForward, forward.ID}, {&storedProfile, profile.ID}, {&storedRoute, route.ID}} {
		if err := repository.DB().First(query.value, query.id).Error; err != nil {
			t.Fatal(err)
		}
	}
	if storedNode.Name != "前海IX2" || storedTunnel.Name != "广州2号 → QIX 共享 Relay" || storedForward.Name != "美国 TRI" || storedProfile.Name != "广州2号 → 前海IX2" || storedRoute.Name != "香港落地业务" {
		t.Fatalf("UTF-8 names did not round-trip: node=%q tunnel=%q forward=%q profile=%q route=%q", storedNode.Name, storedTunnel.Name, storedForward.Name, storedProfile.Name, storedRoute.Name)
	}
}
